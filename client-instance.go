package dmsghttp

import (
	"errors"
	"sync"

	"github.com/SkycoinProject/dmsg/cipher"
	"github.com/SkycoinProject/dmsg/disc"

	"github.com/SkycoinProject/dmsg"
)

// Defaults for dmsg configuration, such as discovery URL
const (
	DefaultDiscoveryURL = "http://dmsg.discovery.skywire.cc"
)

// DMSGClient holds parameters required to instantiate a dmsg client instance
type DMSGClient struct {
	PubKey    cipher.PubKey
	SecKey    cipher.SecKey
	Discovery disc.APIClient
	Config    *dmsg.Config
}

var (
	clients = struct {
		sync.RWMutex
		entries map[cipher.PubKey]*dmsg.Client
	}{entries: make(map[cipher.PubKey]*dmsg.Client)}

	errCreate error = errors.New("dmsg client don't exists and was not created successfully")
)

func DefaultDMSGClient(pubKey cipher.PubKey, secKey cipher.SecKey) *DMSGClient {
	return &DMSGClient{
		PubKey:    pubKey,
		SecKey:    secKey,
		Discovery: disc.NewHTTP(DefaultDiscoveryURL),
		Config:    dmsg.DefaultConfig(),
	}
}

//GetClient returns DMSG client instance.
func GetClient(dmsgC *DMSGClient) (*dmsg.Client, error) {
	if val, ok := clients.entries[dmsgC.PubKey]; ok {
		return val, nil
	}

	clients.Lock()
	clients.entries[dmsgC.PubKey] = dmsg.NewClient(dmsgC.PubKey, dmsgC.SecKey, dmsgC.Discovery, dmsgC.Config)
	clients.Unlock()

	if clients.entries[dmsgC.PubKey] != nil {
		return clients.entries[dmsgC.PubKey], nil
	}

	return nil, errCreate
}
