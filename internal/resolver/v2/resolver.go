package resolver

import (
	"fmt"
	"strings"
	"time"

	"github.com/miekg/dns"
)

const (
	ModeIpv4    = "ipv4"
	ModeIpv6    = "ipv6"
	ModeDefault = ModeIpv4
)

const (
	TimeoutDefault = time.Duration(15 * time.Second)
)

type Response struct {
	Name      string   `json:"name"`
	Addresses []string `json:"addresses"`
	rcode     int
}

type Resolver struct {
	resolver *dns.Client
	timeout  time.Duration
	mode     uint16
}

func NewResolver() *Resolver {
	return &Resolver{
		resolver: &dns.Client{
			SingleInflight: true,
			Timeout:        TimeoutDefault,
		},
		timeout: TimeoutDefault,
		mode:    getQueryTypes(ModeDefault),
	}
}

func (r *Resolver) WithTimeout(t time.Duration) *Resolver {
	r.resolver.Timeout = t
	return r
}

func (r *Resolver) WithMode(m string) *Resolver {
	r.mode = getQueryTypes(m)
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

		msgQuery := &dns.Msg{
			MsgHdr: dns.MsgHdr{
				Id:               dns.Id(),
				RecursionDesired: true,
			},
			Question: []dns.Question{
				{
					Name:   dns.Fqdn(name),
					Qtype:  r.mode,
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

		result[len(result)-1].rcode = msqResponse.MsgHdr.Rcode

		for _, response := range msqResponse.Answer {
			result[len(result)-1].Addresses = append(result[len(result)-1].Addresses, strings.Split(response.String(), "\t")[4])
		}

	}

	return result, nil
}

func FilterResponsesByRcode(rs []Response, rcode int) []Response {
	result := make([]Response, 0)

	for _, r := range rs {
		if r.rcode == rcode {
			result = append(result, r)
		}
	}

	return result
}

func FilterResponsesNoerror(rs []Response) []Response {
	return FilterResponsesByRcode(rs, dns.StringToRcode["NOERROR"])
}

func FilterResponsesNxdomain(rs []Response) []Response {
	return FilterResponsesByRcode(rs, dns.StringToRcode["NXDOMAIN"])
}

func getQueryTypes(m string) uint16 {
	switch m {
	case ModeIpv4:
		return dns.TypeA
	case ModeIpv6:
		return dns.TypeAAAA
	default:
		return dns.TypeA
	}
}
