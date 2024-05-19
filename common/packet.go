package common

import (
	"encoding/json"
)

type PacketType int

const (
	_                            = iota
	PacketTypeMessage PacketType = iota
)

type RawPacket struct {
	Type    PacketType      `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type Packet interface {
	EncodePacket() RawPacket
}

type MessagePacket struct {
	Data string `json:"data"`
}

func (m *MessagePacket) EncodePacket() RawPacket {
	payload, err := json.Marshal(*m)
	if err != nil {
		panic(err)
	}

	return RawPacket{
		Type:    PacketTypeMessage,
		Payload: payload,
	}
}
