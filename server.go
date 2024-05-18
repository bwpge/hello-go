package main

import (
	"fmt"
	"log"
	"net"
)

type Server struct {
	port     uint16
	listener net.Listener
	clients  map[*net.Conn]struct{}
}

func NewServer(port uint16) *Server {
	return &Server{
		port:    port,
		clients: make(map[*net.Conn]struct{}),
	}
}

func (s *Server) Run() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", s.port))
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()
	s.listener = listener
	fmt.Printf("Server listening on: %v\n", s.listener.Addr().String())

	s.runLoop()
	fmt.Println("Server disconnected")

	return nil
}

func (s *Server) runLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	addr := conn.RemoteAddr().String()
	s.clients[&conn] = struct{}{}

	SendPacket(conn, Packet{Type: SERVER_READY})
	fmt.Printf("Client connected: %v\n", addr)

	for {
		n, p, err := ReadPacket(buf, conn)
		if err != nil {
			if !ConnClosedErr(err) {
				panic(err)
			}
			break
		}
		if n == 0 {
			break
		}

		fmt.Printf("%v: %v\n", addr, p.String())

		switch p.Type {
		case MESSAGE_DIRECT:
			SendPacket(conn, Packet{Type: SERVER_ACK})
		case MESSAGE_BROADCAST:
			SendPacket(conn, Packet{Type: SERVER_ACK})
			go s.broadcast(p)
		default:
			SendPacket(conn, Packet{
				Type: ERROR,
				Body: "invalid packet type",
			})
		}
	}

	fmt.Printf("Client disconnected: %v\n", addr)
	delete(s.clients, &conn)
}

func (s *Server) broadcast(p Packet) {
	for client := range s.clients {
		SendPacket(*client, p)
	}
}
