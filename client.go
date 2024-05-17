package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

type Client struct {
	port uint16
	conn net.Conn
}

func NewClient(port uint16) *Client {
	return &Client{
		port: port,
	}
}

func (c *Client) Run() {
	conn, err := net.Dial("tcp", fmt.Sprintf(":%v", c.port))
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()
	c.conn = conn
	fmt.Printf("Connected to %v\n", conn.RemoteAddr().String())
	c.repl()
}

func (c *Client) repl() {
	fmt.Println("Waiting for input, use `QUIT` to exit")
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		msg = strings.TrimSpace(msg)
		if msg == "QUIT" {
			fmt.Println("Goodbye!")
			return
		}

		if _, err = fmt.Fprintf(c.conn, "%v\n", msg); err != nil {
			if !ConnClosedErr(err) {
				log.Fatal(err)
			}
			fmt.Println("Server disconnected")
			return
		}
	}
}
