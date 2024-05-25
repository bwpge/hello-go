package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/charmbracelet/log"
)

const bearerPrefix = "Bearer "

type obj map[string]any

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			log.Debugf("api request: %v, sender=%v", r.URL, r.RemoteAddr)
			next.ServeHTTP(w, r)
		})
}

func (s *WsServer) bearerAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := getBearerToken(r)
		if !s.db.IsValidToken(token) {
			writeErrorJSON(w, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *WsServer) basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || !s.db.AuthUser(user, pass) {
			writeErrorJSON(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *WsServer) apiCreateToken(w http.ResponseWriter, r *http.Request) {
	user, pass, ok := r.BasicAuth()
	if !ok || !s.db.AuthUser(user, pass) {
		writeErrorJSON(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}
	token, err := s.db.CreateToken(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	writeJSON(w, http.StatusCreated, obj{"token": token})
}

func (s *WsServer) apiCheckToken(w http.ResponseWriter, r *http.Request) {
	token := getBearerToken(r)
	if !s.db.IsValidToken(token) {
		writeErrorJSON(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	writeJSON(w, http.StatusOK, obj{"message": "Token is valid"})
}

func (s *WsServer) apiCreateUser(w http.ResponseWriter, r *http.Request) {
	var c struct {
		User string `json:"user"`
		Pass string `json:"pass"`
	}
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	c.User = strings.TrimSpace(c.User)
	c.Pass = strings.TrimSpace(c.Pass)
	if c.User == "" || c.Pass == "" {
		writeErrorJSON(
			w,
			http.StatusBadRequest,
			"Username and password must not be empty or whitespace",
		)
		return
	}

	if err := s.db.CreateUser(c.User, c.Pass); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "User already exists")
		return
	}

	writeJSON(w, http.StatusCreated, obj{})
}

func (s *WsServer) apiStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, obj{
		"status":  "online",
		"clients": len(s.peers),
	})
}

func (s *WsServer) apiGetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.db.GetUsers()
	if err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, err.Error())
	}

	writeJSON(w, http.StatusOK, obj{"users": users})
}

// It makes no sense to allow users to see this data, but just for demo purposes
func (s *WsServer) apiGetUser(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	if userId == "" {
		writeErrorJSON(w, http.StatusBadRequest, "missing required `userId` resource")
		return
	}

	u, err := s.db.UserInfo(userId)
	if err != nil {
		writeErrorJSON(w, http.StatusInternalServerError, "%v", err.Error())
		return
	}
	if u == nil {
		writeJSON(w, http.StatusOK, obj{})
		return
	}

	writeJSON(w, http.StatusOK, u)
}

type endpoint struct {
	method    string
	route     string
	handler   http.Handler
	protected bool
}

func (s *WsServer) registerApi() {
	prefix := "/api/v1"

	for _, e := range []endpoint{
		{method: "POST", route: "/register", handler: http.HandlerFunc(s.apiCreateUser)},
		{method: "GET", route: "/status", handler: http.HandlerFunc(s.apiStatus), protected: true},
		{method: "POST", route: "/users/token", handler: http.HandlerFunc(s.apiCreateToken)},
		{method: "GET", route: "/users/token", handler: http.HandlerFunc(s.apiCheckToken)},
		{method: "GET", route: "/users", handler: http.HandlerFunc(s.apiGetUsers), protected: true},
		{method: "GET", route: "/users/{userId}", handler: http.HandlerFunc(s.apiGetUser), protected: true},
	} {
		route := prefix + e.route
		if e.method != "" {
			route = fmt.Sprintf("%s %s", e.method, route)
		}
		caveat := ""
		if !e.protected {
			caveat = " (unprotected)"
		}
		log.Debugf("creating api route: `%s`%s", route, caveat)

		handler := e.handler
		if e.protected {
			handler = s.bearerAuth(handler)
		}
		http.Handle(route, logMiddleware(handler))
	}
}

func getBearerToken(r *http.Request) string {
	token := r.Header.Get("Authorization")
	if strings.HasPrefix(token, bearerPrefix) {
		return token[len(bearerPrefix):]
	}

	return ""
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	bytes, err := json.Marshal(data)
	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(bytes)
}

func writeErrorJSON(w http.ResponseWriter, status int, msg ...string) {
	var err string
	if len(msg) == 1 {
		err = msg[0]
	} else if len(msg) > 1 {
		err = fmt.Sprintf(msg[0], msg[1:])
	} else {
		err = "User is not authorized to access resource"
	}

	writeJSON(w, status, obj{
		"error": err,
		"code":  status,
	})
}
