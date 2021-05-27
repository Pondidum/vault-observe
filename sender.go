package main

import "time"

type Sender interface {
	Send(Event, map[string]interface{}) error
	Shutdown() error
}

type Event struct {
	Time  time.Time
	Type  string
	Error string

	Request Request

	StartTime time.Time
}

type Request struct {
	ID        string
	Operation string
	Path      string
}
