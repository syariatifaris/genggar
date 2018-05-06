package genggar

import (
	"net"

	"fmt"

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

func NewSubscriberClient(serverAddr string, port int, topic string, processors []*engine.EventProcessor) (engine.Client, error) {
	conn, err := net.Dial(engine.ProtoUDP, fmt.Sprint(serverAddr, ":", port))
	if err != nil {
		return nil, err
	}

	return &engine.ClientImpl{
		Proto: engine.ProtoUDP,
		Port:  port,
		Topic: topic,

		MsgBuff:    make([]byte, engine.MaxBuffer),
		ClientConn: conn,
		Processors: processors,
	}, nil
}
