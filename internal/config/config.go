package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const DefaultFilename = "config.yaml"

type Server struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
	PACFile string `yaml:"pacFile"`
}

type RemoteServer struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func (rs *RemoteServer) String() string {
	return fmt.Sprint(rs.Host, ":", rs.Port)
}

type ForwardedPorts map[uint]RemoteServer

type SSHProxy struct {
	Name           string         `yaml:"name"`
	Host           string         `yaml:"host"`
	ForwardedPorts ForwardedPorts `yaml:"forwardedPorts"`
	Socks5Port     int            `yaml:"socks5Port"`
	Username       string         `yaml:"username"`
}

type Config struct {
	HTTPServer Server     `yaml:"httpServer"`
	SSHProxies []SSHProxy `yaml:"sshProxies"`
}

// Retrieve reads and parses the configuration file.
func Retrieve(location string, cfg interface{}) error {
	return retrieve(location, DefaultFilename, cfg)
}

func retrieve(location string, name string, cfg interface{}) error {
	filename := filepath.Join(location, name)
	content, err := ioutil.ReadFile(filename)
	if err == nil {
		err = yaml.Unmarshal(content, cfg)
	}
	return err
}
