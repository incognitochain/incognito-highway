package process

type Sender interface {
	Send(data []byte, to interface{})
}
