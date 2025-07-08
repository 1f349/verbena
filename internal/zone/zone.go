package zone

import (
	"fmt"
	"github.com/miekg/dns"
	"io"
	"net/netip"
	"strconv"
	"strings"
)

func zoneRecordName(name string) string {
	if name == "" {
		return "@"
	}
	return name
}

type SoaRecord struct {
	Nameserver string
	Admin      string
	Serial     uint32
	Refresh    uint32
	Retry      uint32
	Expire     uint32
	TimeToLive uint32
}

func WriteZone(w io.Writer, origin string, defaultTtl uint32, soa SoaRecord, records []Record) error {
	_, ok := dns.IsDomainName(origin)
	if !ok {
		return fmt.Errorf("invalid zone origin: %s", origin)
	}

	_, err := fmt.Fprintf(w, "$ORIGIN %s\n", dns.Fqdn(origin))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "$TTL %d\n", defaultTtl)
	if err != nil {
		return err
	}

	// SOA record
	_, err = fmt.Fprintf(w, "@\tIN\tSOA\t%s\t%s (\n", dns.Fqdn(soa.Nameserver), dns.Fqdn(soa.Admin))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "\t\t\t%d ; Serial\n", soa.Serial)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "\t\t\t%d ; Refresh\n", soa.Refresh)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "\t\t\t%d ; Retry\n", soa.Retry)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "\t\t\t%d ; Expire\n", soa.Expire)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "\t\t\t%d ) ; Minimum TTL\n", soa.TimeToLive)
	if err != nil {
		return err
	}

	for _, record := range records {
		var val string
		// TODO: valid line parsing better
		switch record.Type {
		case NS:
			_, ok := dns.IsDomainName(record.Value)
			if !ok {
				return fmt.Errorf("invalid NS record: %s", record.Value)
			}
			val = dns.Fqdn(record.Value)
		case MX:
			mxFields := strings.Fields(record.Value)
			if len(mxFields) != 2 {
				return fmt.Errorf("invalid MX record: %s", record.Value)
			}
			mxPriority, err := strconv.ParseUint(mxFields[0], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid MX record: %s", record.Value)
			}
			_, ok := dns.IsDomainName(mxFields[1])
			if !ok {
				return fmt.Errorf("invalid MX record: the provided target is invalid: %s", mxFields[1])
			}
			val = fmt.Sprintf("%d\t%s", mxPriority, dns.Fqdn(mxFields[1]))
		case A:
			addr, err := netip.ParseAddr(record.Value)
			if err != nil {
				return fmt.Errorf("invalid A record %s: %w", record.Value, err)
			}
			if !addr.Is4() {
				return fmt.Errorf("invalid A record %s: this provided address is not IPv4", record.Value)
			}
			val = addr.String()
		case AAAA:
			addr, err := netip.ParseAddr(record.Value)
			if err != nil {
				return fmt.Errorf("invalid AAAA record %s: %w", record.Value, err)
			}
			if !addr.Is6() && !addr.Is4In6() {
				return fmt.Errorf("invalid AAAA record %s: this provided address is not IPv6", record.Value)
			}
			val = addr.String()
		case CNAME:
			_, ok := dns.IsDomainName(record.Value)
			if !ok {
				return fmt.Errorf("invalid CNAME record: %s", record.Value)
			}
			val = dns.Fqdn(record.Value)
		case TXT:
			if len(record.Value) <= 250 {
				// always quote for simplicity
				val = strconv.Quote(record.Value)
			} else {
				val = "(\n"
				l := len(record.Value)
				for i := 0; i < l; i += 100 {
					val += fmt.Sprintf("\t%s\n", strconv.Quote(record.Value[i:min(l, i+100)]))
				}
				val += ")"
			}
		}

		_, err = fmt.Fprintf(w, "%s\tIN\t%s\t%s\n", zoneRecordName(record.Name), record.Type.String(), val)
		if err != nil {
			return err
		}
	}

	return nil
}
