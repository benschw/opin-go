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
	return s.Start()
}

func NewServer(bind string) *Server {
	return &Server{
		Bind:     bind,
		StopChan: make(chan os.Signal),
	}
}

type Server struct {
	Bind     string
	StopChan chan os.Signal
}

// http server gracefull exit on SIGINT
func (s *Server) Start() error {

	originalListener, err := net.Listen("tcp", s.Bind)
	if err != nil {
		return err
	}

	sl, err := stoppableListener.New(originalListener)
	if err != nil {
		return err
	}

	server := http.Server{}

	signal.Notify(s.StopChan, syscall.SIGINT)
	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		defer wg.Done()
		server.Serve(sl)
	}()

	select {
	case signal := <-s.StopChan:
		log.Printf("Got signal:%v\n", signal)
	}
	sl.Stop()
	wg.Wait()

	return nil
}

func (s *Server) Stop() {
	s.StopChan <- syscall.SIGINT
}
