package dmsghttp

import (
	"net/http"
	"time"

	"github.com/SkycoinProject/dmsg"
)

const (
	clientTimeout = 30 * time.Second
)

// Client creates a http client
// Returned client is using dmsg transport protocol instead of tcp for establishing connection
func Client(dmsgC *dmsg.Client) *http.Client {
	transport := Transport{
		DmsgClient: dmsgC,
		RetryCount: 20,
	}

	return &http.Client{
		Transport: transport,
		Jar:       nil,
		Timeout:   clientTimeout,
	}
}
