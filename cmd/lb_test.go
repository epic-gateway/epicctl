package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
)

func TestParsePorts(t *testing.T) {
	var (
		ports []corev1.ServicePort
		err   error
	)

	// Default protocol (should be TCP)
	ports, err = parsePorts("8888")
	assert.Nil(t, err, "parse failed")
	assert.Equal(t, 1, len(ports), "parse returned wrong number of ports")
	assert.Equal(t, int32(8888), ports[0].Port, "parse returned wrong port")
	assert.Equal(t, v1.ProtocolTCP, ports[0].Protocol, "parse returned wrong protocol")

	// Explicit protocol
	ports, err = parsePorts("UDP/8888")
	assert.Nil(t, err, "parse failed")
	assert.Equal(t, 1, len(ports), "parse returned wrong number of ports")
	assert.Equal(t, int32(8888), ports[0].Port, "parse returned wrong port")
	assert.Equal(t, v1.ProtocolUDP, ports[0].Protocol, "parse returned wrong protocol")

	// Multiple ports
	ports, err = parsePorts("UDP/8888,TCP/9999,7777")
	assert.Nil(t, err, "parse failed")
	assert.Equal(t, 3, len(ports), "parse returned wrong number of ports")
	assert.Equal(t, int32(8888), ports[0].Port, "parse returned wrong port")
	assert.Equal(t, v1.ProtocolUDP, ports[0].Protocol, "parse returned wrong protocol")
	assert.Equal(t, int32(9999), ports[1].Port, "parse returned wrong port")
	assert.Equal(t, v1.ProtocolTCP, ports[1].Protocol, "parse returned wrong protocol")
	assert.Equal(t, int32(7777), ports[2].Port, "parse returned wrong port")
	assert.Equal(t, v1.ProtocolTCP, ports[2].Protocol, "parse returned wrong protocol")
}
