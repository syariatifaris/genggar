package genggar

import (
	"fmt"
	"net"

	"github.com/syariatifaris/genggar/engine"
	"github.com/syariatifaris/genggar/subscriber"
)

func NewEventServer(serverAddr string, port int) (engine.Server, error) {
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(serverAddr),
	}

	server, err := net.ListenUDP(engine.ProtoUDP, &addr)
	if err != nil {
		fmt.Printf("Some error %v\n", err)
		return nil, err
	}

	return &engine.ServerImpl{
		Port:  port,
		Proto: engine.ProtoUDP,

		ServerConn:  server,
		MsgBuff:     make([]byte, engine.MaxBuffer),
		Subscribers: make(map[string]subscriber.Client),
	}, nil
}
