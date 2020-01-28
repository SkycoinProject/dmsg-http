package dmsghttp

import (
	"bufio"
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/SkycoinProject/dmsg"
	"github.com/SkycoinProject/dmsg/cipher"
	"github.com/SkycoinProject/dmsg/disc"
)

// Defaults for dmsg configuration, such as discovery URL
const (
	DefaultDiscoveryURL = "http://dmsg.discovery.skywire.cc"
)

// DMSGTransport holds information about client who is initiating communication.
type DMSGTransport struct {
	Discovery  disc.APIClient
	PubKey     cipher.PubKey
	SecKey     cipher.SecKey
	RetryCount uint8
}

// RoundTrip implements golang's http package support for alternative transport protocols.
// In this case DMSG is used instead of TCP to initiate the communication with the server.
func (t DMSGTransport) RoundTrip(req *http.Request) (*http.Response, error) {
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

	serverAddress := dmsg.Addr{PK: pk, Port: port}
	dmsgC, err := getClient(t.PubKey, t.SecKey)
	if err != nil {
		return nil, err
	}

	var (
		stream    *dmsg.Stream
		streamErr error
	)
	for i := uint8(0); i < t.RetryCount; i++ {
		stream, streamErr = dmsgC.DialStream(context.Background(), serverAddress)
		if streamErr != nil {
			log.Printf("Error dialing responder: %s. retrying...", streamErr)
			// Adding this to make sure we have enough time for delegate servers to become available
			time.Sleep(200 * time.Millisecond)
			continue
		}
		streamErr = nil
		break
	}
	if streamErr != nil {
		return nil, streamErr
	}
	defer stream.Close()

	if err := req.Write(stream); err != nil {
		return nil, err
	}

	return http.ReadResponse(bufio.NewReader(stream), req)
}
