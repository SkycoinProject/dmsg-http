package dmsghttp

import (
	"context"
	"log"
	"net/http"

	"github.com/skycoin/dmsg/cipher"
	"github.com/skycoin/dmsg/disc"
	"github.com/skycoin/skycoin/src/util/logging"

	"github.com/skycoin/dmsg"
)

type Server struct {
	PubKey       cipher.PubKey
	SecKey       cipher.SecKey
	Port         uint16
	DiscoveryURL string

	hs *http.Server
}

func (s *Server) Serve(handler http.Handler) error {
	s.hs = &http.Server{Handler: handler}

	if len(s.DiscoveryURL) == 0 {
		s.DiscoveryURL = DefaultDiscoveryURL
	}
	dc := disc.NewHTTP(s.DiscoveryURL)

	client := dmsg.NewClient(s.PubKey, s.SecKey, dc, dmsg.SetLogger(logging.MustGetLogger("dmsgC_httpS")))
	if err := client.InitiateServerConnections(context.Background(), 1); err != nil {
		log.Fatalf("Error initiating server connections by initiator: %v", err)
	}

	list, err := client.Listen(s.Port)
	if err != nil {
		return err
	}
	return s.hs.Serve(list)
}

func (s *Server) Close() error {
	return s.hs.Close()
}
