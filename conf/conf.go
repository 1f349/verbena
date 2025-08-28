package conf

import (
	"fmt"
	"slices"

	"github.com/1f349/verbena/internal/database"
	"github.com/1f349/verbena/internal/utils"
	"gopkg.in/yaml.v3"
)

type Conf struct {
	Listen        string             `yaml:"listen"`
	DB            string             `yaml:"db"`
	Nameservers   NameserverConf     `yaml:"nameservers"`
	ZonePath      string             `yaml:"zonePath"`
	BindGenConf   string             `yaml:"bindGenConf"`
	GeneratorTick utils.DurationText `yaml:"generatorTick"`
	Primary       bool               `yaml:"primary"`
	CommitterTick utils.DurationText `yaml:"committerTick"`
	TokenIssuer   string             `yaml:"tokenIssuer"`
	Cmd           CmdConf            `yaml:"cmd"`
}

type CmdConf struct {
	Rndc      string `yaml:"rndc"`
	CheckConf string `yaml:"checkconf"`
	CheckZone string `yaml:"checkzone"`
}

func (c *CmdConf) LoadDefaults() {
	if c.Rndc == "" {
		c.Rndc = "/usr/sbin/rndc"
	}
	if c.CheckConf == "" {
		c.CheckConf = "/usr/bin/named-checkconf"
	}
	if c.CheckZone == "" {
		c.CheckZone = "/usr/bin/named-checkzone"
	}
}

type NameserverConf struct {
	nameserverMap      map[string][]string
	defaultNameservers []string
}

func NewNameserverConf(slice [][]string) (NameserverConf, error) {
	n := NameserverConf{
		nameserverMap: make(map[string][]string),
	}
	if len(slice) < 1 {
		return NameserverConf{}, fmt.Errorf("at least one nameserver slice must be specified")
	}
	for _, i := range slice {
		if len(i) < 2 {
			return NameserverConf{}, fmt.Errorf("a nameserver slice requires at least 2 nameservers")
		}
		n.nameserverMap[i[0]] = i
	}
	// The length of the default nameserver slice is checked in the above loop
	n.defaultNameservers = slice[0]
	return n, nil
}

func MustNameserverConf(slice [][]string) NameserverConf {
	n, err := NewNameserverConf(slice)
	if err != nil {
		panic(err)
	}
	return n
}

var _ yaml.Marshaler = (*NameserverConf)(nil)
var _ yaml.Unmarshaler = (*NameserverConf)(nil)

func (n NameserverConf) MarshalYAML() (any, error) {
	var slice [][]string
	slice = append(slice, n.defaultNameservers)
	for _, i := range n.nameserverMap {
		if slices.Equal(i, n.defaultNameservers) {
			continue
		}
		slice = append(slice, i)
	}
	return yaml.Marshal(slice)
}

func (n *NameserverConf) UnmarshalYAML(bytes *yaml.Node) error {
	var slice [][]string
	err := bytes.Decode(&slice)
	if err != nil {
		return err
	}
	n2, err := NewNameserverConf(slice)
	if err != nil {
		return err
	}
	*n = n2
	return nil
}

func (n *NameserverConf) GetNameserversForZone(zoneInfo database.Zone) []string {
	slice := n.nameserverMap[zoneInfo.Nameserver]
	if slice != nil {
		return slice
	}
	return n.defaultNameservers
}
