package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
)

type Server struct {
	port     uint16
	listener net.Listener
}

func NewServer(port uint16) *Server {
	return &Server{
		port: port,
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

	addr := conn.RemoteAddr().String()
	fmt.Printf("Client connected: %v\n", addr)
	buf := make([]byte, 1024)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			if !ConnClosedErr(err) {
				log.Fatal(err)
			}
			break
		}
		if n == 0 {
			break
		}

		packet := Packet{}
		if err = json.Unmarshal(buf[:n], &packet); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%v: %v\n", addr, packet.String())
	}

	fmt.Printf("Client disconnected: %v\n", addr)
}
