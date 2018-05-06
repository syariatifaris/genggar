package main

import (
	"log"
	"sync"

	"time"

	"github.com/syariatifaris/genggar"
	"github.com/syariatifaris/genggar/engine"
	"github.com/syariatifaris/genggar/glog"
)

var wg sync.WaitGroup

func main() {
	server, err := genggar.NewEventServer("127.0.0.1", 1234)
	if err != nil {
		glog.ERROR.Fatalln("server error", err.Error())
	}

	glog.INFO.Println("server starting on 127.0.0.1:1234")

	stopServer := make(chan bool)
	wg.Add(1)
	go func() {
		server.Start(stopServer)
	}()

	stopDispatcher := make(chan bool)
	wg.Add(1)
	go func() {
		server.DispatchEventPublisher(stopDispatcher)
	}()

	wg.Add(1)
	go func() {
		time.Sleep(time.Second * 100)
		stopServer <- true
		stopDispatcher <- true
		wg.Done()

		glog.INFO.Println("time elapsed")
	}()

	wg.Add(1)
	go func() {
		for {
			publishFor(server, "ORDER")
		}
	}()

	wg.Wait()
}

func publishFor(server engine.Server, topic string) {
	err := server.PublishEvent(topic, "NEW_ORDER_VERIFIED", "publish event message from server")

	if err != nil {
		log.Println("cannot publish event to all", err.Error())
	}

	time.Sleep(time.Second * 1)
}
