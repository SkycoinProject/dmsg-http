package dmsghttp

import (
	"context"
	"log"
	"net/http"

	"github.com/skycoin/dmsg"
	"github.com/skycoin/dmsg/cipher"
	"github.com/skycoin/dmsg/disc"
)

const (
	DefaultDiscoveryURL = "https://messaging.discovery.skywire.skycoin.net"
)

type DMSGTransport struct {
	Discovery disc.APIClient
	PubKey    cipher.PubKey
	SecKey    cipher.SecKey
}

func (t DMSGTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// ctx := req.Context()

	// init client
	reqClient := dmsg.NewClient(t.PubKey, t.SecKey, t.Discovery)

	// connect to the DMSG server
	if err := reqClient.InitiateServerConnections(context.Background(), 1); err != nil {
		log.Fatalf("Error initiating server connections by initiator: %v", err)
	}

	response := http.Response{}

	return &response, nil
}
