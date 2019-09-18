package dmsghttp

import (
	"context"
	"log"
	"net/http"

	"github.com/skycoin/dmsg/cipher"
	"github.com/skycoin/dmsg/disc"

	"github.com/skycoin/dmsg"
)

type Server struct {
	PubKey       cipher.PubKey
	SecKey       cipher.SecKey
	Port         uint16
	DiscoveryURL string
}

func (s Server) Serve(handler http.Handler) (*dmsg.Server, error) {
	hsrv := http.Server{Handler: handler}

	if len(s.DiscoveryURL) == 0 {
		s.DiscoveryURL = DefaultDiscoveryURL
	}
	dc := disc.NewHTTP(s.DiscoveryURL)

	client := dmsg.NewClient(s.PubKey, s.SecKey, dc)
	if err := client.InitiateServerConnections(context.Background(), 1); err != nil {
		log.Fatalf("Error initiating server connections by initiator: %v", err)
	}

	list, _ := client.Listen(s.Port)
	go func() { _ = hsrv.Serve(list) }()
	return nil, nil
}
