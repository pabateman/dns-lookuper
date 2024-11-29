package resolver

import (
	"context"
	"net"
	"time"
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
	net.Resolver
	timeout time.Duration
	mode    string
}

func NewResolver() *Resolver {
	return &Resolver{
		net.Resolver{},
		TimeoutDefault,
		getIPMode(ModeDefault),
	}
}

func (r *Resolver) WithTimeout(t time.Duration) *Resolver {
	r.timeout = t
	return r
}

func (r *Resolver) WithMode(m string) *Resolver {
	r.mode = getIPMode(m)
	return r
}

func (r *Resolver) Resolve(dn []string) ([]Response, error) {
	responses := make([]Response, 0)

	for _, name := range dn {
		response := Response{
			Name:      name,
			Addresses: make([]net.IP, 0),
			Error:     nil,
		}

		ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
		defer cancel()

		answer, err := r.LookupIP(ctx, r.mode, name)
		if err != nil {
			if dnsError, ok := err.(*net.DNSError); ok {
				if dnsError.IsNotFound || dnsError.IsTimeout {
					response.Error = err
					responses = append(responses, response)
					continue
				} else {
					return nil, err
				}
			} else if addrError, ok := err.(*net.AddrError); ok {
				if addrError.Err != "no suitable address found" {
					return nil, err
				}
			} else {
				return nil, err
			}
		}

		if answer != nil {
			response.Addresses = answer
		}

		responses = append(responses, response)
		cancel()
	}
	return responses, nil
}

func getIPMode(m string) string {
	switch m {
	case ModeIpv4:
		return "ip4"
	case ModeIpv6:
		return "ip6"
	case ModeAll:
		return "ip"
	default:
		return "unsupported"
	}
}
