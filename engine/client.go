package engine

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"github.com/syariatifaris/genggar/glog"
)

type EventFunc func(topic, eventName string, data interface{}) error

type EventProcessor struct {
	Events   []string
	Callback EventFunc
}

type Client interface {
	StartListen(stopChan <-chan bool)
	getEventProcessors() []*EventProcessor
	getTopic() string
}

type ClientImpl struct {
	Port       int
	Proto      string
	Topic      string
	MsgBuff    []byte
	ServerAddr string
	ClientConn net.Conn
	Processors []*EventProcessor

	isStarted bool
}

func (c *ClientImpl) getEventProcessors() []*EventProcessor {
	return c.Processors
}

func (c *ClientImpl) getTopic() string {
	return c.Topic
}

func (c *ClientImpl) StartListen(stopChan <-chan bool) {
	c.isStarted = true
	clientChan := make(chan bool)

	go func(stopListen <-chan bool) {
		for {
			select {
			case <-stopListen:
				c.isStarted = false
				c.ClientConn.Close()
				clientChan <- true
				return
			default:
				//do something
			}
		}
	}(stopChan)

	err := c.registerTopic(c.Topic)
	if err != nil {
		glog.ERROR.Fatalln("subscribe err", err.Error())
		return
	}

	for {
		select {
		case <-clientChan:
			return
		default:
			n, err := bufio.NewReader(c.ClientConn).Read(c.MsgBuff)
			if err != nil {
				if !c.isStarted {
					return
				}

				glog.DEBUG.Println("server read error", err.Error())
				continue
			}

			glog.DEBUG.Println("[server says]:", string(c.MsgBuff[0:n]))
			processor, err := getProcessor(&property{
				msg:    c.MsgBuff[0:n],
				client: c,
			})

			if err != nil {
				glog.DEBUG.Println("unable to resolve process", err.Error())
				continue
			}

			err = processor.exec()
			if err != nil {
				glog.DEBUG.Println("unable to exec process", err.Error())
				continue
			}
		}
	}
}

func (c *ClientImpl) registerTopic(topic string) error {
	if c.ClientConn == nil {
		return errors.New("connection closed")
	}

	cmd := Message{
		Msg: "client do registration",
		Cmd: CmdReg,
		Data: RegisterMessage{
			Topic: topic,
		},
	}

	msg, err := json.Marshal(cmd)
	if err != nil {
		return errors.New(fmt.Sprint("marshall error", err.Error()))
	}

	glog.DEBUG.Println("sending command", string(msg))
	_, err = fmt.Fprintf(c.ClientConn, string(msg))
	if err != nil {
		return errors.New(fmt.Sprint("command err", err.Error()))
	}

	return nil
}
