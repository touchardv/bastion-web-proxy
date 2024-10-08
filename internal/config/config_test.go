package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRetrieving(t *testing.T) {
	cwd, _ := os.Getwd()
	cfg := Config{}
	err := retrieve(cwd, "config.yaml.example", &cfg)

	assert.Nil(t, err)
	assert.Equal(t, "127.0.0.1", cfg.Address)
	assert.True(t, cfg.HTTPServer.Enabled)
	assert.Equal(t, 8080, cfg.HTTPServer.Port)
	assert.Equal(t, "proxy.pac", cfg.HTTPServer.PACFile)

	assert.Equal(t, 2, len(cfg.SSHProxies))
	assert.Equal(t, SSHProxy{Name: "foo", Host: "foo.com", Port: 0, ForwardedPorts: nil,
		Socks5Enabled: true,
		Socks5Port:    1080, Username: "mrfoo"},
		cfg.SSHProxies[0])

	assert.Equal(t, SSHProxy{Name: "bar", Host: "bar.com", Port: 6666, ForwardedPorts: map[uint]RemoteServer{
		1234: {Host: "one.bar.com", Port: 1234},
		5678: {Host: "two.bar.com", Port: 5678}},
		Socks5Enabled: false,
		Socks5Port:    1081, Username: "mrbar"},
		cfg.SSHProxies[1])
}
