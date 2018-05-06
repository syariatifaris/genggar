package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	"time"

	"github.com/syariatifaris/genggar/subscriber"
)

type Server interface {
	Start(stopChan <-chan bool)
	DispatchEventPublisher(stopChan <-chan bool)
	PublishEvent(clientName string, data interface{}) error
	PublishEventToAll(data interface{}) error
	CloseConn() error
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
				log.Println("closing connection")
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
			log.Println("stop reading udp")
			return
		default:
			n, addr, err := s.ServerConn.ReadFromUDP(s.MsgBuff)
			if err != nil {
				if s.isStarted {
					log.Println("udp error", err.Error())
				}

				continue
			}

			msg := string(s.MsgBuff[0:n])
			if len(msg) < CmdLength {
				continue
			}

			cmd := string(s.MsgBuff[0:CmdLength])

			cmdType, err := parseCmd(cmd)
			if err != nil {
				log.Println("command fail", cmd, err.Error())
				continue
			}

			if cmdType == CmdReg {
				cltName := fmt.Sprint(addr.IP.String(), ":", addr.Port)

				err := s.registerSubscriber(cltName, addr)
				if err != nil {
					log.Println("cannot register client", cltName)
					continue
				}

				s.sendData([]byte("[INF]server registration ok!"), addr)
			}
		}
	}
}

//PublishEvent publishes event based on client identifier
func (s *ServerImpl) PublishEvent(clientName string, data interface{}) error {
	if data == nil {
		return errors.New("sent data cannot be null")
	}

	if s.ServerConn == nil {
		if s.Subscribers != nil {
			errors.New("no subscriber found")
		}

		errors.New("server connection is closed")
	}

	if _, ok := s.Subscribers[clientName]; !ok {
		return errors.New(fmt.Sprint("client not found ", clientName))
	}

	sub := s.Subscribers[clientName]
	err := sub.PushBack(data)
	if err != nil {
		return errors.New("unable to push data to buffer")
	}

	return nil
}

//PublishEventToAll broadcast to all available connections
func (s *ServerImpl) PublishEventToAll(data interface{}) error {
	if data == nil {
		return errors.New("sent data cannot be null")
	}

	if s.ServerConn == nil {
		if s.Subscribers != nil {
			errors.New("no subscriber found")
		}

		errors.New("server connection is closed")
	}

	for _, sub := range s.Subscribers {
		err := sub.PushBack(data)
		if err != nil {
			return errors.New("unable to push data to buffer")
		}
	}

	return nil
}

//DispatchEventPublisher dispatch event publisher to read the buffer
func (s *ServerImpl) DispatchEventPublisher() {
	for {
		if len(s.Subscribers) > 0 {
			s.mux.Lock()
			for _, sb := range s.Subscribers {
				if !sb.IsDispatched() {
					sb.SetDispatching(true)
					go s.handleEventBuffer(sb)
					log.Println("dispatch for", sb.GetUDPAddr().String())
				}
			}

			s.mux.Unlock()
		}
		time.Sleep(time.Millisecond * 10)
	}
}

//handleEventBuffer reads the data from event buffer and send the data
func (s *ServerImpl) handleEventBuffer(sb subscriber.Client) {
	for {
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

//CloseConn closes the connection
func (s *ServerImpl) CloseConn() error {
	if s.ServerConn != nil {
		return s.ServerConn.Close()
	}
	return errors.New("empty connection")
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
			log.Println("create client fail", err.Error())
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

//sendData sends the data through UDP
func (s *ServerImpl) sendData(msg []byte, addr *net.UDPAddr) error {
	if s.ServerConn != nil {
		_, err := s.ServerConn.WriteToUDP(msg, addr)
		if err != nil {
			return err
		}

		log.Println("sending to ", addr.String(), ":", string(msg))
		return err
	}
	return errors.New("server unavailable")
}

//parseCmd parses the command from client
func parseCmd(cmd string) (string, error) {
	switch cmd {
	case "[REG]":
		return CmdReg, nil
	default:
		return "", errors.New("invalid command")
	}
}
