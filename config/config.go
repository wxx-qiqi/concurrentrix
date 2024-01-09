package config

import (
	"fmt"

	"github.com/go-ini/ini"
)

var (
	Work      = 0
	ChanNum   = 0
	SendTimes = 0
)

func init() {
	// Read ini configuration file
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
	SendTimes, _ = section.Key("send_times").Int()
}
