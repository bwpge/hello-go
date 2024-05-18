package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"strings"
	"time"
)

var Unauthorized = errors.New("unauthorized")

var InvalidCredentials = errors.New("invalid username or password")

type Server struct {
	port     uint16
	listener net.Listener
	clients  map[*net.Conn]struct{}
	db       *sql.DB
}

func NewServer(port uint16) *Server {
	return &Server{
		port:    port,
		clients: make(map[*net.Conn]struct{}),
		db:      DbConnect(),
	}
}

func (s *Server) Run() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", s.port))
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()
	defer s.db.Close()

	s.listener = listener
	fmt.Printf("Server listening on %v\n", s.listener.Addr().String())

	s.runLoop()
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

	user, err := s.auth(buf, conn)
	if err != nil {
		SendPacket(conn, ErrorPacket(err.Error()))
		return
	}

	name := fmt.Sprintf("%v@%v", user, addr)
	conn.SetReadDeadline(time.Time{})
	s.clients[&conn] = struct{}{}

	SendPacket(conn, Packet{Type: SERVER_READY})
	fmt.Printf("Client connected: %v\n", name)

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

		fmt.Printf("%v: %v\n", name, p.String())

		switch p.Type {
		case MESSAGE_DIRECT:
			SendPacket(conn, Packet{Type: SERVER_ACK})
		case MESSAGE_BROADCAST:
			SendPacket(conn, Packet{Type: SERVER_ACK})
			p.Body = fmt.Sprintf("%v: %v", name, p.Body)
			go s.broadcast(p)
		default:
		}
	}

	fmt.Printf("Client disconnected: %v\n", name)
	delete(s.clients, &conn)
}

func (s *Server) auth(buf []byte, conn net.Conn) (string, error) {
	// client should open the connection with an auth packet, so reject if not ready
	conn.SetReadDeadline(time.Now().Add(time.Second))
	_, packet, err := ReadPacket(buf, conn)
	addr := conn.RemoteAddr().String()

	if err != nil || packet.Type != CLIENT_AUTH {
		fmt.Printf("REJECT unauthorized (%v)\n", addr)
		return "", Unauthorized
	}

	if packet.Body == "guest:" {
		user := fmt.Sprintf("guest%v", rand.Uint32N(99999))
		fmt.Printf("ACCEPT guest user `%v` (%v)\n", user, addr)
		return user, nil
	}

	user, pass, _ := strings.Cut(packet.Body, ":")
	if !AuthUser(s.db, user, pass) {
		fmt.Printf("REJECT invalid credentials (%v)\n", addr)
		return "", InvalidCredentials
	}
	fmt.Printf("ACCEPT authenticated user `%v` (%v)\n", user, addr)

	return user, nil
}

func (s *Server) broadcast(p Packet) {
	for client := range s.clients {
		SendPacket(*client, p)
	}
}
