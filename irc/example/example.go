package main

import (
	"log"
	"net"

	ircinspector "github.com/go-i2p/go-connfilter/irc"
)

func main() {
	listener, err := net.Listen("tcp", ":6667")
	if err != nil {
		log.Fatal(err)
	}

	inspector := ircinspector.New(listener, ircinspector.Config{
		OnMessage: func(msg *ircinspector.Message) error {
			log.Printf("Received message: %s", msg.Raw)
			return nil
		},
		OnNumeric: func(numeric int, msg *ircinspector.Message) error {
			log.Printf("Received numeric response: %d", numeric)
			return nil
		},
	})

	inspector.AddFilter(ircinspector.Filter{
		Command: "PRIVMSG",
		Channel: "#mychannel",
		Callback: func(msg *ircinspector.Message) error {
			msg.Trailing = "[modified] " + msg.Trailing
			return nil
		},
	})

	for {
		conn, err := inspector.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go ircinspector.HandleConnection(conn)
	}
}
