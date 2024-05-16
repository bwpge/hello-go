package main

import (
	"fmt"
	"log"
	"net"
	"strings"
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
			break
		}

		msg := strings.TrimSpace(string(buf[:n]))
		fmt.Printf("%v: %v\n", addr, msg)
	}

	fmt.Printf("Client disconnected: %v\n", addr)
}
