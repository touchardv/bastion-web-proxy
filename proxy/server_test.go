package proxy

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/touchardv/bastion-web-proxy/config"
)

func TestLifecycle(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	cfg := config.Config{
		HTTPServer: config.Server{Address: "127.0.0.1", Port: 0},
		SSHProxies: []config.SSHProxy{
			{Name: "test", Host: "127.0.0.1", Username: "foo", ForwardedPorts: config.ForwardedPorts{
				12345: config.RemoteServer{Host: "target", Port: 1234},
			}},
		},
	}

	go func() {
		time.Sleep(1 * time.Second)
		Stop()
	}()

	Configure(cfg)
	Run()
}
