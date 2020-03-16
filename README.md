# dmsg-http

HTTP library for dmsg.
Provides a custom http transport to send requests using dmsg protocol instead of tcp.

In order to run the tests for this project you should run following line

```bash
go get golang.org/x/net
```

## Examples

In order to instantiate the server you can use following code

```golang
// define port where server will listen
serverPort := uint16(8080)

// prepare the server
sPK, sSK := cipher.GenerateKeyPair()
dmsgClient := dmsg.NewClient(sPK, sSK, dmsgD, dmsg.DefaultConfig())
go dmsgClient.Serve()

time.Sleep(time.Second) // wait for dmsg client to be ready

// prepare server route handling
mux := http.NewServeMux()
mux.HandleFunc("/some-route", func(w http.ResponseWriter, _ *http.Request) {
    _, err := w.Write([]byte("Route response goes here"))
    if err != nil {
        panic(err)
    }
})

// run the server
srv := &http.Server{
    Handler: mux,
}

list, err := dmsgClient.Listen(serverPort)
if err != nil {
    panic(err)
}

sErr := make(chan error, 1)
go func() {
    sErr <- srv.Serve(list)
    close(sErr)
}()

```

If you would like to talk to this server following code will suffice

```golang
// prepare the client
cPK, cSK := cipher.GenerateKeyPair()
dmsgClient := dmsg.NewClient(cPK, cSK, dmsgD, dmsg.DefaultConfig())
go dmsgClient.Serve()

time.Sleep(time.Second) // wait for dmsg client to be ready

dmsgTransport := dmsghttp.Transport{
		DmsgClient: dmsgClient,
		RetryCount: 20,
	}

c := &http.Client{
	Transport:     dmsgTransport, 
	Timeout:       clientTimeout,
}

// make request
req, err := http.NewRequest("GET", fmt.Sprintf("dmsg://%v:%d/some-route", sPK.Hex(), testPort), nil)
resp, err := c.Do(req)
respBody, err := ioutil.ReadAll(resp.Body)
fmt.Println(string(respBody))
```
