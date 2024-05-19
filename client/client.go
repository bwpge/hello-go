package client

import (
	"bufio"
	"fmt"
	"hello-go/common"
	"net"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/gorilla/websocket"
)

type WsClient struct {
	port uint16
	conn *websocket.Conn
	rx   chan *common.Packet
	tx   chan common.Packet
	quit chan struct{}
}

func New(port uint16) *WsClient {
	return &WsClient{
		port: port,
		rx:   make(chan *common.Packet),
		tx:   make(chan common.Packet),
		quit: make(chan struct{}),
	}
}

func (c *WsClient) Close() {
	if c.conn != nil {
		c.conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		)
		c.conn.Close()
	}
}

func (c *WsClient) RemoteAddr() net.Addr {
	if c.conn != nil {
		return c.conn.RemoteAddr()
	}
	panic("connection is nil")
}

func (c *WsClient) LocalAddr() net.Addr {
	if c.conn != nil {
		return c.conn.LocalAddr()
	}
	panic("connection is nil")
}

func (c *WsClient) Run(user string, pass string) {
	conn, _, err := websocket.DefaultDialer.Dial(
		fmt.Sprintf("ws://localhost:%v/ws", c.port),
		map[string][]string{"Origin": {fmt.Sprintf("http://localhost:%d", c.port)}},
	)
	if err != nil {
		log.Fatal(err)
	}

	c.conn = conn
	fmt.Printf("Connected to %v\n", c.RemoteAddr().String())
	defer c.Close()
	go c.repl()
	c.msgLoop()
}

func (c *WsClient) repl() {
	reader := bufio.NewReader(os.Stdin)
	defer func() {
		c.quit <- struct{}{}
	}()

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		input = strings.TrimSpace(input)
		if input == "QUIT" || input == "Q" {
			fmt.Println("Goodbye!")
			break
		}

		c.tx <- &common.MessagePacket{Data: input}
	}
}

func (c *WsClient) msgLoop() {
	for {
		select {
		case p := <-c.tx:
			if err := c.conn.WriteJSON(p.EncodePacket()); err != nil {
				panic(err)
			}
		case p := <-c.rx:
			fmt.Printf("recv: %v\n", p)
		case <-c.quit:
			log.Debugf("exiting message loop")
			return
		}
	}
}