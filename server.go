package dmsghttp

import (
	"net/http"
	"time"

	"github.com/SkycoinProject/dmsg/cipher"
	"github.com/SkycoinProject/dmsg/disc"
)

// Server holds relevant data for server to run properly
// Data includes Public / Secret key pair that identifies the server.
// There is also port on which server will listen.
// Optional parameter is Discovery, if none is provided default one will be used.
// Default dicovery URL is stored as dmsghttp.DefaultDiscoveryURL
type Server struct {
	PubKey    cipher.PubKey
	SecKey    cipher.SecKey
	Port      uint16
	Discovery disc.APIClient

	hs *http.Server
}

// Serve handles request to dmsg server
// Accepts handler holding routes for the current instance
func (s *Server) Serve(handler http.Handler) error {
	s.hs = &http.Server{Handler: handler}

	client, err := getClient(s.PubKey, s.SecKey)
	if err != nil {
		return err
	}

	// this serve invocation opens connectio to the DMSG Server and registers this Client on the Discovery
	go client.Serve()
	time.Sleep(time.Second) // wait until connection is established

	list, err := client.Listen(s.Port)
	if err != nil {
		return err
	}
	return s.hs.Serve(list)
}

// Close closes active Listeners and Connections by invoking http's Close func
func (s *Server) Close() error {
	return s.hs.Close()
}
