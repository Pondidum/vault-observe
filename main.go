package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/jeremywohl/flatten"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
)

func main() {
	err := run(os.Args)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}

func run(args []string) error {

	flags := pflag.NewFlagSet("vault-observe", pflag.ContinueOnError)
	useHoneycomb := flags.Bool("honeycomb", false, "enable sending to honeycomb")
	useZipkin := flags.Bool("zipkin", false, "enable sending to zipkin")
	useDebug := flags.Bool("debug", false, "enable sending to stdout")

	socketPath := flags.String("socket-path", "/tmp/vault-observe.sock", "the unix socket path for vault to send audit events to")

	if err := flags.Parse(args); err != nil {
		if err == pflag.ErrHelp {
			return nil
		}
		return err
	}

	senders := []Sender{}
	if *useHoneycomb {
		fmt.Println("Sending events to Honeycomb")
		honey := NewHoneycombSender(os.Getenv("HONEYCOMB_API_KEY"))
		honey.Init()
		senders = append(senders, honey)
	}

	if *useZipkin {
		fmt.Println("Sending events to Zipkin")
		otel := NewOtelSender()
		otel.Init()
		senders = append(senders, otel)
	}

	if *useDebug {
		fmt.Println("Sending events to stdout")
		senders = append(senders, NewDebugSender())
	}

	if len(senders) == 0 {
		return fmt.Errorf("No senders specified!")
	}

	os.Remove(*socketPath)
	ln, err := net.Listen("unix", *socketPath)
	if err != nil {
		return err
	}
	fmt.Println("Started listening to socket...")
	conn, err := ln.Accept()
	if err != nil {
		return err
	}

	sender := NewCompositeSender(senders)

	for {
		err := processMessage(conn, sender)

		if err != nil && err != io.EOF {
			fmt.Println(err)
		}
	}

}

func processMessage(conn net.Conn, sender Sender) error {
	message, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		return err
	}

	event := map[string]interface{}{}
	if err := json.Unmarshal(message, &event); err != nil {
		return err
	}

	typed := Event{}
	if err := mapstructure.Decode(event, &typed); err != nil {
		return err
	}

	flat, err := flatten.Flatten(event, "", flatten.DotStyle)
	if err != nil {
		return err
	}

	if err := sender.Send(typed, flat); err != nil {
		return err
	}

	return nil
}

type Sender interface {
	Send(Event, map[string]interface{}) error
}

type Event struct {
	Type  string
	Error string

	Request Request
}

type Request struct {
	ID string
}
