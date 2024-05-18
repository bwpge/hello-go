package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
)

type PacketType int

const (
	_                 = iota
	CLIENT_AUTH       = iota
	MESSAGE_BROADCAST = iota
	MESSAGE_DIRECT    = iota
	SERVER_ACK        = iota
	SERVER_READY      = iota
	ERROR             = iota
)

type Packet struct {
	Type PacketType `json:"type"`
	Body string     `json:"body"`
}

func Message(body string) Packet {
	return Packet{
		Type: MESSAGE_DIRECT,
		Body: body,
	}
}

func (p *Packet) String() string {
	return fmt.Sprintf("Packet(type=%v, body='%v')", p.Type, p.Body)
}

func (p *Packet) Error() string {
	if p.Type != ERROR {
		panic("packet is not an error")
	}

	return p.Body
}

func ReadPacket(buf []byte, r io.Reader) (int, Packet, error) {
	n, err := r.Read(buf)
	if err != nil {
	}
	if n == 0 {
		return n, Packet{}, nil
	}

	packet := Packet{}
	if err = json.Unmarshal(buf[:n], &packet); err != nil {
		log.Fatal(err)
	}

	return n, packet, nil
}

func SendPacket(w io.Writer, a any) {
	data, err := json.Marshal(a)
	if err != nil {
		panic(err)
	}

	if _, err = w.Write(data); err != nil {
		panic(err)
	}
}

func ErrorPacket(body string) Packet {
	return Packet{
		Type: ERROR,
		Body: body,
	}
}
