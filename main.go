package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/touchardv/bastion-web-proxy/config"
	"github.com/touchardv/bastion-web-proxy/proxy"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

var (
	logLevel       *string
	configLocation *string
	cfg            config.Config
)

func main() {
	logLevel = pflag.String("log-level", log.InfoLevel.String(), "The logging level (trace, debug, info...)")
	configLocation = pflag.String("config-location", ".", "The path to the directory where the configuration file is stored.")
	pflag.Parse()

	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Fatal(err)
	}
	log.SetFormatter(&log.TextFormatter{DisableLevelTruncation: true, FullTimestamp: true})
	log.SetLevel(level)

	log.Info("Starting...")
	err = config.Retrieve(*configLocation, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	proxy.Configure(cfg)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		cancelFunc()
		proxy.Stop()
	}()

	proxy.Run(ctx)

	log.Info("...Stopped")
	log.Exit(0)
}
