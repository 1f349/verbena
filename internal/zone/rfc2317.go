package zone

import (
	"net/netip"
	"strconv"

	"github.com/gobuffalo/nulls"
)

type Rfc2317PrefixCollector struct {
	m map[netip.Prefix][]netip.Prefix
}

func NewRfc2317PrefixCollector() *Rfc2317PrefixCollector {
	return &Rfc2317PrefixCollector{m: make(map[netip.Prefix][]netip.Prefix)}
}

func (r *Rfc2317PrefixCollector) AddPrefix(p netip.Prefix) {
	bits := p.Bits()
	switch {
	case p.Addr().Is4():
		bits -= bits % 8
	case p.Addr().Is6():
		bits -= bits % 4
	}
	prefix, err := p.Addr().Prefix(bits)
	if err != nil {
		return
	}
	r.m[prefix] = append(r.m[prefix], p)
}

func (r *Rfc2317PrefixCollector) TopLevelPrefixes() []netip.Prefix {
	var prefixes []netip.Prefix
	for k := range r.m {
		prefixes = append(prefixes, k)
	}
	return prefixes
}

func (r *Rfc2317PrefixCollector) CNAMEs(topLevelPrefix netip.Prefix, ttl nulls.UInt32) []Record {
	prefixes := r.m[topLevelPrefix]

	switch {
	case topLevelPrefix.Addr().Is4():
		for _, prefix := range prefixes {
			prefix = prefix.Masked()
		}

	case topLevelPrefix.Addr().Is6():
	}

	// TODO: move this stuff up...

		bits := prefix.Bits()
		tet := 0
		switch prefix.Addr().BitLen() {
		case 0:
			return nil
		case 32:
			tet = 8 // octets (IPv4)
		case 128:
			tet = 4 // hextets (IPv6)
		}

		steps := bits % tet
		if steps == 0 {
			return nil
		}

		prefix = prefix.Masked()

		records := make([]Record, 1<<(tet-steps))

		slice := prefix.Addr().AsSlice()
		baseSegmentCount := bits / tet
		switch tet {
		case 8: // IPv4
			baseSegments := "in-addr.arpa"
			for i := 0; i < baseSegmentCount; i++ {
				baseSegments = strconv.Itoa(int(slice[i])) + "." + baseSegments
			}
			start := int(slice[baseSegmentCount])
			outSegments := strconv.Itoa(start) + "/" + strconv.Itoa(bits) + "." + baseSegments

			for i := 0; i < len(records); i++ {
				j := strconv.Itoa(start + i)
				records[i] = Record{
					Name:       j + "." + baseSegments,
					TimeToLive: ttl,
					Type:       CNAME,
					Value:      j + "." + outSegments,
				}
			}
		case 4: // IPv6
			baseSegments := "ip6.arpa"
			for i := 0; i < baseSegmentCount; i++ {
				baseSegments = nibbleHex(v6Nibble(slice, i)) + "." + baseSegments
			}
			start := v6Nibble(slice, baseSegmentCount)
			outSegments := nibbleHex(start) + "/" + strconv.Itoa(bits) + "." + baseSegments

			for i := 0; i < len(records); i++ {
				j := nibbleHex(start + byte(i))
				records[i] = Record{
					Name:       j + "." + baseSegments,
					TimeToLive: ttl,
					Type:       CNAME,
					Value:      j + "." + outSegments,
				}
			}
		}

		return records
	}
}

func v6Nibble(slice []byte, n int) byte {
	octet := slice[n/2]
	if n&1 == 0 {
		return octet >> 4
	} else {
		return octet & 0xf
	}
}

const nibbleHexTable = "0123456789abcdef"

func nibbleHex(b byte) string {
	return string(nibbleHexTable[b])
}
