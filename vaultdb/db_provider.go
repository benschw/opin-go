package vaultdb

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type DbProvider interface {
	Get() (*gorm.DB, error)
}

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
