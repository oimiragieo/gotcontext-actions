package oidc

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Server is a minimal local OIDC token endpoint for act.
// It does not mint GitHub-trusted tokens; cloud providers will reject them.
type Server struct {
	URL      string
	listener net.Listener
	key      *rsa.PrivateKey
	mu       sync.Mutex
	claims   map[string]any
}

// Start launches an HTTP server that serves request URL + token minting.
func Start(claims map[string]any) (*Server, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	s := &Server{
		listener: ln,
		key:      key,
		claims:   claims,
		URL:      fmt.Sprintf("http://%s", ln.Addr().String()),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/token", s.handleToken)
	mux.HandleFunc("/.well-known/openid-configuration", s.handleDiscovery)
	mux.HandleFunc("/jwks", s.handleJWKS)
	go func() {
		_ = http.Serve(ln, mux)
	}()
	return s, nil
}

// Close stops the mock server.
func (s *Server) Close() error {
	if s == nil || s.listener == nil {
		return nil
	}
	return s.listener.Close()
}

// RequestURL returns ACTIONS_ID_TOKEN_REQUEST_URL value.
func (s *Server) RequestURL() string {
	return s.URL + "/token"
}

func (s *Server) handleDiscovery(w http.ResponseWriter, _ *http.Request) {
	_ = json.NewEncoder(w).Encode(map[string]any{
		"issuer":                                s.URL,
		"jwks_uri":                              s.URL + "/jwks",
		"token_endpoint":                        s.URL + "/token",
		"id_token_signing_alg_values_supported": []string{"RS256"},
	})
}

func (s *Server) handleJWKS(w http.ResponseWriter, _ *http.Request) {
	n := base64.RawURLEncoding.EncodeToString(s.key.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(s.key.E)).Bytes())
	_ = json.NewEncoder(w).Encode(map[string]any{
		"keys": []map[string]string{{
			"kty": "RSA",
			"kid": "act-mock",
			"alg": "RS256",
			"use": "sig",
			"n":   n,
			"e":   e,
		}},
	})
}

func (s *Server) handleToken(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	claims := jwt.MapClaims{
		"iss":   s.URL,
		"aud":   r.URL.Query().Get("audience"),
		"sub":   "repo:local/act:ref:refs/heads/master",
		"iat":   now.Unix(),
		"nbf":   now.Unix(),
		"exp":   now.Add(10 * time.Minute).Unix(),
		"jti":   fmt.Sprintf("act-mock-%d", now.UnixNano()),
		"actor": "nektos/act",
	}
	for k, v := range s.claims {
		claims[k] = v
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "act-mock"
	signed, err := token.SignedString(s.key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"value": signed})
}
