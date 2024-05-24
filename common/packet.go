package common

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/gorilla/websocket"
	"github.com/kelindar/binary"
)

var ErrDisconnected = errors.New("disconnected")

type Packet interface {
	EncodePacket() []byte
}

type RawPacket struct {
	Type    string `json:"type"`
	Payload []byte `json:"payload"`
}

func (p *RawPacket) EncodePacket() []byte {
	data, err := binary.Marshal(*p)
	if err != nil {
		log.Fatal(err)
	}

	return data
}

func (p *RawPacket) String() string {
	t := reflect.TypeOf(*p)
	v := reflect.ValueOf(*p)
	s := t.String() + "{"
	fields := []string{}

	for i := 0; i < v.NumField(); i++ {
		fields = append(fields, fmt.Sprintf("%v: %v", t.Field(i).Name, v.Field(i)))
	}

	s += strings.Join(fields, ", ")
	s += "}"

	return s
}

type PacketReadWriter interface {
	ReadMessage() (int, []byte, error)
	WriteMessage(int, []byte) error
	RemoteAddr() net.Addr
}

func ReadPacket(conn PacketReadWriter) (*RawPacket, error) {
	ty, data, err := conn.ReadMessage()
	if err != nil {
		if websocket.IsUnexpectedCloseError(
			err,
			websocket.CloseNormalClosure,
			websocket.CloseGoingAway,
		) {
			log.Warnf("unexpected closure: %v", err)
		}

		log.Infof("disconnected from %v", conn.RemoteAddr().String())
		return nil, ErrDisconnected
	}

	var packet RawPacket
	if err = binary.Unmarshal(data, &packet); err != nil {
		return nil, err
	}

	log.Infof("recv: addr=%v, type=%v, data=%v", conn.RemoteAddr().String(), ty, packet.String())
	return &packet, nil
}

func WritePacket(conn PacketReadWriter, p Packet) error {
	if err := conn.WriteMessage(websocket.BinaryMessage, p.EncodePacket()); err != nil {
		return err
	}

	return nil
}
