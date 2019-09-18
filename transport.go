package dmsghttp

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

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

	// process remote pub key and port from dmsg-addr request header
	addrSplit := strings.Split(req.Header.Get("dmsg-addr"), ":")
	if len(addrSplit) != 2 {
		return nil, errors.New("Invalid server Pub Key or Port")
	}
	pubKey := cipher.PubKey{}
	pubKey.Set(addrSplit[0])
	rPort, _ := strconv.Atoi(addrSplit[1])
	port := uint16(rPort)

	iTp, err := reqClient.Dial(context.Background(), pubKey, port)
	if err != nil {
		log.Fatalf("Error dialing responder: %v", err)
	}
	defer iTp.Close()

	iTp.Write([]byte("Hello dmsg")) //TODO serialize request body here

	//TODO populate response with response from the server.
	// but first listener on client side is needed, or some alternative for handling response
	// l, _ := reqClient.Listen(port)
	response := http.Response{}

	return &response, nil
}
