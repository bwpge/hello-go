package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"strings"
	"time"

	"github.com/fatih/color"
)

var Unauthorized = errors.New("unauthorized")

var InvalidCredentials = errors.New("invalid username or password")

var NilConnection = errors.New("nil user connection pointer")

type Server struct {
	port  uint16
	users map[*User]struct{}
	db    *Database
}

func NewServer(port uint16) *Server {
	return &Server{
		port:  port,
		users: make(map[*User]struct{}),
	}
}

func (s *Server) Close() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *Server) Run() error {
	s.db = DbConnect()
	defer s.Close()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", s.port))
	if err != nil {
		log.Fatal(err)
	}

	s.acceptLoop(listener)
	return nil
}

func (s *Server) acceptLoop(listener net.Listener) {
	defer listener.Close()
	color.HiBlack("Server listening on %v\n", listener.Addr().String())

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)

	user, err := s.authUser(buf, conn)
	if err != nil {
		SendPacket(conn, ErrorPacket(err.Error()))
		return
	}

	s.users[user] = struct{}{}
	user.SendPacket(Packet{Type: SERVER_READY})
	color.HiBlack("Client connected: %v\n", user.DisplayName())

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

		fmt.Printf("%v: %v\n", user.DisplayName(), p.String())

		switch p.Type {
		case MESSAGE_DIRECT:
			if user.isGuest {
				user.SendPacket(NotAllowedPacket("guests are not allowed to send direct messages"))
				continue
			}
			user.SendPacket(Packet{Type: SERVER_ACK})
		case MESSAGE_BROADCAST:
			if user.isGuest {
				user.SendPacket(NotAllowedPacket("guests are not allowed to send broadcast messages"))
				continue
			}

			user.SendPacket(Packet{Type: SERVER_ACK})
			p.Body = fmt.Sprintf("%v: %v", user.DisplayName(), p.Body)
			go s.broadcast(p)
		default:
		}
	}

	color.HiBlack("Client disconnected: %v\n", user.DisplayName())
	delete(s.users, user)
}

func (s *Server) authUser(buf []byte, conn net.Conn) (*User, error) {
	// client should open the connection with an auth packet, so reject if not ready
	defer conn.SetReadDeadline(time.Time{})
	conn.SetReadDeadline(time.Now().Add(time.Second))
	_, packet, err := ReadPacket(buf, conn)

	addr := conn.RemoteAddr().String()

	if err != nil || packet.Type != CLIENT_AUTH {
		color.Red("REJECT unauthorized (%v)\n", addr)
		return nil, Unauthorized
	}

	if packet.Body == "guest:" {
		name := fmt.Sprintf("guest%v", 10000+rand.Uint32N(89999))
		color.Green("ACCEPT guest user `%v` (%v)\n", name, addr)
		return NewUser(conn, name), nil
	}

	name, pass, _ := strings.Cut(packet.Body, ":")
	if !s.db.AuthUser(name, pass) {
		color.Red("REJECT invalid credentials (%v)\n", addr)
		return nil, InvalidCredentials
	}
	color.Green("ACCEPT authenticated user `%v` (%v)\n", name, addr)

	return NewUser(conn, name), nil
}

func (s *Server) broadcast(p Packet) {
	for user := range s.users {
		user.SendPacket(p)
	}
}

type User struct {
	conn    net.Conn
	name    string
	isGuest bool
}

func NewUser(conn net.Conn, name string) *User {
	return &User{
		conn:    conn,
		name:    name,
		isGuest: strings.HasPrefix(name, "guest"),
	}
}

func (u *User) DisplayName() string {
	return fmt.Sprintf("%s@%s", u.name, u.conn.RemoteAddr().String())
}

func (u *User) SendPacket(p Packet) error {
	if u.conn == nil {
		return NilConnection
	}

	SendPacket(u.conn, p)
	return nil
}

func (u *User) ReadPacket(buf []byte) (int, Packet, error) {
	if u.conn == nil {
		return 0, Packet{}, NilConnection
	}

	return ReadPacket(buf, u.conn)
}
