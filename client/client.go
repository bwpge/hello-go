package client

import (
	"bufio"
	"errors"
	"fmt"
	"hello-go/common"
	"io"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/gorilla/websocket"
)

type WsClient struct {
	port uint16
	conn *websocket.Conn
	rx   chan common.Packet
	tx   chan common.Packet
	quit chan struct{}
}

func New(port uint16) *WsClient {
	return &WsClient{
		port: port,
		rx:   make(chan common.Packet),
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
	otp, err := c.authenticate(user, pass)
	if err != nil {
		log.Error(err)
		return
	}
	log.Debugf("received OTP from server: %s", otp)

	conn, _, err := websocket.DefaultDialer.Dial(
		fmt.Sprintf("ws://localhost:%v/ws?otp=%s", c.port, otp),
		map[string][]string{"Origin": {fmt.Sprintf("http://localhost:%d", c.port)}},
	)
	if err != nil {
		log.Fatal(err)
	}

	c.conn = conn
	log.Infof("connected to %v", c.RemoteAddr().String())
	defer c.Close()
	go c.recv()
	go c.repl()
	c.msgLoop()
}

func (c *WsClient) authenticate(user string, pass string) (string, error) {
	log.Debug("requesting OTP from server")

	url := fmt.Sprintf("http://localhost:%d/login", c.port)
	r, _ := http.NewRequest("GET", url, nil)
	r.SetBasicAuth(user, pass)
	client := http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("server responded with `%s`", resp.Status))
	}

	otp, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(otp), nil
}

func (c *WsClient) recv() {
	for {
		p, err := common.ReadPacket(c.conn)
		if err != nil && err != common.ErrDisconnected {
			log.Error(err)
			break
		}
		if p == nil {
			break
		}
		c.rx <- p
	}

	c.quit <- struct{}{}
}

func (c *WsClient) repl() {
	reader := bufio.NewReader(os.Stdin)
	defer func() {
		c.quit <- struct{}{}
	}()

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatal(err)
			}
		}

		input = strings.TrimSpace(input)
		if input == "QUIT" || input == "Q" {
			log.Debugf("exiting REPL")
			break
		}

		c.tx <- &common.RawPacket{
			Type:    "text",
			Payload: []byte(input),
		}
	}
}

func (c *WsClient) msgLoop() {
	for {
		select {
		case <-c.rx:
			// TODO: handle packet
		case p := <-c.tx:
			if err := common.WritePacket(c.conn, p); err != nil {
				log.Error(err)
				return
			}
		case <-c.quit:
			log.Debugf("exiting message loop")
			return
		}
	}
}
