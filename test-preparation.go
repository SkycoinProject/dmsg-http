package dmsghttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	serverPubKey = "036c4441b76aad343a9073d12f72d024009feabd8a0ccc92d3f88dacc93aafca65" // write down value printed out by server/http-server.go
	serverPort   = "9091"
)

/*********************Server functions*******************/

//SmallRequestHandler generates string with 5 occurrecnes of 'small content, ' phrase
func SmallRequestHandler(w http.ResponseWriter, r *http.Request) {
	var b bytes.Buffer
	for i := 5; i > 0; i-- {
		b.WriteString("small content, ")
	}

	requestHandler(b.String(), w, r)
}

//LargeRequestHandler generates string with 1000 occurrecnes of 'large content, ' phrase
func LargeRequestHandler(w http.ResponseWriter, r *http.Request) {
	var b bytes.Buffer
	for i := 1000; i > 0; i-- {
		b.WriteString("large content, ")
	}

	requestHandler(b.String(), w, r)
}

// writes the response and header
func requestHandler(content string, w http.ResponseWriter, r *http.Request) {
	data := ExampleResponse{
		Content:   content,
		Timestamp: time.Now(),
	}

	fmt.Println("Returning content ", content)
	resp, err := json.Marshal(data)
	fmt.Println("Data object bytes: ", resp)
	if err != nil {
		fmt.Printf("Marshal data error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(resp)
	if err != nil {
		fmt.Printf("Writing response failed due to error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

/******************Client functions************************/

//MakeRequest makes request to the server.
func MakeRequest(path string, client *http.Client) string {
	url := "dmsg://" + serverPubKey + ":" + serverPort + "/" + path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("error creating request for path: ", path)
		return ""
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("requestData request failed due to error: ", err)
		return ""
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			panic(err)
		}
	}()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading data: %v", err)
		return ""
	}

	var data ExampleResponse
	if dataObjectErr := json.Unmarshal(bytes, &data); dataObjectErr != nil {
		fmt.Println("error unmarshaling received data object for path: ", path)
		return ""
	}

	fmt.Printf("Received content %s generated at %v\n", data.Content, data.Timestamp)
	return data.Content
}

//ExampleResponse is model for response message.
type ExampleResponse struct {
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}
