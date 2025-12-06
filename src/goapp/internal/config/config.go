package config

type Server struct {
	System    System    `yaml:"system"`
	Redis     Redis     `yaml:"redis"`
	Mysql     Mysql     `yaml:"mysql"`
	Zap       Zap       `yaml:"zap"`
	LogRotate Logrotate `yaml:"logrotate"`
	CactiCfg  CactiCfg  `yaml:"cactiCfg"`
}

type TaskFinishInfo struct {
	EipSwitchRsPoolFinish      bool
	EniSwitchRsPoolFinish      bool
	EniQueryIdFinish           bool
	PrivateDnsConfigTaskFinish bool
	PrivateDnsRegionTaskFinish bool
}
