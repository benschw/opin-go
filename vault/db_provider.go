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
	Close()
}

func openConnection(dbStr string) (*gorm.DB, error) {
	d, err := gorm.Open("mysql", dbStr)
	if err == nil {
		d.SingularTable(true)
	}

	return &d, err
}

// Static DbProvider builds a gorm.DB connection
// - using a static connection string
func NewStaticDbProvider(dbStr string) (*StaticDbProvider, error) {
	d, err := openConnection(dbStr)
	if err != nil {
		return nil, err
	}

	return &StaticDbProvider{db: d}, nil
}

type StaticDbProvider struct {
	db *gorm.DB
}

func (p *StaticDbProvider) Get() (*gorm.DB, error) {
	return p.db, nil
}
func (p *StaticDbProvider) Close() {
}

// Dynamic DbProvider builds a gorm.DB connection
// - using consul to discover mysql address
// - using vault to generate creds
func NewDbProvider(table string, svcName string) (*DynamicDbProvider, error) {

	ch := make(chan connectionRequest)
	lb := clb.NewDefaultClb(clb.RoundRobin)

	go dbConnectionCreator(ch, lb, svcName, table)

	return &DynamicDbProvider{
		reqCh: ch,
	}, nil
}

type connectionRequest struct {
	Exit bool
	Resp chan connectionResponse
}
type connectionResponse struct {
	Db  gorm.DB
	Err error
}

func dbConnectionCreator(reqCh chan connectionRequest, lb clb.LoadBalancer, svcName string, table string) {
	var db *gorm.DB

	for req := range reqCh {
		resp := connectionResponse{}
		if req.Exit {
			req.Resp <- resp
			return
		}
		if db == nil {
			vault, err := DefaultAppIdClient()
			if err != nil {
				log.Printf("problem building app id client: %s", err)
				resp.Err = err
			} else {
				un, pw, err := getDbCreds(vault)
				if err != nil {
					log.Printf("problem getting db creds: %s", err)
					resp.Err = err
				} else {
					log.Print("db credentials generated by vault")
					db, err = createDbConnection(lb, un, pw, table, svcName)
					if err != nil {
						resp.Err = err
					}
				}
			}
		}
		resp.Db = *db
		req.Resp <- resp
	}
}
func getDbCreds(vault *api.Client) (string, string, error) {
	sec, err := vault.Logical().Read("mysql/creds/todo")
	if err != nil {
		return "", "", err
	}
	un, ok := sec.Data["username"].(string)
	if !ok {
		return "", "", fmt.Errorf("mysql username not found")
	}
	pw, ok := sec.Data["password"].(string)
	if !ok {
		return "", "", fmt.Errorf("mysql password not found")
	}
	return un, pw, nil
}
func createDbConnection(lb clb.LoadBalancer, user string, pass string, table string, svcName string) (*gorm.DB, error) {
	add, err := lb.GetAddress(svcName)
	if err != nil {
		return nil, err
	}

	dbStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True", user, pass, add, table)
	return openConnection(dbStr)
}

type DynamicDbProvider struct {
	vault  *api.Client
	table  string
	svcAdd string
	lb     clb.LoadBalancer
	reqCh  chan connectionRequest
}

func (p *DynamicDbProvider) Get() (*gorm.DB, error) {
	ch := make(chan connectionResponse)
	p.reqCh <- connectionRequest{Resp: ch}

	resp := <-ch
	return &resp.Db, resp.Err
}
func (p *DynamicDbProvider) Close() {
	p.reqCh <- connectionRequest{Exit: true}
}
