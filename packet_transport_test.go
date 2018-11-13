/*
 * Copyright 2018 De-labtory
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package swim_test

import (
	"net"
	"testing"

	"github.com/DE-labtory/swim"
	"github.com/stretchr/testify/assert"
)

func TestNewPacketListener(t *testing.T) {
	// given
	ip := net.ParseIP("127.0.0.1")
	port := 12345

	// when
	l, err := swim.NewPacketListener(ip, port)

	// then
	assert.NotNil(t, l)
	assert.NoError(t, err)

	l.Close()
}

func TestNewPacketTransport(t *testing.T) {
	// given
	c := swim.PacketTransportConfig{
		BindAddress: "127.0.0.1",
		BindPort:    12345,
	}

	// when
	p, err := swim.NewPacketTransport(&c)

	// then
	assert.NotNil(t, p)
	assert.NoError(t, err)

	p.Shutdown()
}

func TestPacketTransport_PacketCh(t *testing.T) {
	// given
	c := swim.PacketTransportConfig{
		BindAddress: "127.0.0.1",
		BindPort:    12345,
	}

	p, err := swim.NewPacketTransport(&c)
	assert.NotNil(t, p)
	assert.NoError(t, err)

	time, err := p.WriteTo([]byte{'a'}, "127.0.0.1:12345")
	assert.NotNil(t, time)
	assert.NoError(t, err)

	// when
	packet := <-p.PacketCh()

	// then
	assert.NotNil(t, packet)
	assert.Equal(t, []byte{'a'}, packet.Buf)
	assert.NotEqual(t, []byte{'b'}, packet.Buf)

	p.Shutdown()
}

func TestPacketTransport_WriteTo(t *testing.T) {
	// given
	c := swim.PacketTransportConfig{
		BindAddress: "127.0.0.1",
		BindPort:    12345,
	}

	p, err := swim.NewPacketTransport(&c)
	assert.NotNil(t, p)
	assert.NoError(t, err)

	// when
	time, err := p.WriteTo([]byte{'a'}, "127.0.0.1:12345")

	// then
	assert.NotNil(t, time)
	assert.NoError(t, err)

	p.Shutdown()
}

func TestPacketTransport_Shutdown(t *testing.T) {
	// given
	c := swim.PacketTransportConfig{
		BindAddress: "127.0.0.1",
		BindPort:    12345,
	}

	p, err := swim.NewPacketTransport(&c)
	assert.NotNil(t, p)
	assert.NoError(t, err)

	// when
	err = p.Shutdown()

	// then
	assert.NoError(t, err)
	_, err = p.WriteTo([]byte{'a'}, "127.0.0.1:12345")
	assert.Error(t, err)
}
