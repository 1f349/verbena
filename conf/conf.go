package conf

import "github.com/1f349/verbena/internal/utils"

type Conf struct {
	Listen        string             `yaml:"listen"`
	DB            string             `yaml:"db"`
	Nameservers   []string           `yaml:"nameservers"`
	ZonePath      string             `yaml:"zonePath"`
	GeneratorTick utils.DurationText `yaml:"generatorTick"`
	Primary       bool               `yaml:"primary"`
	CommitterTick utils.DurationText `yaml:"committerTick"`
	TokenIssuer   string             `yaml:"tokenIssuer"`
}
