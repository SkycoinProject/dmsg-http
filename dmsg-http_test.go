package dmsghttp_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/SkycoinProject/dmsg"
	"github.com/SkycoinProject/dmsg/cipher"
	"github.com/SkycoinProject/dmsg/disc"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/nettest"

	dmsghttp "github.com/SkycoinProject/dmsg-http"
)

const (
	testPort uint16 = 8081
)

func TestDMSGClient(t *testing.T) {
	dmsgD := disc.NewMock()
	dmsgS, dmsgSErr := createDmsgSrv(t, dmsgD)
	defer func() {
		require.NoError(t, dmsgS.Close())
		for err := range dmsgSErr {
			require.NoError(t, err)
		}
	}()

	// generate keys and create server
	sPK, sSK := cipher.GenerateKeyPair()
	httpS := dmsghttp.Server{PubKey: sPK, SecKey: sSK, Port: testPort, Discovery: dmsgD}

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
	c := dmsghttp.DMSGClient(dmsgD, cPK, cSK)

	req, err := http.NewRequest("GET", fmt.Sprintf("dmsg://%v:%d/", sPK.Hex(), testPort), nil)
	require.NoError(t, err)

	resp, err := c.Do(req)
	require.NoError(t, err)

	respB, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "Hello World!", string(respB))
}

func TestDMSGClientTargetingSpecificRoute(t *testing.T) {
	dmsgD := disc.NewMock()
	dmsgS, dmsgSErr := createDmsgSrv(t, dmsgD)
	defer func() {
		require.NoError(t, dmsgS.Close())
		for err := range dmsgSErr {
			require.NoError(t, err)
		}
	}()

	// generate keys and create server
	sPK, sSK := cipher.GenerateKeyPair()
	httpS := dmsghttp.Server{PubKey: sPK, SecKey: sSK, Port: testPort, Discovery: dmsgD}

	mux := http.NewServeMux()
	mux.HandleFunc("/route", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("Routes Work!"))
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
	c := dmsghttp.DMSGClient(dmsgD, cPK, cSK)

	req, err := http.NewRequest("GET", fmt.Sprintf("dmsg://%v:%d/route", sPK.Hex(), testPort), nil)
	require.NoError(t, err)

	resp, err := c.Do(req)
	require.NoError(t, err)

	respB, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "Routes Work!", string(respB))
}

//FIXME - test was failing before imports fix - err: remote has no DelegatedServers
func TestDMSGClientWithMultipleRoutes(t *testing.T) {
	dmsgD := disc.NewMock()
	dmsgS, dmsgSErr := createDmsgSrv(t, dmsgD)
	defer func() {
		require.NoError(t, dmsgS.Close())
		for err := range dmsgSErr {
			require.NoError(t, err)
		}
	}()

	// generate keys and create server
	sPK, sSK := cipher.GenerateKeyPair()
	httpS := dmsghttp.Server{PubKey: sPK, SecKey: sSK, Port: testPort, Discovery: dmsgD}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("Hello World!"))
		if err != nil {
			panic(err)
		}
	})
	mux.HandleFunc("/route1", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("Routes Work!"))
		if err != nil {
			panic(err)
		}
	})
	mux.HandleFunc("/route2", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("Routes really do Work!"))
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
	c := dmsghttp.DMSGClient(dmsgD, cPK, cSK)

	// check root route
	req, err := http.NewRequest("GET", fmt.Sprintf("dmsg://%v:%d/", sPK.Hex(), testPort), nil)
	require.NoError(t, err)

	resp, err := c.Do(req)
	require.NoError(t, err)

	respB, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "Hello World!", string(respB))

	// check route1
	req, err = http.NewRequest("GET", fmt.Sprintf("dmsg://%v:%d/route1", sPK.Hex(), testPort), nil)
	require.NoError(t, err)

	resp, err = c.Do(req)
	require.NoError(t, err)

	respB, err = ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "Routes Work!", string(respB))

	// check route2
	req, err = http.NewRequest("GET", fmt.Sprintf("dmsg://%v:%d/route2", sPK.Hex(), testPort), nil)
	require.NoError(t, err)

	resp, err = c.Do(req)
	require.NoError(t, err)

	respB, err = ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "Routes really do Work!", string(respB))
}

func createDmsgSrv(t *testing.T, dc disc.APIClient) (srv *dmsg.Server, srvErr <-chan error) {
	pk, sk, err := cipher.GenerateDeterministicKeyPair([]byte("s"))
	require.NoError(t, err)
	l, err := nettest.NewLocalListener("tcp")
	require.NoError(t, err)
	srv, err = dmsg.NewServer(pk, sk, "", l, dc)
	require.NoError(t, err)
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve()
		close(errCh)
	}()
	return srv, errCh
}
