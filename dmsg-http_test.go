package dmsghttp_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
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
	testPort uint16 = 8081
)

var (
	publicKey            = "036c4441b76aad343a9073d12f72d024009feabd8a0ccc92d3f88dacc93aafca65"
	secretKey            = "c483329a1057a578d0f1b93a5bb190a3bebd71c03f19df73dd9852066c22fc7b"
	expectedSmallContent = "small content, small content, small content, small content, small content, "
	expectedLargeContent = "" // populated at the start of main, not to keep 1000 occurrences here
)

func TestClientsMapNotConcurent(t *testing.T) {
	pk, sk := cipher.GenerateKeyPair()
	ppk, ssk := cipher.GenerateKeyPair()
	pppk, sssk := cipher.GenerateKeyPair()
	c, _ := dmsghttp.GetClient(pk, sk)
	ca, _ := dmsghttp.GetClient(ppk, ssk)
	cb, _ := dmsghttp.GetClient(pk, sk)
	cc, _ := dmsghttp.GetClient(pppk, sssk)
	cd, _ := dmsghttp.GetClient(ppk, ssk)
	ce, _ := dmsghttp.GetClient(pppk, sssk)

	fmt.Println("Client 1 and 3 should be equal")
	fmt.Println("PK Client 1 : ", c.EntityCommon.LocalPK())
	fmt.Println("PK Client 3 : ", cb.EntityCommon.LocalPK())
	require.Equal(t, c.EntityCommon.LocalPK(), cb.EntityCommon.LocalPK())
	require.Equal(t, c, cb)

	fmt.Println("Client 2 and 5 should be equal")
	fmt.Println("PK Client 2 : ", ca.EntityCommon.LocalPK())
	fmt.Println("PK Client 5 : ", cd.EntityCommon.LocalPK())
	require.Equal(t, ca.EntityCommon.LocalPK(), cd.EntityCommon.LocalPK())
	require.Equal(t, ca, cd)

	fmt.Println("Client 4 and 6 should be equal")
	fmt.Println("PK Client 4 : ", cc.EntityCommon.LocalPK())
	fmt.Println("PK Client 6 : ", ce.EntityCommon.LocalPK())
	require.Equal(t, cc.EntityCommon.LocalPK(), ce.EntityCommon.LocalPK())
	require.Equal(t, cc, ce)
}

func TestClientsMapConcurent(t *testing.T) {
	pk, sk := cipher.GenerateKeyPair()
	ppk, ssk := cipher.GenerateKeyPair()
	pppk, sssk := cipher.GenerateKeyPair()

	var c, ca, cb, cc, cd, ce *dmsg.Client

	go func() {
		c, _ = dmsghttp.GetClient(pk, sk)
		ca, _ = dmsghttp.GetClient(ppk, ssk)
		cb, _ = dmsghttp.GetClient(pk, sk)
		cc, _ = dmsghttp.GetClient(pppk, sssk)
		cd, _ = dmsghttp.GetClient(ppk, ssk)
		ce, _ = dmsghttp.GetClient(pppk, sssk)
	}()

	go func() {
		c, _ = dmsghttp.GetClient(pk, sk)
		ca, _ = dmsghttp.GetClient(ppk, ssk)
		cb, _ = dmsghttp.GetClient(pk, sk)
		cc, _ = dmsghttp.GetClient(pppk, sssk)
		cd, _ = dmsghttp.GetClient(ppk, ssk)
		ce, _ = dmsghttp.GetClient(pppk, sssk)
	}()

	go func() {
		c, _ = dmsghttp.GetClient(pk, sk)
		ca, _ = dmsghttp.GetClient(ppk, ssk)
		cb, _ = dmsghttp.GetClient(pk, sk)
		cc, _ = dmsghttp.GetClient(pppk, sssk)
		cd, _ = dmsghttp.GetClient(ppk, ssk)
		ce, _ = dmsghttp.GetClient(pppk, sssk)
	}()

	time.Sleep(1 * time.Second)

	fmt.Println("Client 1 and 3 should be equal")
	fmt.Println("PK Client 1 : ", c.EntityCommon.LocalPK())
	fmt.Println("PK Client 3 : ", cb.EntityCommon.LocalPK())
	require.Equal(t, c.EntityCommon.LocalPK(), cb.EntityCommon.LocalPK())
	require.Equal(t, c, cb)

	fmt.Println("Client 2 and 5 should be equal")
	fmt.Println("PK Client 2 : ", ca.EntityCommon.LocalPK())
	fmt.Println("PK Client 5 : ", cd.EntityCommon.LocalPK())
	require.Equal(t, ca.EntityCommon.LocalPK(), cd.EntityCommon.LocalPK())
	require.Equal(t, ca, cd)

	fmt.Println("Client 4 and 6 should be equal")
	fmt.Println("PK Client 4 : ", cc.EntityCommon.LocalPK())
	fmt.Println("PK Client 6 : ", ce.EntityCommon.LocalPK())
	require.Equal(t, cc.EntityCommon.LocalPK(), ce.EntityCommon.LocalPK())
	require.Equal(t, cc, ce)
}

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

