package zone

import (
	"net/netip"
	"testing"

	"github.com/gobuffalo/nulls"
	"github.com/stretchr/testify/assert"
)

func TestRfc2317CNAMEs(t *testing.T) {
	const (
		v4base4  = ".240/4.in-addr.arpa"
		v4base6  = ".240/6.in-addr.arpa"
		v4base12 = ".160/12.240.in-addr.arpa"
		v4base14 = ".160/14.240.in-addr.arpa"
		v4base20 = ".192/20.160.240.in-addr.arpa"
		v4base22 = ".192/22.160.240.in-addr.arpa"
		v4base28 = ".128/28.192.160.240.in-addr.arpa"
		v4base30 = ".128/30.192.160.240.in-addr.arpa"
		v6base22 = ".8/22.0.f.f.f.3.ip6.arpa"
		v6base23 = ".8/23.0.f.f.f.3.ip6.arpa"
	)

	tests := []struct {
		prefix netip.Prefix
		stubs  []Record
	}{
		{netip.MustParsePrefix("240.0.0.0/8"), nil},
		{netip.MustParsePrefix("240.0.0.0/16"), nil},
		{netip.MustParsePrefix("240.0.0.0/24"), nil},
		{netip.MustParsePrefix("240.0.0.0/32"), nil},
		{netip.MustParsePrefix("240.0.0.0/4"), []Record{
			{Name: "240.in-addr.arpa", Type: CNAME, Value: "240" + v4base4},
			{Name: "241.in-addr.arpa", Type: CNAME, Value: "241" + v4base4},
			{Name: "242.in-addr.arpa", Type: CNAME, Value: "242" + v4base4},
			{Name: "243.in-addr.arpa", Type: CNAME, Value: "243" + v4base4},
			{Name: "244.in-addr.arpa", Type: CNAME, Value: "244" + v4base4},
			{Name: "245.in-addr.arpa", Type: CNAME, Value: "245" + v4base4},
			{Name: "246.in-addr.arpa", Type: CNAME, Value: "246" + v4base4},
			{Name: "247.in-addr.arpa", Type: CNAME, Value: "247" + v4base4},
			{Name: "248.in-addr.arpa", Type: CNAME, Value: "248" + v4base4},
			{Name: "249.in-addr.arpa", Type: CNAME, Value: "249" + v4base4},
			{Name: "250.in-addr.arpa", Type: CNAME, Value: "250" + v4base4},
			{Name: "251.in-addr.arpa", Type: CNAME, Value: "251" + v4base4},
			{Name: "252.in-addr.arpa", Type: CNAME, Value: "252" + v4base4},
			{Name: "253.in-addr.arpa", Type: CNAME, Value: "253" + v4base4},
			{Name: "254.in-addr.arpa", Type: CNAME, Value: "254" + v4base4},
			{Name: "255.in-addr.arpa", Type: CNAME, Value: "255" + v4base4},
		}},
		{netip.MustParsePrefix("240.0.0.0/6"), []Record{
			{Name: "240.in-addr.arpa", Type: CNAME, Value: "240" + v4base6},
			{Name: "241.in-addr.arpa", Type: CNAME, Value: "241" + v4base6},
			{Name: "242.in-addr.arpa", Type: CNAME, Value: "242" + v4base6},
			{Name: "243.in-addr.arpa", Type: CNAME, Value: "243" + v4base6},
		}},
		{netip.MustParsePrefix("240.160.0.0/12"), []Record{
			{Name: "160.240.in-addr.arpa", Type: CNAME, Value: "160" + v4base12},
			{Name: "161.240.in-addr.arpa", Type: CNAME, Value: "161" + v4base12},
			{Name: "162.240.in-addr.arpa", Type: CNAME, Value: "162" + v4base12},
			{Name: "163.240.in-addr.arpa", Type: CNAME, Value: "163" + v4base12},
			{Name: "164.240.in-addr.arpa", Type: CNAME, Value: "164" + v4base12},
			{Name: "165.240.in-addr.arpa", Type: CNAME, Value: "165" + v4base12},
			{Name: "166.240.in-addr.arpa", Type: CNAME, Value: "166" + v4base12},
			{Name: "167.240.in-addr.arpa", Type: CNAME, Value: "167" + v4base12},
			{Name: "168.240.in-addr.arpa", Type: CNAME, Value: "168" + v4base12},
			{Name: "169.240.in-addr.arpa", Type: CNAME, Value: "169" + v4base12},
			{Name: "170.240.in-addr.arpa", Type: CNAME, Value: "170" + v4base12},
			{Name: "171.240.in-addr.arpa", Type: CNAME, Value: "171" + v4base12},
			{Name: "172.240.in-addr.arpa", Type: CNAME, Value: "172" + v4base12},
			{Name: "173.240.in-addr.arpa", Type: CNAME, Value: "173" + v4base12},
			{Name: "174.240.in-addr.arpa", Type: CNAME, Value: "174" + v4base12},
			{Name: "175.240.in-addr.arpa", Type: CNAME, Value: "175" + v4base12},
		}},
		{netip.MustParsePrefix("240.160.0.0/14"), []Record{
			{Name: "160.240.in-addr.arpa", Type: CNAME, Value: "160" + v4base14},
			{Name: "161.240.in-addr.arpa", Type: CNAME, Value: "161" + v4base14},
			{Name: "162.240.in-addr.arpa", Type: CNAME, Value: "162" + v4base14},
			{Name: "163.240.in-addr.arpa", Type: CNAME, Value: "163" + v4base14},
		}},
		{netip.MustParsePrefix("240.160.192.0/20"), []Record{
			{Name: "192.160.240.in-addr.arpa", Type: CNAME, Value: "192" + v4base20},
			{Name: "193.160.240.in-addr.arpa", Type: CNAME, Value: "193" + v4base20},
			{Name: "194.160.240.in-addr.arpa", Type: CNAME, Value: "194" + v4base20},
			{Name: "195.160.240.in-addr.arpa", Type: CNAME, Value: "195" + v4base20},
			{Name: "196.160.240.in-addr.arpa", Type: CNAME, Value: "196" + v4base20},
			{Name: "197.160.240.in-addr.arpa", Type: CNAME, Value: "197" + v4base20},
			{Name: "198.160.240.in-addr.arpa", Type: CNAME, Value: "198" + v4base20},
			{Name: "199.160.240.in-addr.arpa", Type: CNAME, Value: "199" + v4base20},
			{Name: "200.160.240.in-addr.arpa", Type: CNAME, Value: "200" + v4base20},
			{Name: "201.160.240.in-addr.arpa", Type: CNAME, Value: "201" + v4base20},
			{Name: "202.160.240.in-addr.arpa", Type: CNAME, Value: "202" + v4base20},
			{Name: "203.160.240.in-addr.arpa", Type: CNAME, Value: "203" + v4base20},
			{Name: "204.160.240.in-addr.arpa", Type: CNAME, Value: "204" + v4base20},
			{Name: "205.160.240.in-addr.arpa", Type: CNAME, Value: "205" + v4base20},
			{Name: "206.160.240.in-addr.arpa", Type: CNAME, Value: "206" + v4base20},
			{Name: "207.160.240.in-addr.arpa", Type: CNAME, Value: "207" + v4base20},
		}},
		{netip.MustParsePrefix("240.160.192.0/22"), []Record{
			{Name: "192.160.240.in-addr.arpa", Type: CNAME, Value: "192" + v4base22},
			{Name: "193.160.240.in-addr.arpa", Type: CNAME, Value: "193" + v4base22},
			{Name: "194.160.240.in-addr.arpa", Type: CNAME, Value: "194" + v4base22},
			{Name: "195.160.240.in-addr.arpa", Type: CNAME, Value: "195" + v4base22},
		}},
		{netip.MustParsePrefix("240.160.192.128/28"), []Record{
			{Name: "128.192.160.240.in-addr.arpa", Type: CNAME, Value: "128" + v4base28},
			{Name: "129.192.160.240.in-addr.arpa", Type: CNAME, Value: "129" + v4base28},
			{Name: "130.192.160.240.in-addr.arpa", Type: CNAME, Value: "130" + v4base28},
			{Name: "131.192.160.240.in-addr.arpa", Type: CNAME, Value: "131" + v4base28},
			{Name: "132.192.160.240.in-addr.arpa", Type: CNAME, Value: "132" + v4base28},
			{Name: "133.192.160.240.in-addr.arpa", Type: CNAME, Value: "133" + v4base28},
			{Name: "134.192.160.240.in-addr.arpa", Type: CNAME, Value: "134" + v4base28},
			{Name: "135.192.160.240.in-addr.arpa", Type: CNAME, Value: "135" + v4base28},
			{Name: "136.192.160.240.in-addr.arpa", Type: CNAME, Value: "136" + v4base28},
			{Name: "137.192.160.240.in-addr.arpa", Type: CNAME, Value: "137" + v4base28},
			{Name: "138.192.160.240.in-addr.arpa", Type: CNAME, Value: "138" + v4base28},
			{Name: "139.192.160.240.in-addr.arpa", Type: CNAME, Value: "139" + v4base28},
			{Name: "140.192.160.240.in-addr.arpa", Type: CNAME, Value: "140" + v4base28},
			{Name: "141.192.160.240.in-addr.arpa", Type: CNAME, Value: "141" + v4base28},
			{Name: "142.192.160.240.in-addr.arpa", Type: CNAME, Value: "142" + v4base28},
			{Name: "143.192.160.240.in-addr.arpa", Type: CNAME, Value: "143" + v4base28},
		}},
		{netip.MustParsePrefix("240.160.192.128/30"), []Record{
			{Name: "128.192.160.240.in-addr.arpa", Type: CNAME, Value: "128" + v4base30},
			{Name: "129.192.160.240.in-addr.arpa", Type: CNAME, Value: "129" + v4base30},
			{Name: "130.192.160.240.in-addr.arpa", Type: CNAME, Value: "130" + v4base30},
			{Name: "131.192.160.240.in-addr.arpa", Type: CNAME, Value: "131" + v4base30},
		}},
		{netip.MustParsePrefix("3fff:800::/20"), nil},
		{netip.MustParsePrefix("3fff:800::/24"), nil},
		{netip.MustParsePrefix("3fff:800::/22"), []Record{
			{Name: "8.0.f.f.f.3.ip6.arpa", Type: CNAME, Value: "8" + v6base22},
			{Name: "9.0.f.f.f.3.ip6.arpa", Type: CNAME, Value: "9" + v6base22},
			{Name: "a.0.f.f.f.3.ip6.arpa", Type: CNAME, Value: "a" + v6base22},
			{Name: "b.0.f.f.f.3.ip6.arpa", Type: CNAME, Value: "b" + v6base22},
		}},
		{netip.MustParsePrefix("3fff:800::/23"), []Record{
			{Name: "8.0.f.f.f.3.ip6.arpa", Type: CNAME, Value: "8" + v6base23},
			{Name: "9.0.f.f.f.3.ip6.arpa", Type: CNAME, Value: "9" + v6base23},
		}},
	}
	for _, i := range tests {
		t.Run(i.prefix.String(), func(t *testing.T) {
			assert.EqualValues(t, i.stubs, Rfc2317CNAMEs(i.prefix, nulls.UInt32{}))
		})
	}
}
