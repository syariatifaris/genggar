package engine

import (
	"fmt"
	"net"

	"github.com/syariatifaris/genggar/subscriber"
)

const (
	MaxBuffer = 1024
	CmdLength = 5
	ProtoUDP  = "udp"
)

const (
	CmdReg = "CmdReg"
)

type Server interface {
	ListenForClient()
	DispatchEventPublisher()
	PublishEvent(clientName string, data interface{}) error
	PublishEventToAll(data interface{}) error
	CloseConn() error
}

func NewEventServer(serverAddr string, port int) (Server, error) {
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(serverAddr),
	}

	server, err := net.ListenUDP(ProtoUDP, &addr)
	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return nil, err
	}

	return &serverImpl{
		port:  port,
		proto: ProtoUDP,

		serverConn:  server,
		msgBuff:     make([]byte, MaxBuffer),
		subscribers: make(map[string]subscriber.Client),
	}, nil
}
