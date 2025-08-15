package bind

import (
	"bytes"
	_ "embed"
	"testing"
)

//go:embed named.conf.local.generated
var namedConfLocalGenerated string

func TestWriteBindConfig(t *testing.T) {
	buf := new(bytes.Buffer)
	err := WriteBindConfig(buf, "/etc/bind/zones", []string{"example.com", "example.org", "example.net"})
	if err != nil {
		t.Fatal(err)
	}

	if buf.String() != namedConfLocalGenerated {
		t.Fatal("expected", namedConfLocalGenerated, "actual", buf.String())
	}
}
