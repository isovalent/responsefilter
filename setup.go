package responsefilter

import (
	"net"
	"strings"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

func init() { plugin.Register("responsefilter", setup) }

func setup(c *caddy.Controller) error {
	rf, err := parseResponseFilter(c)
	if err != nil {
		return plugin.Error("responsefilter", err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		rf.Next = next
		return rf
	})

	return nil
}

func parseResponseFilter(c *caddy.Controller) (*ResponseFilter, error) {
	rf := &ResponseFilter{
		Rules: []Rule{},
	}

	for c.Next() {
		// responsefilter {
		//     block abc.com 10.1.1.0/24
		//     block xyz.com 192.168.0.0/16 172.16.0.0/12
		// }
		for c.NextBlock() {
			switch c.Val() {
			case "block":
				args := c.RemainingArgs()
				if len(args) < 2 {
					return nil, c.ArgErr()
				}

				domain := args[0]
				if !strings.HasSuffix(domain, ".") {
					domain = domain + "."
				}

				var cidrs []*net.IPNet
				for _, cidrStr := range args[1:] {
					_, cidr, err := net.ParseCIDR(cidrStr)
					if err != nil {
						return nil, c.Errf("invalid CIDR '%s': %v", cidrStr, err)
					}
					cidrs = append(cidrs, cidr)
				}

				rf.Rules = append(rf.Rules, Rule{
					Domain: domain,
					Blocks: cidrs,
				})
			default:
				return nil, c.Errf("unknown property '%s'", c.Val())
			}
		}
	}

	return rf, nil
}
