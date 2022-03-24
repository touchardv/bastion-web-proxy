package proxy

import (
	"context"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/touchardv/bastion-web-proxy/config"
)

func NewHTTPServer(cfg config.Server) *http.Server {
	path := fmt.Sprint("/", cfg.PACFile)
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, cfg.PACFile)
	})
	return &http.Server{
		Addr:         fmt.Sprint(cfg.Address, ":", cfg.Port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func RunHTTPServer() {
	log.Info("http server listening on: ", httpServer.Addr)
	err := httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Error(err)
	}
}

func StopHTTPServer() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	httpServer.Shutdown(ctx)
}
