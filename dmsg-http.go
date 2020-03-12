package dmsghttp

import (
	"net/http"
	"time"

	"github.com/SkycoinProject/dmsg/cipher"
)

// DefaultDMSGClient creates http Client using default discovery service
// Default value can be found in dmsghttp.DefaultDiscoveryURL
func DefaultClient(pubKey cipher.PubKey, secKey cipher.SecKey) *http.Client {
	return Client(DefaultDMSGClient(pubKey, secKey))
}

// Client creates http Client using provided discovery service and public / secret keypair
// Returned client is using dmsg transport protocol instead of tcp for establishing connection
func Client(dmsgC *DMSGClient) *http.Client {
	transport := Transport{
		DMSGC:      dmsgC,
		RetryCount: 20,
	}

	return &http.Client{
		Transport: transport,
		Jar:       nil,
		Timeout:   time.Second * 30,
	}
}
