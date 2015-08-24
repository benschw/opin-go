package ophttp

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/hydrogen18/stoppableListener"
)

// http server gracefull exit on SIGINT
func StartServer(bind string) error {
	s := NewServer(bind)
	return s.Start(nil)
}

func NewServer(bind string) *Server {
	return &Server{
		Bind:     bind,
		StopChan: make(chan chan error),
		SigChan:  make(chan os.Signal),
	}
}

type Server struct {
	Bind     string
	StopChan chan chan error
	SigChan  chan os.Signal
}

// http server gracefull exit on SIGINT
func (s *Server) Start(mux http.Handler) error {
	if mux == nil {
		mux = http.DefaultServeMux
	}
	originalListener, err := net.Listen("tcp", s.Bind)
	if err != nil {
		return err
	}

	sl, err := stoppableListener.New(originalListener)
	if err != nil {
		return err
	}

	server := http.Server{
		Handler: mux,
	}

	signal.Notify(s.SigChan, syscall.SIGINT)
	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		defer wg.Done()
		server.Serve(sl)
	}()

	var exitNotice chan error

	select {
	case ch := <-s.StopChan:
		log.Printf("Stopping Server\n")
		exitNotice = ch
	case signal := <-s.SigChan:
		log.Printf("Got signal:%v\n", signal)
	}
	sl.Stop()
	wg.Wait()

	if exitNotice != nil {
		exitNotice <- nil
	}
	return nil
}

func (s *Server) Stop() {
	//s.StopChan <- syscall.SIGINT

	err := make(chan error)
	s.StopChan <- err

	<-err
}
