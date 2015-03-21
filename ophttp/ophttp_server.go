package ophttp

import (
	"github.com/hydrogen18/stoppableListener"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// http server gracefull exit on SIGINT
func StartServer(bind string) error {

	originalListener, err := net.Listen("tcp", bind)
	if err != nil {
		return err
	}

	sl, err := stoppableListener.New(originalListener)
	if err != nil {
		return err
	}

	server := http.Server{}

	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, syscall.SIGINT)
	var wg sync.WaitGroup
	go func() {
		wg.Add(1)
		defer wg.Done()
		server.Serve(sl)
	}()

	log.Println("Serving HTTP")
	select {
	case signal := <-stopChan:
		log.Printf("Got signal:%v\n", signal)
	}
	log.Println("Stopping listener")
	sl.Stop()
	log.Println("Waiting on server")
	wg.Wait()

	return nil
}
