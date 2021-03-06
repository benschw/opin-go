package vault

import (
	"fmt"
	"log"
	"time"

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
func NewDbProvider(dbName string, svcName string) (*DynamicDbProvider, error) {

	ch := make(chan connectionRequest)
	lb := clb.NewDefaultClb(clb.RoundRobin)

	go dbConnectionCreator(ch, lb, svcName, dbName)

	return &DynamicDbProvider{reqCh: ch}, nil
}

type connectionRequest struct {
	Exit bool
	Resp chan connectionResponse
}
type connectionResponse struct {
	Db  gorm.DB
	Err error
}

func dbConnectionCreator(reqCh chan connectionRequest, lb clb.LoadBalancer, svcName string, dbName string) {
	var db *gorm.DB
	var lastUpdated int64
	var leaseDuration int64
	for req := range reqCh {
		now := time.Now().Unix()
		resp := connectionResponse{}
		if req.Exit {
			req.Resp <- resp
			return
		}
		if now-lastUpdated > leaseDuration {
			vault, err := DefaultAppIdClient()
			if err != nil {
				resp.Err = fmt.Errorf("problem building app id client: %s", err)
			} else {
				dur, un, pw, err := getDbCreds(vault)
				if err != nil {
					resp.Err = fmt.Errorf("problem getting db creds: %s", err)
				} else {
					log.Print("db credentials generated by vault")
					db, err = createDbConnection(lb, un, pw, dbName, svcName)
					if err != nil {
						resp.Err = err
					}
					lastUpdated = now
					if leaseDuration != int64(dur) {
						leaseDuration = int64(dur)
						log.Printf("db cred lease updated to %ds", leaseDuration)
					}
				}
			}
		}
		if db != nil {
			resp.Db = *db
		}
		req.Resp <- resp
	}
}
func getDbCreds(vault *api.Client) (int, string, string, error) {
	sec, err := vault.Logical().Read("mysql/creds/todo")
	if err != nil {
		return 0, "", "", err
	}

	dur := sec.LeaseDuration

	un, ok := sec.Data["username"].(string)
	if !ok {
		return 0, "", "", fmt.Errorf("mysql username not found")
	}
	pw, ok := sec.Data["password"].(string)
	if !ok {
		return 0, "", "", fmt.Errorf("mysql password not found")
	}
	return dur, un, pw, nil
}
func createDbConnection(lb clb.LoadBalancer, user string, pass string, dbName string, svcName string) (*gorm.DB, error) {
	add, err := lb.GetAddress(svcName)
	if err != nil {
		return nil, err
	}

	dbStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True", user, pass, add, dbName)
	return openConnection(dbStr)
}

type DynamicDbProvider struct {
	reqCh chan connectionRequest
}

func (p *DynamicDbProvider) Get() (*gorm.DB, error) {
	req := connectionRequest{Resp: make(chan connectionResponse)}
	p.reqCh <- req
	resp := <-req.Resp

	return &resp.Db, resp.Err
}
func (p *DynamicDbProvider) Close() {
	req := connectionRequest{Exit: true, Resp: make(chan connectionResponse)}
	p.reqCh <- req
	<-req.Resp
}
