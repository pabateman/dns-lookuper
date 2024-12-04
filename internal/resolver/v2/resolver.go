package resolver

import (
	"net"
	"time"

	"github.com/miekg/dns"
)

const (
	ModeIpv4    = "ipv4"
	ModeIpv6    = "ipv6"
	ModeAll     = "all"
	ModeDefault = ModeAll
)

const (
	TimeoutDefault = time.Duration(15 * time.Second)
)

type Response struct {
	Name      string   `json:"name"`
	Addresses []net.IP `json:"addresses"`
	Error     error    `json:"-"`
}

type Resolver struct {
	resolver *dns.Client
	timeout  time.Duration
	mode     string
}

func NewResolver() *Resolver {
	return &Resolver{
		resolver: &dns.Client{
			Timeout: TimeoutDefault,
		},
		timeout: TimeoutDefault,
	}
}

func (r *Resolver) WithTimeout(t time.Duration) *Resolver {
	r.resolver.Timeout = t
	return r
}

func (r *Resolver) Resolve(dn []string) ([]Response, error) {
	return nil, nil
}
