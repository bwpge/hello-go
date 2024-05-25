package server

import (
	"fmt"
	"hello-go/common"
	"net/http"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gorilla/websocket"
)

type (
	PeerMap map[*Peer]struct{}
	OtpMap  map[string]*Otp
)

type WsServer struct {
	port     uint16
	db       *common.Database
	peers    PeerMap
	otps     OtpMap
	upgrader websocket.Upgrader
	sync.RWMutex
}

func New(port uint16) *WsServer {
	origins := map[string]struct{}{
		fmt.Sprintf("http://localhost:%v", port): {},
		// used for gui testing
		"https://websocketking.com": {},
	}

	return &WsServer{
		port:  port,
		peers: make(PeerMap),
		otps:  make(OtpMap),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  2048,
			WriteBufferSize: 2048,
			CheckOrigin: func(r *http.Request) bool {
				o := r.Header.Get("Origin")
				_, ok := origins[o]
				return ok
			},
		},
	}
}

func (s *WsServer) Run() error {
	s.db = common.DbConnect()
	defer s.db.Close()

	http.HandleFunc("GET /login", s.authOTP)
	http.HandleFunc("/ws", s.serveWS)
	s.registerApi()

	// DEBUG: testing client recv
	go func() {
		packet := &common.RawPacket{Type: "heartbeat", Payload: []byte{}}

		for {
			time.Sleep(time.Second * 5)
			s.RLock()
			for p := range s.peers {
				common.WritePacket(p.conn, packet)
			}
			s.RUnlock()
		}
	}()

	// watch for expired otps
	go func() {
		for {
			time.Sleep(time.Second * 5)
			for k, otp := range s.otps {
				if otp.IsExpired() {
					log.Debugf("removing expired OTP %v", k)
					delete(s.otps, k)
				}
			}
		}
	}()

	log.Debugf("server listening on :%v", s.port)
	return http.ListenAndServe(fmt.Sprintf(":%v", s.port), nil)
}

func (s *WsServer) serveWS(w http.ResponseWriter, r *http.Request) {
	log.Debugf("websocket request from: %v", r.RemoteAddr)

	key := r.URL.Query().Get("otp")
	if key == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	otp, ok := s.otps[key]
	if !ok || !otp.Validate(key) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	delete(s.otps, key)

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err, *r)
		if conn != nil {
			conn.Close()
		}
		return
	}

	log.Debugf("upgraded to websocket: %v", conn.RemoteAddr())
	go s.handle(NewPeer(conn))
}

func (s *WsServer) authOTP(w http.ResponseWriter, r *http.Request) {
	user, pass, ok := r.BasicAuth()
	if !ok || !s.db.AuthUser(user, pass) {
		log.Warnf("REJECT unauthorized user `%s` from %v", user, r.RemoteAddr)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	log.Infof("ACCEPT authorized user `%s` from %v", user, r.RemoteAddr)
	otp := NewOtp()
	log.Debugf("creating OTP for %v: %v", r.RemoteAddr, otp.value)
	s.otps[otp.value] = otp

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(otp.value))
}

func (s *WsServer) add(p *Peer) {
	s.Lock()
	defer s.Unlock()

	s.peers[p] = struct{}{}
}

func (s *WsServer) remove(p *Peer) {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.peers[p]; ok {
		log.Debugf("removing client: %v", p.Name())
		p.conn.Close()
		delete(s.peers, p)
	}
}

func (s *WsServer) handle(p *Peer) {
	s.add(p)
	defer s.remove(p)

	p.recv()
}
