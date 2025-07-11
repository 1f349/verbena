package zone

import (
	"bytes"
	_ "embed"
	"encoding/hex"
	"github.com/gobuffalo/nulls"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

const exampleDomainKey = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" +
	"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" +
	"cccccccccccccccccccccccccccccccccccccccccccccccccc" +
	"dddddddddddddddddddddddddddddddddddddddddddddddddd" +
	"eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee" +
	"ffffffffffffffffffffffffffffffffffffffffffffffffff" +
	"gggggggggggggggggggggggggggggggggggggggggggggggggg" +
	"hhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhhh" +
	"iiiiiiiiii"

//go:embed test.zone
var testZoneFile string

func TestWriteZone(t *testing.T) {
	var nu32 nulls.UInt32

	buf := new(bytes.Buffer)
	err := WriteZone(buf, "example.com", 86400, SoaRecord{
		Nameserver: "dns1.example.com",
		Admin:      "hostmaster.example.com",
		Serial:     2001062501,
		Refresh:    21600,
		Retry:      3600,
		Expire:     604800,
		TimeToLive: 86400,
	}, []Record{
		{"@", nu32, NS, "dns1.example.com"},
		{"@", nu32, NS, "dns2.example.com"},
		{"@", nu32, MX, "10 mail.example.com"},
		{"@", nu32, MX, "20 mail2.example.com"},
		{"dns1", nu32, A, "10.0.1.1"},
		{"dns1", nu32, AAAA, "2001:db8::1:1"},
		{"dns2", nu32, A, "10.0.1.2"},
		{"dns2", nu32, AAAA, "2001:db8::1:2"},
		{"server1", nu32, A, "10.0.1.5"},
		{"server1", nu32, AAAA, "2001:db8::1:5"},
		{"server2", nu32, A, "10.0.1.6"},
		{"server2", nu32, AAAA, "2001:db8::1:6"},
		{"ftp", nu32, A, "10.0.1.3"},
		{"ftp", nu32, AAAA, "2001:db8::1:3"},
		{"", nu32, A, "10.0.1.4"},
		{"", nu32, AAAA, "2001:db8::1:4"},
		{"mail", nu32, A, "10.0.2.1"},
		{"mail", nu32, AAAA, "2001:db8::2:1"},
		{"mail2", nu32, A, "10.0.2.2"},
		{"mail2", nu32, AAAA, "2001:db8::2:2"},
		{"www", nu32, CNAME, "server1.example.com"},
		{"sixinfour", nu32, A, "10.0.6.4"},
		{"sixinfour", nu32, AAAA, "64:ff9b::10.0.6.4"},
		{"", nu32, TXT, "google-site-verification=zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"},
		{"", nu32, TXT, "v=spf1 include:_spf.example.com -all"},
		{"*", nu32, TXT, "v=spf1 include:_spf.example.com -all"},
		{"_dmarc", nu32, TXT, "v=DMARC1; p=quarantine; sp=quarantine; pct=50; rua=mailto:dmarcreports@example.com; ruf=mailto:dmarcfailurereports@example.com; adkim=r; aspf=r;"},
		{"mail._domainkey", nu32, TXT, exampleDomainKey},
	})
	if err != nil {
		t.Fatal(err)
	}

	genLines := slices.Collect(strings.Lines(buf.String()))
	lineIdx := 0

	for expectedLine := range strings.Lines(testZoneFile) {
		expectedLine, _, _ = strings.Cut(expectedLine, ";")
		expectedLine = strings.TrimSpace(expectedLine)
		if expectedLine == "" {
			continue
		}

		line, _, _ := strings.Cut(genLines[lineIdx], ";")
		line = strings.TrimSpace(line)

		if expectedLine != line {
			t.Log("expected = ", hex.EncodeToString([]byte(expectedLine)))
			t.Log("actual = ", hex.EncodeToString([]byte(line)))
			t.Fatal("expected", expectedLine, "actual", line)
		}
		lineIdx++
	}

	checkWithBindCheckZone(t, buf.Bytes(), "example.com")
}

func checkWithBindCheckZone(t *testing.T, data []byte, zoneName string) {
	tempFile, err := os.CreateTemp("", "verbena-test-*.zone")
	if err != nil {
		t.Fatal(err)
	}
	defer tempFile.Close()
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write(data)
	if err != nil {
		t.Fatal(err)
	}

	err = tempFile.Close()
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("/usr/bin/named-checkzone", zoneName, tempFile.Name())
	err = cmd.Run()
	if err != nil {
		t.Logf("exit code = %d", cmd.ProcessState.ExitCode())
		t.Fatal(err)
	}
}
