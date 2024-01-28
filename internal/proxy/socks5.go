package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"sync/atomic"

	log "github.com/sirupsen/logrus"
	"github.com/things-go/go-socks5"
	"github.com/things-go/go-socks5/statute"
	"github.com/touchardv/bastion-web-proxy/internal/config"
)

func (s *sshproxy) startSocks5Server(ctx context.Context) {
	localAddr, _ := net.ResolveTCPAddr("tcp", fmt.Sprint("127.0.0.1:", s.cfg.Socks5Port))
	log.Info("socks5 server listening on: ", localAddr)
	s.socksServer = socks5.NewServer(
		socks5.WithLogger(s),
		socks5.WithResolver(s),
		socks5.WithConnectHandle(s.handleSocks5Connect),
	)

	listener, err := net.ListenTCP("tcp", localAddr)
	if err != nil {
		log.Error("Error listening: ", err)
		return
	}
	defer listener.Close()
	s.socksListener = listener
	err = s.socksServer.Serve(listener)
	select {
	case <-ctx.Done():
		return
	default:
		if err != nil {
			log.Error("Error accepting connection: ", err)
		}
	}
}

func (s *sshproxy) Resolve(ctx context.Context, name string) (context.Context, net.IP, error) {
	// dummy resolving
	return ctx, net.IP{}, nil
}

func (s *sshproxy) Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

func (s *sshproxy) handleSocks5Connect(ctx context.Context, writer io.Writer, request *socks5.Request) error {
	remoteServer := config.RemoteServer{
		Host: request.RawDestAddr.FQDN,
		Port: request.RawDestAddr.Port,
	}
	target, err := s.sshConnection.Tunnel(ctx, remoteServer)
	if err != nil {
		resp := errToResponse(err)
		if err := socks5.SendReply(writer, resp, nil); err != nil {
			return fmt.Errorf("failed to send reply, %v", err)
		}
		return fmt.Errorf("connect to %v failed, %v", request.RawDestAddr, err)
	}
	defer target.Close()
	log.Info("Connected: ", request.RawDestAddr.FQDN)
	defer log.Info("Disconnected: ", request.RawDestAddr.FQDN)
	atomic.AddInt32(&s.connCount, 1)
	defer atomic.AddInt32(&s.connCount, -1)

	if err := socks5.SendReply(writer, statute.RepSuccess, target.LocalAddr()); err != nil {
		return fmt.Errorf("failed to send reply, %v", err)
	}

	// note: re-use the socks5 proxy routines
	errCh := make(chan error, 2)
	go func() { errCh <- s.socksServer.Proxy(target, request.Reader) }()
	go func() { errCh <- s.socksServer.Proxy(writer, target) }()

	for i := 0; i < 2; i++ {
		e := <-errCh
		if e != nil {
			log.Warn(e)
			return e
		}
	}
	return nil
}

func errToResponse(err error) uint8 {
	msg := err.Error()
	resp := statute.RepHostUnreachable
	if strings.Contains(msg, "refused") {
		resp = statute.RepConnectionRefused
	} else if strings.Contains(msg, "network is unreachable") {
		resp = statute.RepNetworkUnreachable
	}
	return resp
}
