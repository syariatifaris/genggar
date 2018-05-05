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

type serverImpl struct {
	port  int
	proto string

	mux         sync.Mutex
	msgBuff     []byte
	serverConn  *net.UDPConn
	subscribers map[string]subscriber.Client
}

//ListenForClient listens for incoming client
func (s *serverImpl) ListenForClient() {
	for {
		n, addr, err := s.serverConn.ReadFromUDP(s.msgBuff)
		if err != nil {
			log.Println("udp error", err.Error())
			continue
		}

		msg := string(s.msgBuff[0:n])
		if len(msg) < CmdLength {
			continue
		}

		cmd := string(s.msgBuff[0:CmdLength])

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

//PublishEvent publishes event based on client identifier
func (s *serverImpl) PublishEvent(clientName string, data interface{}) error {
	if data == nil {
		return errors.New("sent data cannot be null")
	}

	if s.serverConn == nil {
		if s.subscribers != nil {
			errors.New("no subscriber found")
		}

		errors.New("server connection is closed")
	}

	if _, ok := s.subscribers[clientName]; !ok {
		return errors.New(fmt.Sprint("client not found ", clientName))
	}

	sub := s.subscribers[clientName]
	err := sub.PushBack(data)
	if err != nil {
		return errors.New("unable to push data to buffer")
	}

	return nil
}

//PublishEventToAll broadcast to all available connections
func (s *serverImpl) PublishEventToAll(data interface{}) error {
	if data == nil {
		return errors.New("sent data cannot be null")
	}

	if s.serverConn == nil {
		if s.subscribers != nil {
			errors.New("no subscriber found")
		}

		errors.New("server connection is closed")
	}

	for _, sub := range s.subscribers {
		err := sub.PushBack(data)
		if err != nil {
			return errors.New("unable to push data to buffer")
		}
	}

	return nil
}

//DispatchEventPublisher dispatch event publisher to read the buffer
func (s *serverImpl) DispatchEventPublisher() {
	for {
		if len(s.subscribers) > 0 {
			s.mux.Lock()
			for _, sb := range s.subscribers {
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
func (s *serverImpl) handleEventBuffer(sb subscriber.Client) {
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
func (s *serverImpl) CloseConn() error {
	if s.serverConn != nil {
		return s.serverConn.Close()
	}
	return errors.New("empty connection")
}

//registerSubscriber register new clients, add to pool
func (s *serverImpl) registerSubscriber(name string, addr *net.UDPAddr) error {
	if _, ok := s.subscribers[name]; !ok {
		client, err := subscriber.NewClient(subscriber.Property{
			Address:   addr,
			Name:      name,
			MaxBuffer: MaxBuffer,
		})

		if err != nil {
			log.Println("create client fail", err.Error())
			return err
		}

		s.subscribers[name] = client
		return nil
	}

	return errors.New("subscriber exists")
}

//getSubscriber gets the subscriber
func (s *serverImpl) getSubscriber(name string) (subscriber.Client, error) {
	if s, ok := s.subscribers[name]; ok {
		return s, nil
	}
	return nil, errors.New(fmt.Sprint("subsriber not found", name))
}

//sendData sends the data through UDP
func (s *serverImpl) sendData(msg []byte, addr *net.UDPAddr) error {
	if s.serverConn != nil {
		_, err := s.serverConn.WriteToUDP(msg, addr)
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
