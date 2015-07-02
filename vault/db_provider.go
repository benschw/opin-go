package vault

import (
	"fmt"
	"log"

	"github.com/benschw/dns-clb-go/clb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hashicorp/vault/api"
	"github.com/jinzhu/gorm"
)

type DbProvider interface {
	Get() (*gorm.DB, error)
}

// Static DbProvider builds a gorm.DB connection
// - using a static connection string
func NewStaticDbProvider(dbStr string) (*StaticDbProvider, error) {
	d, err := gorm.Open("mysql", dbStr)
	if err != nil {
		return nil, err
	}
	d.SingularTable(true)

	return &StaticDbProvider{db: &d}, nil
}

type StaticDbProvider struct {
	db *gorm.DB
}

func (p *StaticDbProvider) Get() (*gorm.DB, error) {
	return p.db, nil
}

// Dynamic DbProvider builds a gorm.DB connection
// - using consul to discover mysql address
// - using vault to generate creds
func NewDbProvider(table string, serviceAddress string) (*DynamicDbProvider, error) {
	lb := clb.NewDefaultClb(clb.RoundRobin)

	client, err := DefaultAppIdClient()
	if err != nil {
		return nil, err
	}
	return &DynamicDbProvider{
		vault:  client,
		table:  table,
		lb:     lb,
		svcAdd: serviceAddress,
	}, nil
}

type DynamicDbProvider struct {
	vault  *api.Client
	table  string
	svcAdd string
	lb     clb.LoadBalancer
}

func (p *DynamicDbProvider) Get() (*gorm.DB, error) {
	if token := p.vault.Token(); token != "" {
		log.Printf("using token client %s", token)
	} else {
		log.Fatal("no VAULT_TOKEN supplied!")
	}

	sec, err := p.vault.Logical().Read("mysql/creds/todo")
	if err != nil {
		log.Fatal(err)
	}
	un, ok := sec.Data["username"].(string)
	if !ok {
		return nil, fmt.Errorf("username not found")
	}
	pw, ok := sec.Data["password"].(string)
	if !ok {
		return nil, fmt.Errorf("password not found")
	}
	return p.createDbConnection(un, pw)
}
func (p *DynamicDbProvider) createDbConnection(user string, pass string) (*gorm.DB, error) {
	add, err := p.lb.GetAddress(p.svcAdd)
	if err != nil {
		return nil, err
	}
	str := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True", user, pass, add, p.table)

	static, err := NewStaticDbProvider(str)
	if err != nil {
		return nil, err
	}
	return static.Get()
}