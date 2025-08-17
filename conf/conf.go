package conf

import "github.com/1f349/verbena/internal/utils"

type Conf struct {
	Listen        string             `yaml:"listen"`
	DB            string             `yaml:"db"`
	Nameservers   []string           `yaml:"nameservers"`
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
