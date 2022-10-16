package zoom

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"io"
)


type ZoomApp struct{
	Host string
	ClientID string
	ClientSecret string
	RedirectURL string
	SessionSecret string
}

type Config map[string]string


func readConfig(filename string) (Config, error) {

	config := Config{}

	if len(filename) == 0 {
		return nil,  fmt.Errorf("len of filename is invalid")
	}

	file, err := os.Open(filename)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadString('\n')
		if equal := strings.Index(line, "="); equal >= 0 {
			if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
				value := ""
				if len(line) > equal {
					value = strings.TrimSpace(line[equal+1:])
				}
				// assign the config map
				config[key] = value
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}

func GetZoomApp() (ZoomApp, error) {
	filename := "../.env"
	// Read config file
	config, err := readConfig(filename)
	
	if err != nil {
		return ZoomApp{}, err
	}

	zoomApp := ZoomApp{}
	zoomApp.Host = config["ZM_HOST"]
	zoomApp.ClientID = config["ZM_CLIENT_ID"]
	zoomApp.ClientSecret = config["ZM_CLIENT_SECRET"]
	zoomApp.RedirectURL = config["ZM_REDIRECT_URL"]
	zoomApp.SessionSecret = config["SESSION_SECRET"]

	return zoomApp, nil
}
