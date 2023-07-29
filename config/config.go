package config

import (
	"log"

	"gopkg.in/ini.v1"
)

var (
	//修改了 config.ini 文件后，需要同步修改此处和 assets/config.ini 文件里的 version
	CONFIG_VERSION = "1.4.0"
)

func GetConfig() *ini.File {
	cfg, err := ini.Load("config.ini")
	ini.PrettyFormat = false
	if err != nil {
		log.Fatalln("无法读取 config.ini 文件，请检查文件是否存在。错误信息:", err.Error())
	}
	if cfg.Section("config").Key("version").String() != CONFIG_VERSION {
		UpdateConfig(cfg)
	}
	return cfg
}

func UpdateConfig(cfg *ini.File) {
	var cfgVersion = cfg.Section("config").Key("version").String()
	switch cfgVersion {
	case "1.2.0":
		//From 1.2.0 to 1.4.0
		//Change sth
		cfg.Section("config").Key("version").SetValue("1.4.0")
		cfg.SaveTo("config.ini")
	default:
		//
	}
}
