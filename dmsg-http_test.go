package dmsghttp_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/skycoin/dmsg/disc"

	"github.com/skycoin/dmsg"
	dmsghttp "github.com/skycoin/dmsg-http"
	"github.com/skycoin/dmsg/cipher"
)

// import httpdmsg

const (
	testPort uint16 = 1563

	testDC = "http://localhost:9090"
)

func TestDMSGClient(t *testing.T) {
	// generate keys and create server
	serverPK, serverSK := cipher.GenerateKeyPair()
	dc := disc.NewHTTP(testDC)
	server := dmsg.NewClient(serverPK, serverSK, dc)
	// connect to the DMSG server
	if err := server.InitiateServerConnections(context.Background(), 1); err != nil {
		log.Fatalf("Error initiating server connections by server: %v", err)
	}
	// bind to port and start listening for incoming messages
	sListener, err := server.Listen(testPort)
	if err != nil {
		log.Fatalf("Error listening by server on port %d: %v", testPort, err)
	}

	// generate keys and initiate client
	clientPK, clientSK := cipher.GenerateKeyPair()
	c := dmsghttp.DMSGClient(testDC, clientPK, clientSK)

	c.Get(fmt.Sprintf("%v:%d", serverPK, testPort))
}

func TestDMSGClientTargetingSpecificRoute(t *testing.T) {
	//TODO implement this
}
