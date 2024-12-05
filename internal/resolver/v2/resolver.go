package resolver

import (
	"fmt"
	"slices"
	"strings"
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
	Addresses []string `json:"addresses"`
	Error     error    `json:"-"`
}

type Resolver struct {
	resolver *dns.Client
	timeout  time.Duration
	mode     []uint16
}

func NewResolver() *Resolver {
	return &Resolver{
		resolver: &dns.Client{
			SingleInflight: true,
			Timeout:        TimeoutDefault,
		},
		timeout: TimeoutDefault,
		mode:    getQueryType(ModeDefault),
	}
}

func (r *Resolver) WithTimeout(t time.Duration) *Resolver {
	r.resolver.Timeout = t
	return r
}

func (r *Resolver) WithMode(m string) *Resolver {
	r.mode = getQueryType(m)
	return r
}

func (r *Resolver) Resolve(dn []string) ([]Response, error) {
	result := make([]Response, 0)
	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		return nil, err
	}

	for _, name := range dn {
		result = append(result, Response{
			Name:      name,
			Addresses: make([]string, 0),
		},
		)

		for _, queryType := range r.mode {
			msgQuery := &dns.Msg{
				MsgHdr: dns.MsgHdr{
					Id:               dns.Id(),
					RecursionDesired: true,
				},
				Question: []dns.Question{
					{
						Name:   dns.Fqdn(name),
						Qtype:  queryType,
						Qclass: dns.ClassINET,
					},
				},
			}

			msqResponse, _, err := r.resolver.Exchange(
				msgQuery,
				fmt.Sprintf("%s:%s", config.Servers[0], "53"),
			)

			if err != nil {
				return nil, err
			}

			for _, response := range msqResponse.Answer {
				result = createOrAppendResponse(result, name, strings.Split(response.String(), "\t")[4])
			}
		}

	}

	return result, nil
}

func getQueryType(m string) []uint16 {
	switch m {
	case ModeIpv4:
		return []uint16{dns.TypeA}
	case ModeIpv6:
		return []uint16{dns.TypeAAAA}
	default:
		return []uint16{dns.TypeA, dns.TypeAAAA}
	}
}

func createOrAppendResponse(s []Response, name string, address string) []Response {
	index := slices.IndexFunc(s,
		func(r Response) bool {
			return r.Name == name
		},
	)
	if index == -1 {
		return append(s, Response{
			Name: name,
			Addresses: []string{
				address,
			},
		})
	} else {
		s[index].Addresses = append(s[index].Addresses, address)
		return s
	}
}
