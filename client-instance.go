package dmsghttp

import (
	"errors"
	"sync"

	"github.com/SkycoinProject/dmsg/cipher"
	"github.com/SkycoinProject/dmsg/disc"

	"github.com/SkycoinProject/dmsg"
)

var (
	clients = struct {
		sync.RWMutex
		entries map[cipher.PubKey]*dmsg.Client
	}{entries: make(map[cipher.PubKey]*dmsg.Client)}

	errCreate error = errors.New("dmsg client don't exists and was not created successfully")
)

//GetClient returns DMSG client instance.
func GetClient(pubKey cipher.PubKey, secKey cipher.SecKey) (*dmsg.Client, error) {
	if val, ok := clients.entries[pubKey]; ok {
		return val, nil
	}

	clients.Lock()
	clients.entries[pubKey] = dmsg.NewClient(pubKey, secKey, disc.NewHTTP(DefaultDiscoveryURL), dmsg.DefaultConfig())
	clients.Unlock()

	if clients.entries[pubKey] != nil {
		return clients.entries[pubKey], nil
	}

	return nil, errCreate
}
