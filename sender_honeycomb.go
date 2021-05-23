package main

import (
	"crypto/rand"
	"fmt"

	libhoney "github.com/honeycombio/libhoney-go"
)

type HoneycombSender struct{}

func NewHoneycombSender(apikey string) *HoneycombSender {

	libhoney.Init(libhoney.Config{
		WriteKey: apikey,
		Dataset:  "vault-observe",
	})

	return &HoneycombSender{}
}

func (h *HoneycombSender) Send(typed Event, event map[string]interface{}) error {

	ev := libhoney.NewEvent()
	ev.AddField("trace.trace_id", typed.Request.ID)

	if typed.Type == "request" {
		ev.AddField("trace.span_id", typed.Request.ID)
	} else {
		ev.AddField("trace.parent_id", typed.Request.ID)
		ev.AddField("trace.span_id", generateSpanID())
	}

	ev.AddField("service_name", "vault")
	ev.AddField("name", typed.Type)

	for key, val := range event {
		ev.AddField(key, val)
	}

	ev.Send()
	return nil
}

func (h *HoneycombSender) Shutdown() error {
	libhoney.Close()
	return nil
}

func generateSpanID() string {
	spanBytes := make([]byte, 16)
	rand.Read(spanBytes)

	return fmt.Sprintf("%x", spanBytes)
}
