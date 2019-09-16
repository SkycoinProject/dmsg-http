package dmsghttp

import (
	"context"
	"net"
	"net/http"

	"github.com/skycoin/dmsg/cipher"
	"github.com/skycoin/skycoin/src/util/logging"

	"github.com/skycoin/dmsg"
)

type Server struct {
	PubKey cipher.PubKey
	SecKey cipher.SecKey
}

func (s Server) Serve(l net.Listener, handler http.Handler) error {
	ctx := context.Background()
	srv := &http.Server{Handler: handler}
	logger := logging.MustGetLogger("dmsg-server")

	// tcpConn, err := net.Dial("tcp", entry.Server.Address)
	// if err != nil {
	// 	return err
	// }
	// ns, err := noise.New(noise.HandshakeXK, noise.Config{
	// 	LocalPK:   c.pk,
	// 	LocalSK:   c.sk,
	// 	RemotePK:  srvPK,
	// 	Initiator: true,
	// })
	// if err != nil {
	// 	return err
	// }
	// nc, err := noise.WrapConn(tcpConn, ns, TransportHandshakeTimeout)
	// if err != nil {
	// 	return err
	// }

	conn := dmsg.NewClientConn(logger, nc, c.pk, srvPK, c.pm)
	// if err := conn.readOK(); err != nil {
	// 	return nil, err
	// }
	return conn.Serve(ctx) //TODO (srdjan) serve handles incomming but it reguires a whole lot of client side things to be configured
	// return srv.Serve(l)
}
