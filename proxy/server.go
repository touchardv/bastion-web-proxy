package proxy

import (
	"context"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/touchardv/bastion-web-proxy/config"
)

var (
	httpServer *http.Server
	sshProxies map[string]*sshproxy
	wg         sync.WaitGroup
)

func Configure(cfg config.Config) {
	httpServer = NewHTTPServer(cfg.HTTPServer)
	sshProxies = make(map[string]*sshproxy)
	for _, c := range cfg.SSHProxies {
		sshProxies[c.Name] = NewSSHProxy(c)
	}
}

func Run(ctx context.Context) {
	for _, s := range sshProxies {
		wg.Add(1)
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

	for _, p := range sshProxies {
		log.Debug("Stopping: ssh proxy - ", p.cfg.Name)
		p.Stop()
	}
}
