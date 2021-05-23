package main

import (
	"fmt"
	"strings"
)

type CompositeSender struct {
	senders []Sender
}

func NewCompositeSender(senders []Sender) *CompositeSender {
	return &CompositeSender{senders: senders}
}

func (c *CompositeSender) Send(typed Event, event map[string]interface{}) error {
	errors := []string{}

	for _, sender := range c.senders {
		if err := sender.Send(typed, event); err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) == 1 {
		return fmt.Errorf(errors[0])
	}

	if len(errors) > 1 {
		return fmt.Errorf("Errors:\n\n%s", strings.Join(errors, "\n"))
	}

	return nil
}
