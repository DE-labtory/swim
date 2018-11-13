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

package swim

import (
	"fmt"
	"net"
	"time"
	"sync"
)

const (
	// udpPacketBufSize is used to buffer incoming packets during read
	// operations.
	udpPacketBufSize = 65536

	// udpRecvBufSize is a large buffer size that we attempt to set UDP
	// sockets to in order to handle a large volume of messages.
	udpRecvBufSize = 2 * 1024 * 1024
)

// PacketTransportConfig is used to configure a udp transport
type PacketTransportConfig struct {
	// BindAddrs is representing an address to use for UDP communication
	BindAddress string

	// BindPort is the port to listen on, for the address specified above
	BindPort int
}

// PacketTransport implements Transport interface, which is used ONLY for connectionless UDP packet operations
type PacketTransport struct {
	config         *PacketTransportConfig
	packetCh       chan *Packet
	packetListener *net.UDPConn
	isShutDown     bool
	lock           sync.RWMutex
}

func NewPacketListener(ip net.IP, port int) (*net.UDPConn, error) {
	udpAddr := &net.UDPAddr{IP: ip, Port: port}
	udpListener, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return nil, fmt.Errorf("Failed to start UDP listener on %s port %d : %v", udpAddr.String(), port, err)
	}

	if err := setUDPRecvBuf(udpListener); err != nil {
		return nil, fmt.Errorf("Failed to resize UDP buffer: %v", err)
	}

	return udpListener, nil
}

func NewPacketTransport(config *PacketTransportConfig) (*PacketTransport, error) {
	ip := net.ParseIP(config.BindAddress)
	port := config.BindPort

	packetListener, err := NewPacketListener(ip, port)
	if err != nil {
		return nil, err
	}

	t := PacketTransport{
		config:         config,
		packetCh:       make(chan *Packet),
		packetListener: packetListener,
		isShutDown:     false,
		lock:           sync.RWMutex{},
	}

	go t.listen()

	return &t, nil
}

func (t *PacketTransport) WriteTo(b []byte, addr string) (time.Time, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return time.Time{}, err
	}

	if _, err = t.packetListener.WriteTo(b, udpAddr); err != nil {
		return time.Time{}, err
	}

	return time.Now(), nil
}

func (t *PacketTransport) PacketCh() <-chan *Packet {
	return t.packetCh
}

func (t *PacketTransport) Shutdown() error {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.isShutDown = true
	return t.packetListener.Close()
}

// udpListen is a long running goroutine that accepts incoming UDP packets and
// hands them off to the packet channel.
func (t *PacketTransport) listen() {
	for {
		// Do a blocking read into a fresh buffer. Grab a time stamp as
		// close as possible to the I/O.
		buf := make([]byte, udpPacketBufSize)
		n, addr, err := t.packetListener.ReadFrom(buf)
		ts := time.Now()

		if t.isShutDown {
			break
		}

		if err != nil {
			fmt.Printf("[ERR] Error reading UDP packet: %v", err)
			continue
		}

		// Check the length - it needs to have at least one byte to be a
		// proper message.
		if n < 1 {
			fmt.Printf("[ERR] UDP packet too short (%d bytes) %s", len(buf), addr.String())
			continue
		}

		// Ingest the packet.
		t.packetCh <- &Packet{
			Buf:       buf[:n],
			Addr:      addr,
			Timestamp: ts,
		}
	}
}

// setUDPRecvBuf is used to resize the UDP receive window. The function
// attempts to set the read buffer to `udpRecvBuf` but backs off until
// the read buffer can be set.
func setUDPRecvBuf(c *net.UDPConn) error {
	size := udpRecvBufSize
	var err error
	for size > 0 {
		// If the read buffer is set, return nil
		if err = c.SetReadBuffer(size); err == nil {
			return nil
		}

		// If the read buffer size is large, back off.
		// Because smaller size will be more easy to set to read buffer.
		size = size / 2
	}
	return err
}
