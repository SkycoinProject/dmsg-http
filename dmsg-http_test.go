package dmsghttp_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	dmsghttp "github.com/skycoin/dmsg-http"
	"github.com/skycoin/dmsg/cipher"
)

// import httpdmsg

const (
	testPort uint16 = 8081

	testDC = "http://localhost:9090"
)

func TestDMSGClient(t *testing.T) {
	// generate keys and create server
	serverPK, serverSK := cipher.GenerateKeyPair()
	server := dmsghttp.Server{serverPK, serverSK, uint16(testPort), testDC}

	mux := http.NewServeMux()
	th := &timeHandler{format: time.RFC1123}
	mux.Handle("/", th)
	_, err := server.Serve(mux)

	if err != nil {
		fmt.Printf("Error is %v", err)
	}

	// generate keys and initiate client
	clientPK, clientSK := cipher.GenerateKeyPair()
	c := dmsghttp.DMSGClient(testDC, clientPK, clientSK)

	req, _ := http.NewRequest("GET", "locahost/", nil)
	req.Header.Add("dmsg-addr", fmt.Sprintf("%v:%d", serverPK.Hex(), testPort))
	c.Do(req)
}

func TestDMSGClientTargetingSpecificRoute(t *testing.T) {
	//TODO implement this
}

type timeHandler struct {
	format string
}

func (th *timeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tm := time.Now().Format(th.format)
	w.Write([]byte("The time is: " + tm))
}
