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
	"net"
	"time"
)

// Packet is used to provide some metadata about incoming packets from peers
// over a packet connection
type Packet struct {
	// Buffer is the contents of the packet.
	Buf []byte

	// From has the address of the peer. This is an actual net.Addr so we
	// can expose some concrete details about incoming packets.
	Addr net.Addr

	// Timestamp is the time when the packet was received.
	Timestamp time.Time
}

// UDP Transport is used to udp abstract over communicating with other peers.
type UDPTransport interface {

	// WriteTo is a packet-oriented interface that fires off the given
	// payload to the given address in a connectionless fashion. This should
	// return a time stamp that's as close as possible to when the packet
	// was transmitted to help make accurate RTT measurements during probes.
	//
	// We also treat the address here as a
	// string, similar to Dial, so it's network neutral, so this usually is
	// in the form of "host:port".
	WriteTo(b []byte, addr string) (time.Time, error)

	// PacketCh returns a channel that can be read to receive incoming
	// packets from other peers.
	PacketCh() <-chan *Packet

	// Shutdown all running go-routine inside UDPTransport
	Shutdown() error
}
