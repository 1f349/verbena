package conf

type Conf struct {
	Listen   string `yaml:"listen"`
	DB       string `yaml:"db"`
	ZonePath string `yaml:"zonePath"`
}
