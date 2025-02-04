package main

import (
	"fmt"
	"log"
	"net"

	ircinspector "github.com/go-i2p/go-connfilter/irc"
)

func main() {
	// Listen on local port
	listener, err := net.Listen("tcp", ":6667")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	// Create inspector with message logging
	inspector := ircinspector.New(listener, ircinspector.Config{
		OnMessage: func(msg *ircinspector.Message) error {
			log.Printf("Message: %s", msg.Raw)
			return nil
		},
	})

	// Block NICK changes
	inspector.AddFilter(ircinspector.Filter{
		Command: "NICK",
		Callback: func(msg *ircinspector.Message) error {
			return fmt.Errorf("NICK changes not allowed")
		},
	})

	// Modify PRIVMSGs
	inspector.AddFilter(ircinspector.Filter{
		Command: "PRIVMSG",
		Callback: func(msg *ircinspector.Message) error {
			msg.Trailing = "[filtered] " + msg.Trailing
			return nil
		},
	})

	log.Printf("IRC proxy listening on %s", listener.Addr())

	for {
		conn, err := inspector.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(clientConn net.Conn) {
	defer clientConn.Close()

	// Connect to upstream IRC server
	serverConn, err := net.Dial("tcp", "irc.libera.chat:6667")
	if err != nil {
		log.Printf("Server connection error: %v", err)
		return
	}
	defer serverConn.Close()

	// Forward data between connections
	go func() {
		buffer := make([]byte, 4096)
		for {
			n, err := clientConn.Read(buffer)
			if err != nil {
				return
			}
			serverConn.Write(buffer[:n])
		}
	}()

	buffer := make([]byte, 4096)
	for {
		n, err := serverConn.Read(buffer)
		if err != nil {
			return
		}
		clientConn.Write(buffer[:n])
	}
}
