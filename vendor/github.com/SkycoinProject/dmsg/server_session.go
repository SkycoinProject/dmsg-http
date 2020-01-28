package dmsg

import (
	"io"
	"net"

	"github.com/SkycoinProject/yamux"

	"github.com/SkycoinProject/dmsg/netutil"
	"github.com/SkycoinProject/dmsg/noise"
)

// ServerSession represents a session from the perspective of a dmsg server.
type ServerSession struct {
	*SessionCommon
}

func makeServerSession(entity *EntityCommon, conn net.Conn) (ServerSession, error) {
	var sSes ServerSession
	sSes.SessionCommon = new(SessionCommon)
	sSes.nMap = make(noise.NonceMap)
	if err := sSes.SessionCommon.initServer(entity, conn); err != nil {
		return sSes, err
	}
	return sSes, nil
}

// Close implements io.Closer
func (ss *ServerSession) Close() (err error) {
	if ss != nil {
		if ss.SessionCommon != nil {
			err = ss.SessionCommon.Close()
		}
		ss.rMx.Lock()
		ss.nMap = nil
		ss.rMx.Unlock()
	}
	return err
}

// Serve serves the session.
func (ss *ServerSession) Serve() {
	for {
		yStr, err := ss.ys.AcceptStream()
		if err != nil {
			switch err {
			case yamux.ErrSessionShutdown, io.EOF:
				ss.log.WithError(err).Info("Stopping session...")
			default:
				ss.log.WithError(err).Warn("Failed to accept stream, stopping session...")
			}
			return
		}

		ss.log.Info("Serving stream.")
		go func(yStr *yamux.Stream) {
			err := ss.serveStream(yStr)
			ss.log.WithError(err).Info("Stopped stream.")
		}(yStr)
	}
}

func (ss *ServerSession) serveStream(yStr *yamux.Stream) error {
	readRequest := func() (StreamRequest, error) {
		obj, err := ss.readObject(yStr)
		if err != nil {
			return StreamRequest{}, err
		}
		req, err := obj.ObtainStreamRequest()
		if err != nil {
			return StreamRequest{}, err
		}
		// TODO(evanlinjin): Implement timestamp tracker.
		if err := req.Verify(0); err != nil {
			return StreamRequest{}, err
		}
		if req.SrcAddr.PK != ss.rPK {
			return StreamRequest{}, ErrReqInvalidSrcPK
		}
		return req, nil
	}

	// Read request.
	req, err := readRequest()
	if err != nil {
		return err
	}

	// Obtain next session.
	ss2, ok := ss.entity.serverSession(req.DstAddr.PK)
	if !ok {
		return ErrReqNoSession
	}

	// Forward request and obtain/check response.
	yStr2, resp, err := ss2.forwardRequest(req)
	if err != nil {
		return err
	}

	// Forward response.
	if err := ss.writeObject(yStr, resp); err != nil {
		return err
	}

	// Serve stream.
	return netutil.CopyReadWriteCloser(yStr, yStr2)
}

func (ss *ServerSession) forwardRequest(req StreamRequest) (yStr *yamux.Stream, respObj SignedObject, err error) {
	defer func() {
		if err != nil && yStr != nil {
			ss.log.
				WithError(yStr.Close()).
				Debugf("After forwardRequest failed, the yamux stream is closed.")
		}
	}()

	if yStr, err = ss.ys.OpenStream(); err != nil {
		return nil, nil, err
	}
	if err = ss.writeObject(yStr, req.raw); err != nil {
		return nil, nil, err
	}
	if respObj, err = ss.readObject(yStr); err != nil {
		return nil, nil, err
	}
	var resp StreamResponse
	if resp, err = respObj.ObtainStreamResponse(); err != nil {
		return nil, nil, err
	}
	if err = resp.Verify(req); err != nil {
		return nil, nil, err
	}
	return yStr, respObj, nil
}
