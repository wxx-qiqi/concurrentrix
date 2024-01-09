package config

import (
	"fmt"
	"github.com/go-ini/ini"
)

var (
	Work    = 1
	ChanNum = 10
	Again   = 2
)

func init() {
	// 读取INI配置文件
	cfg, err := ini.Load("config/config.ini")
	if err != nil {
		fmt.Errorf("ini error config: %v", err)
		return
	}

	// 读取配置项
	section := cfg.Section("work")
	if section == nil {
		fmt.Errorf("ini error work")
		return
	}

	Work, _ = section.Key("work").Int()
	ChanNum, _ = section.Key("chan").Int()
	Again, _ = section.Key("again").Int()
}
