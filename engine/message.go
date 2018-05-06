package engine

type Message struct {
	Cmd  string      `json:"cmd"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type RegisterMessage struct {
	Topic string `json:"topic"`
}

type EventMessage struct {
	Event string `json:"event"`
	UUID  string `json:"uuid"`
}
