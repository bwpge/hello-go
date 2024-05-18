package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/fatih/color"
)

var BroadcastColor = color.New(color.FgMagenta).Add(color.Bold)

type Client struct {
	port uint16
	conn net.Conn
	rx   chan Packet
	tx   chan Packet
	quit chan struct{}
}

func NewClient(port uint16) *Client {
	return &Client{
		port: port,
		rx:   make(chan Packet),
		tx:   make(chan Packet),
		quit: make(chan struct{}),
	}
}

func (c *Client) Run(user string, pass string) {
	conn, err := net.Dial("tcp", fmt.Sprintf(":%v", c.port))
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()
	c.conn = conn
	color.HiBlack("Connected to %v\n", conn.RemoteAddr().String())
	SendPacket(conn, Packet{
		Type: CLIENT_AUTH,
		Body: fmt.Sprintf("%v:%v", user, pass),
	})

	go c.recv()
	go c.repl()
	c.msgLoop()
}

func (c *Client) recv() {
	buf := make([]byte, 1024)

	for {
		n, p, err := ReadPacket(buf, c.conn)
		if err != nil {
			if !ConnClosedErr(err) {
				panic(err)
			}
			break
		}
		if n == 0 {
			break
		}
		c.rx <- p
	}
}

func (c *Client) repl() {
	reader := bufio.NewReader(os.Stdin)

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		input = strings.TrimSpace(input)
		if input == "QUIT" {
			fmt.Println("Goodbye!")
			c.quit <- struct{}{}
			break
		}

		var ty PacketType
		if strings.HasPrefix(input, "!") {
			ty = MESSAGE_BROADCAST
			input = input[1:]
		} else {
			ty = MESSAGE_DIRECT
		}

		c.tx <- Packet{Type: ty, Body: input}
	}
}

func (c *Client) msgLoop() {
	for {
		select {
		case p := <-c.tx:
			data, err := json.Marshal(p)
			if err != nil {
				panic(err)
			}
			if _, err := c.conn.Write(data); err != nil {
				if !ConnClosedErr(err) {
					panic(err)
				}
				fmt.Println("Server disconnected")
				return
			}
		case p := <-c.rx:
			switch p.Type {
			case SERVER_ACK:
			case SERVER_READY:
				color.Green("SERVER READY")
			case MESSAGE_BROADCAST:
				BroadcastColor.Printf("%v\n", p.Body)
			case ERROR:
				color.Red("ERROR: %v\n", p.Error())
				return
			default:
				color.Cyan("%v\n", p.String())
			}
		case <-c.quit:
			return
		}
	}
}
