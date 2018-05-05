package subscriber

import (
	"container/list"
	"errors"
	"log"
	"net"
	"sync"
)

var (
	ErrBufferFull = errors.New("buffer full")
)

type Client interface {
	PushBack(data interface{}) error
	PushFront(data interface{}) error
	PopFront() (interface{}, error)
	GetBufferLen() int

	GetUDPAddr() *net.UDPAddr
	SetDispatching(bool)
	IsDispatched() bool

	LogAllElemFront()
}

type Property struct {
	Name    string
	Address *net.UDPAddr
	//to do
	MaxBuffer int
}

type clientImpl struct {
	mux        sync.Mutex
	prop       Property
	evtBuffer  *list.List
	dispatched bool
}

func NewClient(prop Property) (Client, error) {
	if prop.MaxBuffer == 0 {
		return nil, errors.New("buffer length should more than 0")
	}
	return &clientImpl{
		prop:      prop,
		evtBuffer: list.New(),
	}, nil
}

func (c *clientImpl) PushBack(data interface{}) error {
	if c.evtBuffer != nil {
		if c.evtBuffer.Len() <= c.prop.MaxBuffer {
			c.mux.Lock()
			c.evtBuffer.PushBack(data)
			c.mux.Unlock()
			return nil
		}
		return ErrBufferFull
	}
	return errors.New("buffer is not initialized")
}

func (c *clientImpl) PushFront(data interface{}) error {
	if c.evtBuffer != nil {
		if c.evtBuffer.Len() <= c.prop.MaxBuffer {
			c.mux.Lock()
			c.evtBuffer.PushFront(data)
			c.mux.Unlock()
			return nil
		}
		return ErrBufferFull
	}
	return errors.New("buffer is not initialized")
}

func (c *clientImpl) PopFront() (interface{}, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.evtBuffer != nil {
		if c.evtBuffer.Len() > 0 {
			elem := c.evtBuffer.Front()
			defer func() {
				if elem != nil {
					c.evtBuffer.Remove(elem)
				}
			}()
			return elem.Value, nil
		}
		return nil, errors.New("buffer empty")
	}
	return nil, errors.New("buffer is not initialized")
}

func (c *clientImpl) GetBufferLen() int {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.evtBuffer.Len()
}

func (c *clientImpl) GetUDPAddr() *net.UDPAddr {
	return c.prop.Address
}

func (c *clientImpl) LogAllElemFront() {
	if c.evtBuffer.Len() == 0 {
		log.Println("buffer empty")
	}

	for el := c.evtBuffer.Front(); el != nil; el = el.Next() {
		log.Println("buffer data:", el.Value)
	}
}

func (c *clientImpl) SetDispatching(bool) {
	c.mux.Lock()
	c.dispatched = true
	c.mux.Unlock()
}

func (c *clientImpl) IsDispatched() bool {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.dispatched
}
