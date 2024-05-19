package server

import "github.com/gorilla/websocket"

type Peer struct {
	conn *websocket.Conn
}

func (p Peer) Name() string {
	return p.conn.RemoteAddr().String()
}
