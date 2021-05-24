package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"

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

	socketPath := flags.String("socket-path", "observe.sock", "the unix socket path for vault to send audit events to")

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
		senders = append(senders, honey)
	}

	if *useZipkin {
		fmt.Println("Sending events to Zipkin")
		otel, err := NewOtelSender()
		if err != nil {
			return err
		}

		senders = append(senders, otel)
	}

	if *useDebug {
		fmt.Println("Sending events to stdout")
		senders = append(senders, NewDebugSender())
	}

	if len(senders) == 0 {
		return fmt.Errorf("No senders specified!")
	}

	sender := NewCompositeSender(senders)

	handleSignals(sender)

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

func handleSignals(sender Sender) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-signals
		fmt.Printf("Recieved %s, stopping\n", s)
		if err := sender.Shutdown(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		os.Exit(0)
	}()
}
