package proxy

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/touchardv/bastion-web-proxy/internal/config"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type sshConnection struct {
	name   string
	host   string
	cfg    *ssh.ClientConfig
	client *ssh.Client
	mux    sync.Mutex
}

func newSSHConnection(cfg config.SSHProxy) *sshConnection {
	return &sshConnection{
		name: cfg.Name,
		host: cfg.Host,
		cfg: &ssh.ClientConfig{
			User:            cfg.Username,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         30 * time.Second,
		},
	}
}

func (c *sshConnection) Close() {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.client != nil {
		c.client.Close()
		c.client = nil
		log.Info("ssh connection down - ", c.name)
	}
}

func (c *sshConnection) Tunnel(ctx context.Context, server config.RemoteServer) (net.Conn, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	err := c.dial()
	if err == nil {
		addrs, err := c.resolve(ctx, server.Host)
		if err == nil {
			return c.client.Dial("tcp", fmt.Sprint(addrs[0], ":", server.Port))
		}
		return nil, err
	}
	return nil, err
}

func (c *sshConnection) dial() error {
	if c.client != nil {
		return nil
	}

	socket := os.Getenv("SSH_AUTH_SOCK")
	conn, err := net.Dial("unix", socket)
	if err == nil {
		defer conn.Close()

		agentClient := agent.NewClient(conn)
		c.cfg.Auth = []ssh.AuthMethod{
			ssh.PublicKeysCallback(agentClient.Signers),
		}
		client, err := ssh.Dial("tcp", fmt.Sprint(c.host, ":22"), c.cfg)
		if err == nil {
			log.Info("ssh connection up - ", c.name)
			c.client = client
			go c.watchdog()
		}
		return err
	}
	return err
}

func (c *sshConnection) watchdog() {
	err := c.client.Wait()

	log.Errorf("ssh connection error - %s : %s", c.name, err)
	c.Close()
}

func (c *sshConnection) resolve(ctx context.Context, host string) ([]net.IP, error) {
	r := net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			// resolve using tcp on the remote ssh server
			return c.client.Dial("tcp", "127.0.0.1:53")
		},
	}
	return r.LookupIP(ctx, "ip4", host)
}
