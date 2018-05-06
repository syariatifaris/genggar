package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"github.com/syariatifaris/genggar/glog"
	"github.com/syariatifaris/genggar/subscriber"
	"github.com/syariatifaris/genggar/util"
)

const (
	CmdReg   = "[REG]"
	CmdInfo  = "[INF]"
	CmdRetry = "[RET]"
	CmdEvent = "[EVT]"
)

type property struct {
	msg    []byte
	data   interface{}
	addr   *net.UDPAddr
	server Server
	client Client
}

type processor interface {
	exec() error
}

func getProcessor(prop *property) (processor, error) {
	var message Message
	err := json.Unmarshal(prop.msg, &message)
	if err != nil {
		return nil, err
	}
	prop.data = message.Data

	switch message.Cmd {
	case CmdReg:
		return &registerProcessor{
			prop: prop,
		}, nil
	case CmdEvent:
		return &eventProcessor{
			prop: prop,
		}, nil
	}
	return nil, errors.New("undefined processor")
}

//Region Register Processor

type registerProcessor struct {
	prop *property
}

func (r *registerProcessor) getSubscribedTopic() (string, error) {
	var rMsg RegisterMessage
	data, err := json.Marshal(r.prop.data)
	if err != nil {
		return "", errors.New(fmt.Sprint("obtain topic fail", err.Error()))
	}

	err = json.Unmarshal(data, &rMsg)
	if err != nil {
		return "", errors.New(fmt.Sprint("obtain topic fail", err.Error()))
	}

	return rMsg.Topic, nil
}

func (r *registerProcessor) exec() error {
	name := fmt.Sprint(r.prop.addr.IP.String(), ":", r.prop.addr.Port)

	sub, err := r.prop.server.getSubscriber(name)
	if err == nil && sub != nil {
		glog.INFO.Println("subscrieber exists", name)
		return nil
	}

	topic, err := r.getSubscribedTopic()
	if err != nil {
		return err
	}

	client, err := subscriber.NewClient(subscriber.Property{
		Address:   r.prop.addr,
		Name:      name,
		MaxBuffer: MaxBuffer,
		Topic:     topic,
	})

	if err != nil {
		glog.ERROR.Println("fail create client", err.Error())
		return err
	}
	r.prop.server.addSubscriber(name, client)

	msg, err := json.Marshal(Message{
		Cmd: CmdInfo,
		Msg: "client registration success",
	})

	r.prop.server.sendData(msg, client.GetUDPAddr())
	return nil
}

//Region Event Accept Processor

type eventProcessor struct {
	prop *property
}

func (r *eventProcessor) getEvent() (string, error) {
	var eMsg EventMessage
	data, err := json.Marshal(r.prop.data)
	if err != nil {
		return "", errors.New(fmt.Sprint("obtain event fail", err.Error()))
	}

	err = json.Unmarshal(data, &eMsg)
	if err != nil {
		return "", errors.New(fmt.Sprint("obtain event fail", err.Error()))
	}

	return eMsg.Event, nil
}

func (r *eventProcessor) exec() error {
	if r.prop.client == nil {
		return errors.New("client does not exist")
	}

	event, err := r.getEvent()
	if err != nil {
		return err
	}

	processors := r.prop.client.getEventProcessors()
	for _, proc := range processors {
		if proc.Events == nil {
			return errors.New("processors empty")
		}

		if util.InArrayStr(event, proc.Events) {
			err := proc.Callback(r.prop.client.getTopic(), event, r.prop.data)
			if err != nil {
				errMsg := fmt.Sprintf("processor error for %s %s", event, err.Error())
				return errors.New(errMsg)
			}
		}
	}

	return nil
}
