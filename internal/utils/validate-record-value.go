package utils

import (
	"net/netip"
	"strconv"
	"strings"
)

func ValidateRecordValue(ty string, value string) bool {
	switch ty {
	case "SOA":
		return false
	case "NS":
		return ValidateDomainName(value)
	case "MX":
		fields := strings.SplitN(value, "\t", 3)
		if len(fields) != 2 {
			return false
		}
		if _, err := strconv.ParseUint(fields[0], 10, 32); err != nil {
			return false
		}
		return ValidateDomainName(fields[1])
	case "A":
		addr, err := netip.ParseAddr(value)
		return err == nil && addr.Is4()
	case "AAAA":
		addr, err := netip.ParseAddr(value)
		return err == nil && addr.Is6()
	case "CNAME":
		return ValidateDomainName(value)
	case "TXT":
		return true
	case "SRV":
		fields := strings.SplitN(value, "\t", 5)
		if len(fields) != 4 {
			return false
		}
		// Priority
		if _, err := strconv.ParseUint(fields[0], 10, 32); err != nil {
			return false
		}
		// Weight
		if _, err := strconv.ParseUint(fields[1], 10, 32); err != nil {
			return false
		}
		// Port
		if _, err := strconv.ParseUint(fields[2], 10, 16); err != nil {
			return false
		}
		// Target
		return ValidateDomainName(fields[3])
	case "CAA":
		fields := strings.SplitN(value, "\t", 3)
		if len(fields) != 2 {
			return false
		}
		return fields[0] == "issue" || fields[1] == "issuewild"
	default:
		return false
	}
}
