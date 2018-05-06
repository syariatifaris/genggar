package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	"time"

	"github.com/syariatifaris/genggar/glog"
	"github.com/syariatifaris/genggar/subscriber"
	"github.com/syariatifaris/genggar/util"
)

const (
	MaxBuffer = 1024
	ProtoUDP  = "udp"
)

type Server interface {
	//region public functions
	Start(stopChan <-chan bool)
	DispatchEventPublisher(stopChan <-chan bool)
	PublishEvent(topic, event, message string) error

	//region private functions
	registerSubscriber(name string, addr *net.UDPAddr) error
	sendData(msg []byte, addr *net.UDPAddr) error
	getSubscriber(name string) (subscriber.Client, error)
	addSubscriber(string, subscriber.Client)
}

type ServerImpl struct {
	Port  int
	Proto string

	mux         sync.Mutex
	MsgBuff     []byte
	ServerConn  *net.UDPConn
	Subscribers map[string]subscriber.Client

	isStarted bool
}

//Start listens for incoming client
func (s *ServerImpl) Start(stopListen <-chan bool) {
	s.isStarted = true
	serverChan := make(chan bool)

	go func(stopListen <-chan bool) {
		for {
			select {
			case <-stopListen:
				s.isStarted = false
				s.ServerConn.Close()
				serverChan <- true
				return
			default:
				//do something
			}
		}
	}(stopListen)

	for {
		select {
		case <-serverChan:
			return
		default:
			n, addr, err := s.ServerConn.ReadFromUDP(s.MsgBuff)
			if err != nil {
				if !s.isStarted {
					return
				}

				glog.ERROR.Println("udp error", err.Error())
				continue
			}

			processor, err := getProcessor(&property{
				msg:    s.MsgBuff[0:n],
				server: s,
				addr:   addr,
			})
			if err != nil {
				glog.ERROR.Println("unable to resolve process", err.Error())
				continue

			}

			err = processor.exec()
			if err != nil {
				glog.ERROR.Println("unable to exec process", err.Error())
				continue
			}
		}
	}
}

//DispatchEventPublisher dispatch event publisher to read the buffer
func (s *ServerImpl) DispatchEventPublisher(stopChan <-chan bool) {
	stopDispatchChan := make(chan bool)
	for {
		select {
		case <-stopChan:
			stopDispatchChan <- true
			return
		default:
			if len(s.Subscribers) > 0 {
				s.mux.Lock()
				for _, sb := range s.Subscribers {
					if !sb.IsDispatched() {
						sb.SetDispatching(true)
						go s.handleEventBuffer(sb, stopDispatchChan)
						glog.DEBUG.Println("dispatch for", sb.GetUDPAddr().String())
					}
				}

				s.mux.Unlock()
			}
			time.Sleep(time.Millisecond * 10)
		}
	}
}

//PublishEvent publishes event based on client identifier
func (s *ServerImpl) PublishEvent(topic, event, message string) error {
	uuid, _ := util.GetV4UUID()
	data := Message{
		Cmd: CmdEvent,
		Msg: message,
		Data: EventMessage{
			Event: event,
			UUID:  uuid,
		},
	}

	if s.ServerConn == nil {
		if s.Subscribers != nil {
			errors.New("no subscriber found")
		}

		errors.New("server connection is closed")
	}

	for _, sub := range s.Subscribers {
		if sub.GetTopicName() == topic {
			err := sub.PushBack(data)
			if err != nil {
				return errors.New("unable to push data to buffer")
			}
		}
	}

	return nil
}

//handleEventBuffer reads the data from event buffer and send the data
func (s *ServerImpl) handleEventBuffer(sb subscriber.Client, stopDispatchChan <-chan bool) {
	for {
		select {
		case <-stopDispatchChan:
			return
		default:
			if s.isStarted {
				if sb.GetBufferLen() > 0 {
					data, err := sb.PopFront()
					if err != nil {
						log.Println("pop fail", err.Error())
						continue
					}

					msg, err := json.Marshal(data)
					if err != nil {
						log.Println("marshall fail", err.Error())
						sb.PushBack(msg)
						continue
					}

					//perform send data through UDP
					err = s.sendData(msg, sb.GetUDPAddr())
					if err != nil {
						log.Println("send data fail", err.Error())
						sb.PushBack(msg)
						continue
					}
				}
			}
		}
	}
}

//registerSubscriber register new clients, add to pool
func (s *ServerImpl) registerSubscriber(name string, addr *net.UDPAddr) error {
	if _, ok := s.Subscribers[name]; !ok {
		client, err := subscriber.NewClient(subscriber.Property{
			Address:   addr,
			Name:      name,
			MaxBuffer: MaxBuffer,
		})

		if err != nil {
			glog.ERROR.Println("create client fail", err.Error())
			return err
		}

		s.Subscribers[name] = client
		return nil
	}

	return errors.New("subscriber exists")
}

//getSubscriber gets the subscriber
func (s *ServerImpl) getSubscriber(name string) (subscriber.Client, error) {
	if s, ok := s.Subscribers[name]; ok {
		return s, nil
	}
	return nil, errors.New(fmt.Sprint("subsriber not found", name))
}

//addSubscriber add subscriber to subscriber pool
func (s *ServerImpl) addSubscriber(name string, subs subscriber.Client) {
	s.Subscribers[name] = subs
}

//sendData sends the data through UDP
func (s *ServerImpl) sendData(msg []byte, addr *net.UDPAddr) error {
	if s.ServerConn != nil {
		_, err := s.ServerConn.WriteToUDP(msg, addr)
		if err != nil {
			return err
		}

		glog.DEBUG.Println("sending to ", addr.String(), ":", string(msg))
		return err
	}
	return errors.New("server unavailable")
}
