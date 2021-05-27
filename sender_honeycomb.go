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

	duration := typed.Time.Sub(typed.StartTime).Milliseconds()

	ev := libhoney.NewEvent()
	ev.Timestamp = typed.StartTime
	ev.AddField("duration_ms", duration)
	ev.AddField("trace.trace_id", typed.Request.ID)
	ev.AddField("trace.span_id", typed.Request.ID)

	ev.AddField("service_name", "vault")
	ev.AddField("name", fmt.Sprintf("%s %s", typed.Request.Operation, typed.Request.Path))

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
