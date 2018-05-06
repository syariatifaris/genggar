package main

import (
	"sync"
	"time"

	"github.com/syariatifaris/genggar"
	"github.com/syariatifaris/genggar/engine"
	"github.com/syariatifaris/genggar/glog"
)

var wg sync.WaitGroup

func main() {
	port := 1234
	addr := "127.0.0.1"
	topic := "ORDER"

	client, err := genggar.NewSubscriberClient(
		addr, port, topic,
		[]*engine.EventProcessor{
			{
				Events:   []string{"NEW_ORDER_VERIFIED", "NEW_ORDER_WAITING"},
				Callback: newOrderEvent,
			},
			{
				Events:   []string{"REJECT_BY_SELLER", "REJECT_BY_SYSTEM"},
				Callback: rejectOrderEvent,
			},
		},
	)

	if err != nil {
		glog.ERROR.Fatalln("client err", err.Error())
		return
	}

	glog.INFO.Println("client started at", time.Now().String())

	stopChan := make(chan bool)
	wg.Add(1)
	go func() {
		client.StartListen(stopChan)
	}()

	wg.Add(1)
	go func() {
		time.Sleep(time.Second * 100)
		stopChan <- true
		glog.INFO.Println("time elapsed")
	}()

	wg.Wait()
}

func newOrderEvent(topic, eventName string, data interface{}) error {
	glog.INFO.Printf("[processor][topic:%s][event:%s] in processing\n", topic, eventName)
	return nil
}

func rejectOrderEvent(topic, eventName string, data interface{}) error {
	glog.INFO.Printf("[processor][topic:%s][event:%s] in processing\n", topic, eventName)
	return nil
}
