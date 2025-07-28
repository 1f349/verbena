package utils

import (
	"github.com/miekg/dns"
	"regexp"
)

var validateDnsName = regexp.MustCompile("^[a-z0-9-.]+$")

func ValidateDomainName(domain string) bool {
	_, isDomainName := dns.IsDomainName(domain)
	if !isDomainName {
		return false
	}
	return validateDnsName.MatchString(domain)
}
