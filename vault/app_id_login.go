package vault

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/benschw/dns-clb-go/clb"
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

type LoginRequest struct {
	AppId  string `json:"app_id"`
	UserId string `json:"user_id"`
}

func ApiIdLogin(config *api.Config, appId string, userId string) (*LoginResponse, error) {
	req := &LoginRequest{AppId: appId, UserId: userId}
	var resp LoginResponse

	log.Printf("%+v", req)
	r, err := rest.MakeRequest("POST", fmt.Sprintf("%s/v1/auth/app-id/login", config.Address), req)
	if err != nil {
		log.Printf("Problem logging in: %s", err)
		return nil, err
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

func NewAppIdClient(config *api.Config, appId string, userId string) (*api.Client, error) {
	client, err := api.NewClient(config)
	if err != nil {
		log.Printf("Problem creating default client %s", err)
		return client, err
	}

	login, err := ApiIdLogin(config, appId, userId)
	if err != nil {
		return client, err
	}

	client.SetToken(login.Auth.ClientToken)
	return client, nil
}

func DefaultAppIdClient() (*api.Client, error) {
	lb := clb.NewDefaultClb(clb.RoundRobin)
	vaultAddr, err := lb.GetAddress("vault.service.consul")
	if err != nil {
		log.Printf("couldn't discover vault server address: %s", err)
		return nil, err
	}

	config := api.DefaultConfig()
	config.Address = fmt.Sprintf("http://%s:%d", vaultAddr.Address, vaultAddr.Port)

	appId := os.Getenv("VAULT_APP_ID")
	if appId == "" {
		return nil, fmt.Errorf("VAULT_APP_ID environment variable not set")
	}

	userId := os.Getenv("VAULT_USER_ID")
	if userId == "" {
		return nil, fmt.Errorf("VAULT_USER_ID environment variable not set")
	}

	return NewAppIdClient(config, appId, userId)
}
