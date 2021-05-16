package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"

	libhoney "github.com/honeycombio/libhoney-go"
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

	err := libhoney.Init(libhoney.Config{
		APIKey:  os.Getenv("HONEYCOMB_API_KEY"),
		Dataset: "vault-observe",
	})
	if err != nil {
		return err
	}

	ln, err := net.Listen("unix", "/tmp/audit.sock")
	if err != nil {
		return err
	}

	conn, err := ln.Accept()
	if err != nil {
		return err
	}

	for {
		message, err := bufio.NewReader(conn).ReadBytes('\n')
		if err != nil {
			return err
		}

		partial := map[string]interface{}{}
		if err := json.Unmarshal(message, &partial); err != nil {
			return err
		}

		ev := libhoney.NewEvent()
		ev.Add(partial)
		if err := ev.Send(); err != nil {
			return err
		}

		// picker := EventPicker{}
		// if err := mapstructure.Decode(partial, &picker); err != nil {
		// 	return err
		// }

		// if picker.Type == "request" {
		// 	entry := audit.AuditRequestEntry{}
		// 	if err := mapstructure.Decode(partial, &entry); err != nil {
		// 		return err
		// 	}

		// }

		// if picker.Type == "response" {
		// 	entry := audit.AuditResponseEntry{}
		// 	if err := mapstructure.Decode(partial, &entry); err != nil {
		// 		fmt.Println(err)
		// 		return 1
		// 	}
		// }
		// // fmt.Println("Got " + picker.Type)

	}

}

type EventPicker struct {
	Type string
}
