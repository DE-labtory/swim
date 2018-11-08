package swim

// PacketTransportConfig is used to configure a net transport
type PacketTransportConfig struct {
	// BindAddrs is representing list of address to use for  UDP communication
	BindAddress string

	// BindPort is the port to listen on. for each address specified above
	BindPort int
}

// NetTransport implements Transport interface, which is used ONLY for connectionless UDP packet operations
type PacketTransport struct {
	config *PacketTransportConfig
}