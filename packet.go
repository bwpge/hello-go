package main

import "fmt"

type PacketType int

const (
	READY             = iota
	NEED_USERNAME     = iota
	NEED_PASSWORD     = iota
	MESSAGE_DIRECT    = iota
	MESSAGE_BROADCAST = iota
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
