package dmsghttp_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/SkycoinProject/dmsg"
	"github.com/SkycoinProject/dmsg/cipher"
	"github.com/SkycoinProject/dmsg/disc"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/nettest"

	dmsghttp "github.com/SkycoinProject/dmsg-http"
)

const (
	testPort      uint16 = 8081
	clientTimeout        = 30 * time.Second
)

func TestDmsgHTTP(t *testing.T) {
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
	dmsgServerClient := dmsg.NewClient(sPK, sSK, dmsgD, dmsg.DefaultConfig())
	go dmsgServerClient.Serve()

	time.Sleep(time.Second) // wait for dmsg client to be ready

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("Hello World!"))
		if err != nil {
			panic(err)
		}
	})

	srv := &http.Server{
		Handler: mux,
	}

	list, err := dmsgServerClient.Listen(testPort)
	if err != nil {
		panic(err)
	}

	sErr := make(chan error, 1)
	go func() {
		sErr <- srv.Serve(list)
		close(sErr)
	}()
	defer func() {
		require.NoError(t, srv.Close())
		err := <-sErr
		require.Error(t, err)
		require.Equal(t, "http: Server closed", err.Error())
	}()

	// generate keys and initiate client
	cPK, cSK := cipher.GenerateKeyPair()
	dmsgClient := dmsg.NewClient(cPK, cSK, dmsgD, dmsg.DefaultConfig())
	go dmsgServerClient.Serve()

	time.Sleep(time.Second) // wait for dmsg client to be ready

	dmsgTransport := dmsghttp.Transport{
		DmsgClient: dmsgClient,
		RetryCount: 20,
	}
	c := &http.Client{
		Transport: dmsgTransport,
		Timeout:   clientTimeout,
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("dmsg://%v:%d/", sPK.Hex(), testPort), nil)
	require.NoError(t, err)

	resp, err := c.Do(req)
	require.NoError(t, err)

	respB, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "Hello World!", string(respB))
}

func TestDmsgHTTPTargetingSpecificRoute(t *testing.T) {
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
	dmsgServerClient := dmsg.NewClient(sPK, sSK, dmsgD, dmsg.DefaultConfig())
	go dmsgServerClient.Serve()

	time.Sleep(time.Second) // wait for dmsg client to be ready

	mux := http.NewServeMux()
	mux.HandleFunc("/route", func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("Routes Work!"))
		if err != nil {
			panic(err)
		}
	})

	srv := &http.Server{
		Handler: mux,
	}

	list, err := dmsgServerClient.Listen(testPort)
	if err != nil {
		panic(err)
	}

	sErr := make(chan error, 1)
	go func() {
		sErr <- srv.Serve(list)
		close(sErr)
	}()
	defer func() {
		require.NoError(t, srv.Close())
		err := <-sErr
		require.Error(t, err)
		require.Equal(t, "http: Server closed", err.Error())
	}()

	// generate keys and initiate client
	cPK, cSK := cipher.GenerateKeyPair()
	dmsgClient := dmsg.NewClient(cPK, cSK, dmsgD, dmsg.DefaultConfig())
	go dmsgClient.Serve()

	time.Sleep(time.Second) // wait for dmsg client to be ready

	dmsgTransport := dmsghttp.Transport{
		DmsgClient: dmsgClient,
		RetryCount: 20,
	}
	c := &http.Client{
		Transport: dmsgTransport,
		Timeout:   clientTimeout,
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("dmsg://%v:%d/route", sPK.Hex(), testPort), nil)
	require.NoError(t, err)

	resp, err := c.Do(req)
	require.NoError(t, err)

	respB, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "Routes Work!", string(respB))
}

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
	dmsgServerClient := dmsg.NewClient(sPK, sSK, dmsgD, dmsg.DefaultConfig())
	go dmsgServerClient.Serve()

	time.Sleep(time.Second) // wait for dmsg client to be ready

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

	srv := &http.Server{
		Handler: mux,
	}

	list, err := dmsgServerClient.Listen(testPort)
	if err != nil {
		panic(err)
	}

	sErr := make(chan error, 1)
	go func() {
		sErr <- srv.Serve(list)
		close(sErr)
	}()
	defer func() {
		require.NoError(t, srv.Close())
		err := <-sErr
		require.Error(t, err)
		require.Equal(t, "http: Server closed", err.Error())
	}()

	// generate keys and initiate client
	cPK, cSK := cipher.GenerateKeyPair()
	dmsgClient := dmsg.NewClient(cPK, cSK, dmsgD, dmsg.DefaultConfig())
	go dmsgClient.Serve()

	time.Sleep(time.Second) // wait for dmsg client to be ready

	dmsgTransport := dmsghttp.Transport{
		DmsgClient: dmsgClient,
		RetryCount: 20,
	}
	c := &http.Client{
		Transport: dmsgTransport,
		Timeout:   clientTimeout,
	}
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
	srv = dmsg.NewServer(pk, sk, dc, 10)
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(l, "")
		close(errCh)
	}()
	return srv, errCh
}
