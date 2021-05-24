package main

type Sender interface {
	Send(Event, map[string]interface{}) error
	Shutdown() error
}

type Event struct {
	Type  string
	Error string

	Request Request
}

type Request struct {
	ID string
}
