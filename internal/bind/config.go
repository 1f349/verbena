package bind

import (
	"fmt"
	"io"
	"path"
	"strconv"
)

func WriteBindConfig(w io.Writer, zonesPath string, origins []string) error {
	for _, zone := range origins {
		// zone "example.com" IN {
		// <tab>type master;
		// <tab>file "/etc/bind/zones/example.com.zone";
		// };
		_, err := fmt.Fprintf(w, "zone %s IN {\n", strconv.Quote(zone))
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, "\ttype master;\n")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, "\tfile %s;\n", strconv.Quote(path.Join(zonesPath, zone+".zone")))
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(w, "};\n")
		if err != nil {
			return err
		}
	}
	return nil
}
