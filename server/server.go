package server

import (
	"encoding/json"
	"fmt"
	"hello-go/common"
	"net/http"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
	"github.com/gorilla/websocket"
)

type PeerMap map[Peer]struct{}

type WsServer struct {
	port     uint16
	db       *common.Database
	peers    PeerMap
	upgrader websocket.Upgrader
	sync.RWMutex
}

func New(port uint16) *WsServer {
	origins := map[string]struct{}{
		fmt.Sprintf("http://localhost:%v", port): {},
		"https://websocketking.com":              {},
	}

	return &WsServer{
		port:  port,
		peers: make(PeerMap),
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
	http.HandleFunc("/", s.index)
	http.HandleFunc("/ws", s.serveWS)
	// TODO: implement REST API

	log.Debugf("server listening on :%v", s.port)
	return http.ListenAndServe(fmt.Sprintf(":%v", s.port), nil)
}

func (s *WsServer) index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "online"})
}

func (s *WsServer) serveWS(w http.ResponseWriter, r *http.Request) {
	log.Debugf("websocket request from: %v", r.RemoteAddr)

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err, *r)
		if conn != nil {
			conn.Close()
		}
		return
	}

	log.Debugf("upgraded to websocket: %v", conn.RemoteAddr())
	go s.handle(Peer{conn: conn})
}

func (s *WsServer) add(p Peer) {
	s.Lock()
	defer s.Unlock()

	s.peers[p] = struct{}{}
}

func (s *WsServer) remove(p Peer) {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.peers[p]; ok {
		log.Debugf("removing client: %v", p.Name())
		p.conn.Close()
		delete(s.peers, p)
	}
}

func (s *WsServer) handle(p Peer) {
	s.add(p)
	defer s.remove(p)

	// TODO: implement read/write channels
	s.readPackets(p)
}

func (s *WsServer) readPackets(p Peer) {
	for {
		ty, data, err := p.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Warnf("unexpected closure: %v", err)
			}
			log.Infof("client disconnected: %v", p.Name())
			break
		}

		log.Infof("recv: addr=%v, type=%v, data=%v", p.Name(), ty, strings.TrimSpace(string(data)))
	}
}
