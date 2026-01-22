package responsefilter

import (
	"context"
	"net"

	"github.com/coredns/coredns/plugin"
	"github.com/miekg/dns"
)

type ResponseFilter struct {
	Next  plugin.Handler
	Rules []Rule
}

type Rule struct {
	Domain string
	Blocks []*net.IPNet
}

func (rf *ResponseFilter) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	// Create a response writer wrapper to intercept the response
	rw := &responseWriter{ResponseWriter: w, rf: rf, request: r}
	return plugin.NextOrFailure(rf.Name(), rf.Next, ctx, rw, r)
}

func (rf *ResponseFilter) Name() string { return "responsefilter" }

type responseWriter struct {
	dns.ResponseWriter
	rf      *ResponseFilter
	request *dns.Msg
}

func (rw *responseWriter) WriteMsg(res *dns.Msg) error {
	// Check if response should be blocked
	if rw.rf.shouldBlock(res) {
		// Return REFUSED instead of the actual response
		refused := new(dns.Msg)
		refused.SetRcode(rw.request, dns.RcodeRefused)
		return rw.ResponseWriter.WriteMsg(refused)
	}
	return rw.ResponseWriter.WriteMsg(res)
}

func (rf *ResponseFilter) shouldBlock(res *dns.Msg) bool {
	if res == nil || len(res.Answer) == 0 {
		return false
	}

	// Extract query name
	qname := ""
	if len(res.Question) > 0 {
		qname = res.Question[0].Name
	}

	// Check each answer record
	for _, rr := range res.Answer {
		switch record := rr.(type) {
		case *dns.A:
			if rf.isBlocked(record.Hdr.Name, record.A) {
				return true
			}
		case *dns.AAAA:
			if rf.isBlocked(record.Hdr.Name, record.AAAA) {
				return true
			}
		case *dns.CNAME:
			// For CNAME, check using the original query name
			if qname != "" {
				// Continue checking - CNAME itself doesn't have an IP
				continue
			}
		}
	}

	return false
}

func (rf *ResponseFilter) isBlocked(fqdn string, ip net.IP) bool {
	for _, rule := range rf.Rules {
		if dns.IsSubDomain(rule.Domain, fqdn) {
			for _, cidr := range rule.Blocks {
				if cidr.Contains(ip) {
					return true
				}
			}
		}
	}
	return false
}
