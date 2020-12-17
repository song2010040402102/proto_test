package config

import (
	"github.com/astaxie/beego/logs"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Cfg struct {
	Server   *ServerCfg `yaml:"server"`
	Bridge   *BridgeCfg `yaml:"bridge"`
	Protocol string     `yaml:"protocol"`
}

type ServerCfg struct {
	Listen string `yaml:"listen"`
}

type BridgeCfg struct {
	Listen string `yaml:"listen"`
	Web    string `yaml:"web"`
}

func init() {
	buf, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		logs.Error("Config read error:", err)
		return
	}
	err = yaml.Unmarshal(buf, &g_cfg)
	if err != nil {
		logs.Error("Config parse error:", err)
		return
	}
}

func Get() *Cfg {
	return g_cfg
}

var g_cfg *Cfg
