// Package discovery creates a Dialer which may use a Service Discovery implementation to look up network address, rather than just DNS.
// This allows you to plug in virtually any service discovery service into clients which use a standard Dial or DialContext function to connect to resources
package discovery

import (
	"context"
	"net"
	"time"
)

// Discoverer looks up the hostname in the service discovery system and returns the Network Address
type Discoverer interface {
	// Discover the network address for a given network and hostname
	//
	// `network` is the name of the network (for example, "tcp", "udp")
	// `addr` is the address to discover ("example.com:80")
	//
	// If the network address is not discoverable by the discoverer, it should return an error type that implements `BypassDiscovery() bool` and returns `true`
	// Returning discovery.ErrUndiscoverable will satisfy this requirement
	Discover(network, addr string) (net.Addr, error)
}

type bypasser interface {
	BypassDiscovery() bool
}

// DiscovererFunc implements Discoverer for functions
type DiscovererFunc func(network, host string) (net.Addr, error)

// Discover implementation for Discoverer
func (fn DiscovererFunc) Discover(network, addr string) (net.Addr, error) {
	return fn(network, addr)
}

// Dialer uses the Discoverer to dial the service instead of the given network/addr
type Dialer struct {
	Discoverer    Discoverer
	DialerContext func(ctx context.Context, network, addr string) (net.Conn, error)
}

// NewDialer creates a new discoverer dialer with the default embedded dialer
func NewDialer(d Discoverer) *Dialer {
	return &Dialer{
		DialerContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		Discoverer: d,
	}
}

// DialContext implements net.Dailer DialContext with custom address discovery code provided by the Discoverer
// If the network/address is undiscoverable (Discoverer returns ErrUndiscoverable), it will instead dial the network
// address directly without any service discovery beyond resolving the address with the net.Dailer resolver
// If Discoverer is nil, it will also just dial the address directly without any discovery
func (d *Dialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	if d.Discoverer == nil {
		return d.DialerContext(ctx, network, addr)
	}
	// Either the matcher was not provided (match all) or it matched
	// So use the discoverer to get the network/addr instead
	naddr, err := d.Discoverer.Discover(network, addr)
	if err != nil {
		if b, ok := err.(bypasser); ok && b.BypassDiscovery() {
			// Bypass the discovery service and dial the address normally
			return d.DialerContext(ctx, network, addr)
		}
		return nil, err
	}
	return d.DialerContext(ctx, naddr.Network(), naddr.String())
}

// Dial implements net.Dailer Dial with custom address discovery code provided by the Discoverer
// Same as calling DialContext(context.Background(), network, addr)
func (d *Dialer) Dial(network, addr string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, addr)
}
