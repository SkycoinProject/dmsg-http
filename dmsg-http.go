package dmsghttp

import (
	"net/http"
	"time"

	"github.com/SkycoinProject/dmsg"
	"github.com/SkycoinProject/dmsg/cipher"
	"github.com/SkycoinProject/dmsg/disc"
	"github.com/SkycoinProject/skycoin/src/util/logging"
)

// DefaultDMSGClient creates http Client using default discovery service
// Default value can be found in dmsghttp.DefaultDiscoveryURL
func DefaultDMSGClient(pubKey cipher.PubKey, secKey cipher.SecKey) *http.Client {
	return DMSGClient(disc.NewHTTP(DefaultDiscoveryURL), pubKey, secKey)
}

// DMSGClient creates http Client using provided discovery service and public / secret keypair
// Returned client is using dmsg transport protocol instead of tcp for establishing connection
func DMSGClient(discovery disc.APIClient, pubKey cipher.PubKey, secKey cipher.SecKey) *http.Client {
	transport := DMSGTransport{
		Discovery:  discovery,
		PubKey:     pubKey,
		SecKey:     secKey,
		RetryCount: 20,
	}
	transport.dmsgC = dmsg.NewClient(pubKey, secKey, discovery, dmsg.SetLogger(logging.MustGetLogger("dmsgC_httpC")))

	return &http.Client{
		Transport: transport,
		Jar:       nil,
		Timeout:   time.Second * 30,
	}
}
