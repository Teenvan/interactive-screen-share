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
	
	var zm_host string = "https://zoom.us"
	var zm_client_id string = "3HFvhLMeRZyz7K6Cjr62Q"
	var zm_client_secret string = "Y6edLK2mnAbDC78bOgoxgQt7DdQay03i"
	var zm_redirect_url = "https://rides-centres-profit-disclaimer.trycloudflare.com/auth"

	if len(config["ZM_HOST"]) != 0 {
		zm_host = config["ZM_HOST"]
	}

	if len(config["ZM_CLIENT_ID"]) != 0 {
		zm_client_id = config["ZM_CLIENT_ID"]
	}

	if len(config["ZM_CLIENT_SECRET"]) != 0 {
		zm_client_secret = config["ZM_CLIENT_SECRET"]	
	}

	if len(config["ZM_REDIRECT_URL"]) != 0 {
		zm_redirect_url = config["ZM_REDIRECT_URL"]
	}

	zoomApp.Host = zm_host
	zoomApp.ClientID = zm_client_id
	zoomApp.ClientSecret = zm_client_secret
	zoomApp.RedirectURL = zm_redirect_url

	return zoomApp, nil
}
