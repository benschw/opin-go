package vault

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/benschw/opin-go/rest"
	"github.com/hashicorp/vault/api"
)

type AuthInfo struct {
	ClientToken   string `json:"client_token"`
	LeaseDuration int    `json:"lease_duration"`
	Renewable     bool   `json:"renewable"`
}
type LoginResponse struct {
	Auth          AuthInfo `json:"auth"`
	LeaseDuration int      `json:"lease_duration"`
	LeaseId       string   `json:""`
	Renewable     bool     `json:""`
}

type AppIdLoginConfig struct {
	AppId  string `json:"app_id"`
	UserId string `json:"user_id"`
}

func DefaultAppIdLoginConfig() (*AppIdLoginConfig, error) {
	appId := os.Getenv("VAULT_APP_ID")
	if appId == "" {
		return nil, fmt.Errorf("VAULT_APP_ID environment variable not set")
	}

	userId := os.Getenv("VAULT_USER_ID")
	if userId == "" {
		return nil, fmt.Errorf("VAULT_USER_ID environment variable not set")
	}

	return &AppIdLoginConfig{AppId: appId, UserId: userId}, nil
}

func ApiIdLogin(config *api.Config, req *AppIdLoginConfig) (*LoginResponse, error) {
	var resp LoginResponse

	r, err := rest.MakeRequest("POST", fmt.Sprintf("%s/v1/auth/app-id/login", config.Address), req)
	if err != nil {
		return nil, fmt.Errorf("Problem logging in: %s", err)
	}
	err = rest.ProcessResponseEntity(r, &resp, http.StatusOK)
	if err != nil {
		respBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		log.Print(string(respBody[:]))
	}
	return &resp, err
}

func NewAppIdClient(config *api.Config, loginConfig *AppIdLoginConfig) (*api.Client, error) {
	client, err := api.NewClient(config)
	if err != nil {
		log.Printf("Problem creating default client %s", err)
		return client, err
	}

	login, err := ApiIdLogin(config, loginConfig)
	if err != nil {
		return client, err
	}

	client.SetToken(login.Auth.ClientToken)
	return client, nil
}

func DefaultAppIdClient() (*api.Client, error) {
	config, err := DefaultConsulConfig()
	if err != nil {
		log.Printf("couldn't discover vault server address: %s", err)
		return nil, err
	}

	loginConfig, err := DefaultAppIdLoginConfig()
	if err != nil {
		log.Printf("couldn't find app_id or user_id: %s", err)
		return nil, err
	}
	return NewAppIdClient(config, loginConfig)
}