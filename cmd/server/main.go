package main

import (
	"log"
	"sync"

	"fmt"

	"time"

	"github.com/syariatifaris/genggar"
	"github.com/syariatifaris/genggar/engine"
)

var wg sync.WaitGroup

type MockData struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
}

func main() {
	server, err := genggar.NewEventServer("127.0.0.1", 1234)
	if err != nil {
		log.Fatalln("server error", err.Error())
	}
	defer server.CloseConn()

	log.Println("server starting on 127.0.0.1:1234")

	stop := make(chan bool)
	wg.Add(1)
	go func() {
		server.Start(stop)
	}()

	wg.Add(1)
	go func() {
		time.Sleep(time.Second * 2)
		log.Println("time elapsed")
		stop <- true
		wg.Done()
	}()

	//var i int
	//wg.Add(1)
	//go server.DispatchEventPublisher()
	//
	//wg.Add(1)
	//go func() {
	//	for {
	//		publishFromChannel(i, "First Channel", server)
	//		i++
	//	}
	//}()

	wg.Wait()
}

func publishFromChannel(count int, channel string, server engine.Server) {
	err := server.PublishEventToAll(MockData{
		ID:      count,
		Message: fmt.Sprintf("[Channel:%s] Send Data", channel),
	})

	if err != nil {
		log.Println("cannot publish event to all", err.Error())
	}

	time.Sleep(time.Second * 1)
}
