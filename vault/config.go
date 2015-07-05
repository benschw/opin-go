package vault

import (
	"fmt"

	"github.com/benschw/dns-clb-go/clb"
	"github.com/hashicorp/vault/api"
)

func DefaultConsulConfig() (*api.Config, error) {
	lb := clb.NewDefaultClb(clb.RoundRobin)
	vaultAddr, err := lb.GetAddress("vault.service.consul")
	if err != nil {
		return nil, fmt.Errorf("couldn't discover vault server address: %s", err)
	}

	config := api.DefaultConfig()
	config.Address = fmt.Sprintf("http://%s:%d", vaultAddr.Address, vaultAddr.Port)

	return config, nil
}
