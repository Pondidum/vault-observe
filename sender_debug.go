package main

import "fmt"

type DebugSender struct{}

func NewDebugSender() *DebugSender {
	return &DebugSender{}
}

func (c *DebugSender) Send(typed Event, event map[string]interface{}) error {

	fmt.Println(typed.Type + ":")

	for key, value := range event {
		fmt.Printf("  %s: %v\n", key, value)
	}

	return nil
}
