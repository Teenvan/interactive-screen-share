package http

import (
	"context"
	"encoding/json"
	"image/jpeg"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"m1k1o/neko/internal/config"
	"m1k1o/neko/internal/types"
	"m1k1o/neko/internal/zoom"
)

type Server struct {
	logger zerolog.Logger
	router *chi.Mux
	http   *http.Server
	conf   *config.Server
}

const contextHeader = "x-zoom-app-context"

func New(conf *config.Server, webSocketHandler types.WebSocketHandler, desktop types.DesktopManager) *Server {
	logger := log.With().Str("module", "http").Logger()

	router := chi.NewRouter()
	router.Use(middleware.RequestID) // Create a request ID for each request
	router.Use(middleware.RequestLogger(&logformatter{logger}))
	router.Use(middleware.Recoverer) // Recover from panics without crashing server
	
	// Set middleware response header for Zoom app
	router.Use(middleware.SetHeader("Strict-Transport-Security", "max-age=31536000"))
	router.Use(middleware.SetHeader("X-Content-Type-Options", "nosniff"))
	router.Use(middleware.SetHeader("Content-Security-Policy", `default-src 'self';
	style-src 'report-sample' 'self' 'unsafe-inline';
	script-src * 'self' https://appssdk.zoom.us 'unsafe-inline';
	object-src 'none';
	base-uri 'self';
	connect-src * 'self' ws://3c97-137-220-76-2.eu.ngrok.io;
	font-src * 'self';
	frame-src 'self';
	img-src 'self';
	manifest-src 'self';
	media-src 'self';
	worker-src 'none';`))

	router.Use(middleware.SetHeader("Referrer-Policy", "same-origin"))
	router.Use(middleware.SetHeader("X-Frame-Option", "same-origin"))

	if conf.PathPrefix != "/" {
		router.Use(func(h http.Handler) http.Handler {
			return http.StripPrefix(conf.PathPrefix, h)
		})
	}

	router.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		err := webSocketHandler.Upgrade(w, r)
		if err != nil {
			logger.Warn().Err(err).Msg("failed to upgrade websocket conection")
		}
	})

	router.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
		password := r.URL.Query().Get("pwd")
		isAdmin, err := webSocketHandler.IsAdmin(password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		if !isAdmin {
			http.Error(w, "bad authorization", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		stats := webSocketHandler.Stats()
		if err := json.NewEncoder(w).Encode(stats); err != nil {
			logger.Warn().Err(err).Msg("failed writing json error response")
		}
	})

	router.Get("/screenshot.jpg", func(w http.ResponseWriter, r *http.Request) {
		password := r.URL.Query().Get("pwd")
		isAdmin, err := webSocketHandler.IsAdmin(password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		if !isAdmin {
			http.Error(w, "bad authorization", http.StatusUnauthorized)
			return
		}

		if webSocketHandler.IsLocked("login") {
			http.Error(w, "room is locked", http.StatusLocked)
			return
		}

		quality, err := strconv.Atoi(r.URL.Query().Get("quality"))
		if err != nil {
			quality = 90
		}

		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Content-Type", "image/jpeg")

		img := desktop.GetScreenshotImage()
		if err := jpeg.Encode(w, img, &jpeg.Options{Quality: quality}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("true"))
	})

	router.Get("/auth", func (w http.ResponseWriter, r *http.Request)  {
		code := r.URL.Query().Get("code")
		logger.Info().Msgf("Code : %s", code)
		
		zoomClient := zoom.NewClient()

		if zoomClient != nil {
			// Get access token from zoom
			accessToken, err := zoomClient.GetToken(code)

			if err != nil {
				logger.Err(err).Msg("Retrieving token failed")
			}

			deeplink, err := zoomClient.GetDeepLink(accessToken)
			if err != nil {
				logger.Err(err).Msg("Retrieving deep link failed")
			}

			http.Redirect(w, r, deeplink, http.StatusSeeOther)
		}
	})

	fs := http.FileServer(http.Dir(conf.Static))
	router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat(conf.Static + r.URL.Path); !os.IsNotExist(err) {
			fs.ServeHTTP(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	server := &http.Server{
		Addr:    conf.Bind,
		Handler: router,
	}

	return &Server{
		logger: logger,
		router: router,
		http:   server,
		conf:   conf,
	}
}

func (s *Server) Start() {
	if s.conf.Cert != "" && s.conf.Key != "" {
		go func() {
			if err := s.http.ListenAndServeTLS(s.conf.Cert, s.conf.Key); err != http.ErrServerClosed {
				s.logger.Panic().Err(err).Msg("unable to start https server")
			}
		}()
		s.logger.Info().Msgf("https listening on %s", s.http.Addr)
	} else {
		go func() {
			if err := s.http.ListenAndServe(); err != http.ErrServerClosed {
				s.logger.Panic().Err(err).Msg("unable to start http server")
			}
		}()
		s.logger.Warn().Msgf("http listening on %s", s.http.Addr)
	}
}

func (s *Server) Shutdown() error {
	return s.http.Shutdown(context.Background())
}
