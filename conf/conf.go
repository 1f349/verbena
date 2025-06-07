package conf

type Conf struct {
	Listen   ListenConf `yaml:"listen"`
	DB       string     `yaml:"db"`
	ZonePath string     `yaml:"zonePath"`
}

type ListenConf struct {
	Http string `yaml:"http"`
	Api  string `yaml:"api"`
}
