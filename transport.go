package dmsghttp

import (
	"bufio"
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/SkycoinProject/dmsg"
	"github.com/SkycoinProject/dmsg/cipher"
	"github.com/SkycoinProject/dmsg/disc"
)

// Defaults for dmsg configuration, such as discovery URL
const (
	DefaultDiscoveryURL = "http://dmsg.discovery.skywire.skycoin.com"
)

// DMSGTransport holds information about client who is initiating communication.
type DMSGTransport struct {
	Discovery  disc.APIClient
	PubKey     cipher.PubKey
	SecKey     cipher.SecKey
	RetryCount uint8

	dmsgC      *dmsg.Client // DMSG Client singleton
	clientInit sync.Once    // have only one client init per DMSGTransport instance
}

// RoundTrip implements golang's http package support for alternative transport protocols.
// In this case DMSG is used instead of TCP to initiate the communication with the server.
func (t DMSGTransport) RoundTrip(req *http.Request) (*http.Response, error) {

	// connect to the DMSG server
	if err := t.dmsgC.InitiateServerConnections(context.Background(), 1); err != nil {
		log.Fatalf("Error initiating server connections by initiator: %v", err)
	}

	// process remote pub key and port from dmsg-addr request header
	addrSplit := strings.Split(req.Host, ":")
	if len(addrSplit) != 2 {
		return nil, errors.New("Invalid server Pub Key or Port")
	}
	var pk cipher.PubKey
	if err := pk.Set(addrSplit[0]); err != nil {
		return nil, err
	}
	rPort, _ := strconv.Atoi(addrSplit[1])
	port := uint16(rPort)

	var (
		transport    *dmsg.Transport
		transportErr error
	)
	for i := uint8(0); i < t.RetryCount; i++ {
		transport, transportErr = t.dmsgC.Dial(context.Background(), pk, port)
		if transportErr != nil {
			log.Println("Transport was not established, retrying...")
			// Adding this to make sure we have enough time for delegate servers to become available
			time.Sleep(200 * time.Millisecond)
			continue
		}
		transportErr = nil
		break
	}
	if transportErr != nil {
		return nil, transportErr
	}
	defer transport.Close()

	if err := req.Write(transport); err != nil {
		return nil, err
	}

	return http.ReadResponse(bufio.NewReader(transport), req)
}