func TestEofShowcase(t *testing.T) {
	go createDmsgSrvForEofShowcase()
	time.Sleep(5 * time.Second)
	createDmsgClientForEofShowcase(t)
}

func createDmsgSrv(t *testing.T, dc disc.APIClient) (srv *dmsg.Server, srvErr <-chan error) {
	pk, sk, err := cipher.GenerateDeterministicKeyPair([]byte("s"))
	require.NoError(t, err)
	l, err := nettest.NewLocalListener("tcp")
	require.NoError(t, err)
	srv = dmsg.NewServer(pk, sk, dc)
	require.NoError(t, err)
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(l, "")
		close(errCh)
	}()
	return srv, errCh
}

func createDmsgSrvForEofShowcase() {
	port := uint16(9091) // use any port you like here, make sure it's referenced in the cmd/client/http-client.go
	dmsgD := disc.NewHTTP("http://http://dmsg.discovery.skywire.cc/")
	//dmsgD := disc.NewHTTP("http://localhost:9090")

	//sPK, sSK := cipher.GenerateKeyPair()

	pK := cipher.PubKey{}
	sK := cipher.SecKey{}
	pK.UnmarshalText([]byte(publicKey))
	sK.UnmarshalText([]byte(secretKey))

	// note down public key printed out bellow and use the value in cmd/client/http-client.go 'serverPubKey' value
	fmt.Printf("Starting http server on public key: %s and port: %v", pK.Hex(), port)

	httpS := dmsghttp.Server{
		PubKey:    pK,
		SecKey:    sK,
		Port:      port,
		Discovery: dmsgD,
	}

	// prepare server route handling
	mux := http.NewServeMux()
	mux.HandleFunc("/small", dmsghttp.SmallRequestHandler)
	mux.HandleFunc("/large", dmsghttp.LargeRequestHandler)

	// run the server
	sErr := make(chan error, 1)
	sErr <- httpS.Serve(mux)
	close(sErr)
	fmt.Println(sErr)
}

func createDmsgClientForEofShowcase(t *testing.T) {
	time.Sleep(5 * time.Second)
	var b bytes.Buffer
	countExpectedLargeContent := 1000 // how many times we expect 'large content, ' phrase to appear in the response
	for i := countExpectedLargeContent; i > 0; i-- {
		b.WriteString("large content, ")
	}
	expectedLargeContent = b.String()

	//sPK, sSK := cipher.GenerateKeyPair()
	disc := disc.NewHTTP("http://http://dmsg.discovery.skywire.cc/")
	//disc := disc.NewHTTP("http://localhost:9090")
	pK := cipher.PubKey{}
	sK := cipher.SecKey{}
	pK.UnmarshalText([]byte(publicKey))
	sK.UnmarshalText([]byte(secretKey))
	client := dmsghttp.DMSGClient(disc, pK, sK)

	// this one is returned ok
	resp := dmsghttp.MakeRequest("small", client)
	if resp != expectedSmallContent {
		fmt.Printf("Received small content is not what's expected. Expected %s but received %s\n", expectedSmallContent, resp)
		t.Fail()
		return
	}
	fmt.Println("Small content is ok")

	// this one fails with following message "Error reading data:  http: unexpected EOF reading trailer"
	resp = dmsghttp.MakeRequest("large", client)
	if resp != expectedLargeContent {
		largeContentRegex := regexp.MustCompile("large content,")
		matches := largeContentRegex.FindAllStringIndex(resp, -1)
		countActual := len(matches) // how many times we received expected phrase,
		fmt.Printf("Received large content is not what's expected. Expected %v but received %v 'large content' occurrences\n", countExpectedLargeContent, countActual)
		t.Fail()
		return
	}
	fmt.Print("Large content is ok")
}
