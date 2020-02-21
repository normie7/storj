package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/normie7/storj/filesender"
)

func main() {

	addr := parseArgs()

	r, err := filesender.NewRelay(addr)
	if err != nil {
		log.Fatal(err)
	}

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err := r.Start()
		if err != nil {
			log.Fatal("Error listening: ", err)
		}
	}()

	<-termChan
	r.Stop()
}

func parseArgs() (addr string) {
	if len(os.Args) != 2 {
		log.Fatal("expected 'address' argument")
	}

	p := strings.Split(os.Args[1], ":")
	if len(p) != 2 || p[0] != "" {
		log.Fatal("wrong address format. try :relayPort")
	}

	return os.Args[1]
}
