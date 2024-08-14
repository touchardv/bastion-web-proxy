package proxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToDialAddress(t *testing.T) {
	assert.Equal(t, "bla:22", toDialAddress(&sshConnection{host: "bla"}))
	assert.Equal(t, "ble:6666", toDialAddress(&sshConnection{host: "ble", port: 6666}))
}
