package server

import (
	"hello-go/common"

	"github.com/charmbracelet/log"
	"github.com/gorilla/websocket"
)

type Peer struct {
	conn *websocket.Conn
	tx   chan common.Packet
}

func NewPeer(conn *websocket.Conn) *Peer {
	return &Peer{
		conn: conn,
		tx:   make(chan common.Packet),
	}
}

func (p *Peer) Name() string {
	return p.conn.RemoteAddr().String()
}

func (p *Peer) recv() {
	for {
		p, err := common.ReadPacket(p.conn)
		if err != nil && err != common.ErrDisconnected {
			log.Error(err)
			break
		}
		if p == nil {
			break
		}
	}
}
