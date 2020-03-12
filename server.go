package dmsghttp

import (
	"net/http"
	"time"
)

// Server holds relevant data for server to run properly
// Data includes parameters to instantiate a dmsgclient and a port on which server will listen.
type Server struct {
	DMSGC *DMSGClient
	Port  uint16
	hs    *http.Server
}

// Serve handles request to dmsg server
// Accepts handler holding routes for the current instance
func (s *Server) Serve(handler http.Handler) error {
	s.hs = &http.Server{Handler: handler}

	client, err := GetClient(s.DMSGC)
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
