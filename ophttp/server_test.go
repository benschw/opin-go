package ophttp

import (
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/benschw/opin-go/rando"
	"github.com/benschw/opin-go/rest"
	"github.com/gorilla/mux"
)

func Handle(res http.ResponseWriter, req *http.Request) {
	rest.SetOKResponse(res, nil)
}

func NewApp(bind string) *App {
	server := NewServer(bind)
	return &App{
		S: server,
	}
}

type App struct {
	S *Server
}

func (s *App) Start() error {
	log.Println("HTTP: Starting Demo Http Server")

	r := mux.NewRouter()

	r.HandleFunc("/foo", Handle).Methods("GET")

	mux := http.NewServeMux()
	mux.Handle("/", r)
	//http.Handle("/", r)

	return s.S.Start(mux)
}
func (s *App) Stop() {
	s.S.Stop()
}

func TestServerStops(t *testing.T) {
	port := uint16(rando.Port())
	add := fmt.Sprintf(":%d", port)
	a := NewApp(add)
	go a.Start()

	r, err := http.Get("http://golang.org/")
	if err != nil {
		t.Errorf("%s", err)
	}
	if r.StatusCode != 200 {
		t.Errorf("%s", err)
	}
	a.Stop()

	a2 := NewApp(add)
	go a2.Start()
	a2.Stop()
}
