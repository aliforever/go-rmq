package rmq

type RmqOptions struct {
	reconnectTries int
	reconnectDelay int
}

func NewRmqOptions() *RmqOptions {
	return &RmqOptions{}
}

func (r *RmqOptions) SetReconnectTries(tries int) *RmqOptions {
	r.reconnectTries = tries

	return r
}

func (r *RmqOptions) SetReconnectDelay(delay int) *RmqOptions {
	r.reconnectDelay = delay

	return r
}
