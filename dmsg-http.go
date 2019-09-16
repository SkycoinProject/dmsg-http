package dmsghttp

import (
	"net/http"
	"time"

	"github.com/skycoin/dmsg/cipher"
	"github.com/skycoin/dmsg/disc"
)

// DefaultDMSGClient creates http Client using default discovery service
func DefaultDMSGClient(pubKey cipher.PubKey, secKey cipher.SecKey) *http.Client {
	// TODO check is there better way to handle pub and sec key
	return DMSGClient(DefaultDiscoveryURL, pubKey, secKey)
}

// DMSGClient creates http Client using provided discovery service and public / secret keypair
func DMSGClient(dicoveryAddress string, pubKey cipher.PubKey, secKey cipher.SecKey) *http.Client {
	transport := DMSGTransport{
		Discovery: disc.NewHTTP(dicoveryAddress),
	}
	timeout, err := time.ParseDuration("30s")
	if err != nil {
		//TODO add log
		timeout = time.Minute
	}
	return &http.Client{
		Transport: transport,
		Jar:       nil,
		Timeout:   timeout,
	}
}
