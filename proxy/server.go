package proxy

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/touchardv/bastion-web-proxy/config"
)

var (
	httpServer *http.Server
	sshProxies map[string]*sshproxy
	ctx        context.Context
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
)

func Configure(cfg config.Config) {
	httpServer = NewHTTPServer(cfg.HTTPServer)
	sshProxies = make(map[string]*sshproxy)
	for _, c := range cfg.SSHProxies {
		sshProxies[c.Name] = NewSSHProxy(c)
	}
	ctx, cancelFunc = context.WithCancel(context.Background())
}

func Run() {
	wg.Add(1)
	go func() {
		defer wg.Done()
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		select {
		case <-c:
			Stop()
		case <-ctx.Done():
		}
	}()

	wg.Add(len(sshProxies))
	for _, s := range sshProxies {
		go func(p *sshproxy) {
			defer wg.Done()
			log.Debug("Starting: ssh proxy - ", p.cfg.Name)
			p.Run(ctx)
			log.Debug("Stopped: ssh proxy - ", p.cfg.Name)
		}(s)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Debug("Starting: http server")
		RunHTTPServer()
		log.Debug("Stopped: http server")
	}()

	wg.Wait()
}

func Stop() {
	log.Debug("Stopping: http server")
	StopHTTPServer()

	cancelFunc()
	for _, p := range sshProxies {
		log.Debug("Stopping: ssh proxy - ", p.cfg.Name)
		p.Stop()
	}
}
