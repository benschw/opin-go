package vault

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/benschw/opin-go/rest"
	"github.com/hashicorp/vault/api"
)

type Errors struct {
	Errors []string `json:"errors"`
}
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
	addr := fmt.Sprintf("%s/v1/auth/app-id/login", config.Address)

	return loginRequest(addr, req)
}
func loginRequest(addr string, req *AppIdLoginConfig) (*LoginResponse, error) {
	var resp LoginResponse
	log.Printf("Logging in at %s", addr)

	r, err := rest.MakeRequest("POST", addr, req)
	if err != nil {
		return nil, fmt.Errorf("Problem logging in: %s", err)
	}
	err = rest.ProcessResponseEntity(r, &resp, http.StatusOK)
	if err != nil {
		if r.StatusCode == http.StatusTemporaryRedirect {
			loc, err := r.Location()
			if err != nil {
				return nil, err
			}
			log.Print("Following Temporary Redirect")
			return loginRequest(loc.String(), req)
		} else {

			var errors Errors
			if err2 := rest.ForceProcessResponseEntity(r, errors); err2 == nil {
				for e := range errors.Errors {
					log.Print(e)
				}
			}
		}
	}
	return &resp, err
}

func NewAppIdClient(config *api.Config, loginConfig *AppIdLoginConfig) (*api.Client, error) {
	client, err := api.NewClient(config)
	if err != nil {
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
