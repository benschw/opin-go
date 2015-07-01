package vaultdb

import (
	"fmt"

	"github.com/benschw/dns-clb-go/clb"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type DbProvider interface {
	Get() (*gorm.DB, error)
}

// Static DbProvider builds a gorm.DB connection
// - using a static connection string
func NewStatic(dbStr string) (*StaticDbProvider, error) {
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
func New(table string, serviceAddress string) *DynamicDbProvider {
	lb := clb.NewDefaultClb(clb.RoundRobin)
	return &DynamicDbProvider{
		table:  table,
		lb:     lb,
		svcAdd: serviceAddress,
	}
}

type DynamicDbProvider struct {
	table  string
	svcAdd string
	lb     clb.LoadBalancer
}

func (p *DynamicDbProvider) Get() (*gorm.DB, error) {
	return nil, nil
}
func (p *DynamicDbProvider) createDbConnection(user string, pass string) (*gorm.DB, error) {
	add, err := p.lb.GetAddress(p.svcAdd)
	if err != nil {
		return nil, err
	}
	str := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True", user, pass, add, p.table)

	static, err := NewStatic(str)
	if err != nil {
		return nil, err
	}
	return static.Get()
}
