package dmsghttp_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/skycoin/dmsg/cipher"
	"github.com/stretchr/testify/require"

	dmsghttp "github.com/skycoin/dmsg-http"
)

// import httpdmsg

const (
	testPort uint16 = 8081

	testDC = "http://localhost:9090"
)

func TestDMSGClient(t *testing.T) {
	// generate keys and create server
	sPK, sSK := cipher.GenerateKeyPair()
	httpS := dmsghttp.Server{PubKey: sPK, SecKey: sSK, Port: testPort, DiscoveryURL: testDC}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("Hello World!"))
		if err != nil {
			panic(err)
		}
	})

	sErr := make(chan error, 1)
	go func() {
		sErr <- httpS.Serve(mux)
		close(sErr)
	}()
	defer func() {
		require.NoError(t, httpS.Close())
		err := <-sErr
		require.Error(t, err)
		require.Equal(t, "http: Server closed", err.Error())
	}()

	// generate keys and initiate client
	cPK, cSK := cipher.GenerateKeyPair()
	c := dmsghttp.DMSGClient(testDC, cPK, cSK)

	req, err := http.NewRequest("GET", fmt.Sprintf("dmsg://%v:%d/", sPK.Hex(), testPort), nil)
	require.NoError(t, err)

	resp, err := c.Do(req)
	require.NoError(t, err)

	respB, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "Hello World!", string(respB))
}

func TestDMSGClientTargetingSpecificRoute(t *testing.T) {
	//TODO implement this
}
