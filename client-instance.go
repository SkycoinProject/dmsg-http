package dmsghttp

import (
	"errors"
	"sync"

	"github.com/SkycoinProject/dmsg/cipher"
	"github.com/SkycoinProject/dmsg/disc"

	"github.com/SkycoinProject/dmsg"
)

var (
	singleton *dmsg.Client
	once      sync.Once

	usedPub cipher.PubKey
	usedSec cipher.SecKey

	errWrongPubKeyUsed error = errors.New("dmsg client already initialized with different pub key")
	errWrongSecKeyUsed error = errors.New("dmsg client already initialized with different sec key")
)

func getClient(pubKey cipher.PubKey, secKey cipher.SecKey) (*dmsg.Client, error) {
	once.Do(func() {
		singleton = dmsg.NewClient(pubKey, secKey, disc.NewHTTP(DefaultDiscoveryURL), dmsg.DefaultConfig())
		usedPub = pubKey
		usedSec = secKey
	})
	if pubKey != usedPub {
		return nil, errWrongPubKeyUsed
	}
	if secKey != usedSec {
		return nil, errWrongSecKeyUsed
	}
	return singleton, nil
}
