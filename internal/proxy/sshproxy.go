package proxy

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"

	"github.com/touchardv/bastion-web-proxy/internal/config"

	log "github.com/sirupsen/logrus"
	socks5 "github.com/things-go/go-socks5"
)

type sshproxy struct {
	cfg           config.SSHProxy
	socksServer   *socks5.Server
	socksListener *net.TCPListener
	sshConnection *sshConnection
	fwdListeners  map[uint]*net.TCPListener
	connCount     int32
	mutex         sync.Mutex
	wg            sync.WaitGroup
}

func NewSSHProxy(cfg config.SSHProxy) *sshproxy {
	p := &sshproxy{
		cfg:           cfg,
		sshConnection: newSSHConnection(cfg),
		fwdListeners:  make(map[uint]*net.TCPListener),
	}
	p.socksServer = newSocks5Server(p)
	return p
}

func (s *sshproxy) Run(ctx context.Context, localAddress string) {
	s.wg.Add(len(s.cfg.ForwardedPorts))
	for localPort, remoteServer := range s.cfg.ForwardedPorts {
		go func(localPort uint, remoteServer config.RemoteServer) {
			defer s.wg.Done()
			log.Debugf("Starting forward server: %d -> %s:%d", localPort, remoteServer.Host, remoteServer.Port)
			s.startForwardServer(ctx, localAddress, localPort, remoteServer)
			log.Debugf("Stopped forward server: %d -> %s:%d", localPort, remoteServer.Host, remoteServer.Port)
		}(localPort, remoteServer)
	}

	if s.cfg.Socks5Enabled {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			log.Debug("Starting: socks5 server - ", s.cfg.Name)
			s.startSocks5Server(ctx, localAddress)
			log.Debug("Stopped: socks5 server - ", s.cfg.Name)
		}()
	}

	s.wg.Wait()
	s.sshConnection.Close()
}

func (s *sshproxy) startForwardServer(ctx context.Context, localAddress string, localPort uint, remoteServer config.RemoteServer) {
	localAddr, _ := net.ResolveTCPAddr("tcp", fmt.Sprint(localAddress, ":", localPort))
	listener, err := net.ListenTCP("tcp", localAddr)
	if err != nil {
		log.Error("Error listening: ", err)
		return
	}
	defer listener.Close()
	s.mutex.Lock()
	s.fwdListeners[localPort] = listener
	s.mutex.Unlock()

	log.Info("forward server listening on: ", localAddr)
	for {
		inConn, err := listener.Accept()
		select {
		case <-ctx.Done():
			return
		default:
			if err != nil {
				log.Error("Error accepting connection: ", err)
				return
			}
			go s.handlePortForwardConnect(localPort, inConn)
		}
	}
}

func (s *sshproxy) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.cfg.Socks5Enabled {
		log.Debug("Stopping: socks5 server - ", s.cfg.Name)
		s.socksListener.Close()
	}

	for localPort, remoteServer := range s.cfg.ForwardedPorts {
		log.Debugf("Stopping forward server: %d -> %s:%d", localPort, remoteServer.Host, remoteServer.Port)
		s.fwdListeners[localPort].Close()
	}
}

func (s *sshproxy) handlePortForwardConnect(localPort uint, inConn net.Conn) {
	defer inConn.Close()

	ctx := context.Background()
	remoteServer := s.cfg.ForwardedPorts[localPort]
	log.Debugf("%s -> %s: accepted", inConn.RemoteAddr(), remoteServer.String())
	outConn, err := s.sshConnection.Tunnel(ctx, remoteServer)
	if err != nil {
		log.Errorf("%s -> %s: %s", inConn.RemoteAddr(), remoteServer.String(), err)
		return
	}
	defer outConn.Close()
	log.Debugf("%s -> %s: connected", inConn.RemoteAddr(), remoteServer.String())
	atomic.AddInt32(&s.connCount, 1)
	defer atomic.AddInt32(&s.connCount, -1)

	// note: re-use the socks5 proxy routines
	errCh := make(chan error, 2)
	go func() { errCh <- s.socksServer.Proxy(outConn, inConn) }()
	go func() { errCh <- s.socksServer.Proxy(inConn, outConn) }()

	for i := 0; i < 2; i++ {
		e := <-errCh
		if e != nil {
			log.Warnf("%s -> %s: %s", inConn.RemoteAddr(), remoteServer.String(), e)
			return
		}
	}
	log.Debugf("%s -> %s: disconnected", inConn.RemoteAddr(), remoteServer.String())
}
