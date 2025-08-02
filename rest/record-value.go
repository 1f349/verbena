package rest

import (
	"errors"
	"fmt"
	"github.com/1f349/verbena/internal/utils"
	"github.com/1f349/verbena/internal/zone"
	"net/netip"
	"strconv"
	"strings"
)

type RecordValue struct {
	Text       string      `json:"text,omitempty"`
	Target     string      `json:"target,omitempty"`
	IP         *netip.Addr `json:"ip,omitempty"`
	Preference int32       `json:"preference,omitempty"`
	Priority   int32       `json:"priority,omitempty"`
	Weight     int32       `json:"weight,omitempty"`
	Port       uint16      `json:"port,omitempty"`
	Flags      uint8       `json:"flags,omitempty"`
	Tag        string      `json:"tag,omitempty"`
	Value      string      `json:"value,omitempty"`
}

func (v RecordValue) IsValidForType(recordType string) bool {
	ty := zone.RecordTypeFromString(recordType)
	switch ty {
	case zone.NS:
		return utils.ValidateDomainName(v.Target)
	case zone.MX:
		return v.Preference > 0 && utils.ValidateDomainName(v.Target)
	case zone.A:
		return v.IP.Is4()
	case zone.AAAA:
		return v.IP.Is6()
	case zone.CNAME:
		return utils.ValidateDomainName(v.Target)
	case zone.TXT:
		return v.Text != ""
	case zone.SRV:
		return v.Priority > 0 && v.Weight >= 0 && v.Port > 0 && utils.ValidateDomainName(v.Target)
	case zone.CAA:
		return (v.Tag == "issue" || v.Tag == "issuewild") && !strings.ContainsAny(v.Value, "\t")
	default:
		return false
	}
}

func (v RecordValue) ToValueString(recordType string) string {
	ty := zone.RecordTypeFromString(recordType)
	switch ty {
	case zone.NS:
		return v.Target
	case zone.MX:
		return fmt.Sprintf("%d %s", v.Preference, v.Target)
	case zone.A:
		return v.IP.String()
	case zone.AAAA:
		return v.IP.String()
	case zone.CNAME:
		return v.Target
	case zone.TXT:
		return v.Text
	case zone.SRV:
		return fmt.Sprintf("%d %d %d %s", v.Priority, v.Weight, v.Port, v.Target)
	case zone.CAA:
		return fmt.Sprintf("%d %s %s", v.Flags, v.Tag, v.Value)
	default:
		return ""
	}
}

func ParseRecordValue(recordType string, value string) (RecordValue, error) {
	ty := zone.RecordTypeFromString(recordType)
	switch ty {
	case zone.NS:
		if !utils.ValidateDomainName(value) {
			return RecordValue{}, errors.New("invalid NS record")
		}
		return RecordValue{Target: value}, nil
	case zone.MX:
		fields := strings.SplitN(value, "\t", 3)
		if len(fields) != 2 {
			return RecordValue{}, errors.New("invalid MX record")
		}
		preference, err := strconv.ParseUint(fields[0], 10, 32)
		if err != nil {
			return RecordValue{}, errors.New("invalid MX record")
		}
		if !utils.ValidateDomainName(fields[1]) {
			return RecordValue{}, errors.New("invalid MX record")
		}
		return RecordValue{
			Preference: int32(preference),
			Target:     fields[1],
		}, nil
	case zone.A:
		v4, err := netip.ParseAddr(value)
		if err != nil || !v4.Is4() {
			return RecordValue{}, errors.New("invalid A record")
		}
		return RecordValue{IP: &v4}, nil
	case zone.AAAA:
		v6, err := netip.ParseAddr(value)
		if err != nil || !v6.Is6() {
			return RecordValue{}, errors.New("invalid AAAA record")
		}
		return RecordValue{IP: &v6}, nil
	case zone.CNAME:
		if !utils.ValidateDomainName(value) {
			return RecordValue{}, errors.New("invalid CNAME record")
		}
		return RecordValue{Target: value}, nil
	case zone.TXT:
		return RecordValue{Text: value}, nil
	case zone.SRV:
		fields := strings.SplitN(value, "\t", 5)
		if len(fields) != 4 {
			return RecordValue{}, errors.New("invalid SRV record")
		}
		// Priority
		priority, err := strconv.ParseUint(fields[0], 10, 32)
		if err != nil {
			return RecordValue{}, errors.New("invalid SRV record")
		}
		// Weight
		weight, err := strconv.ParseUint(fields[1], 10, 32)
		if err != nil {
			return RecordValue{}, errors.New("invalid SRV record")
		}
		// Port
		port, err := strconv.ParseUint(fields[2], 10, 16)
		if err != nil {
			return RecordValue{}, errors.New("invalid SRV record")
		}
		// Target
		if !utils.ValidateDomainName(fields[3]) {
			return RecordValue{}, errors.New("invalid SRV record")
		}
		return RecordValue{
			Priority: int32(priority),
			Weight:   int32(weight),
			Port:     uint16(port),
			Target:   fields[3],
		}, nil
	case zone.CAA:
		fields := strings.SplitN(value, "\t", 4)
		if len(fields) != 3 {
			return RecordValue{}, errors.New("invalid CAA record")
		}
		flags, err := strconv.ParseUint(fields[0], 10, 8)
		if err != nil {
			return RecordValue{}, errors.New("invalid CAA record")
		}
		if fields[1] != "issue" && fields[1] != "issuewild" {
			return RecordValue{}, errors.New("invalid CAA record")
		}
		return RecordValue{
			Flags: uint8(flags),
			Tag:   fields[1],
			Value: fields[2],
		}, nil
	default:
		return RecordValue{}, errors.New("invalid record type")
	}
}
