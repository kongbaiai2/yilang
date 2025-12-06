package config

type System struct {
	Env   string `yaml:"env"`
	Debug bool   `yaml:"debug"`
	// Auth      bool   `yaml:"auth"`
	HttpPort  string `yaml:"httpPort"`
	PprofPort string `yaml:"pprofPort"`
	GinMode   string `yaml:"ginMode"`
	IpAddress string `yaml:"ipAddress"`
}

type Mysql struct {
	Path         string `yaml:"path"`
	MaxIdleConns int    `yaml:"maxIdleConns"`
	MaxOpenConns int    `yaml:"maxOpenConns"`
	Debug        bool   `yaml:"logMode"`
	LogZap       string `yaml:"logZap"`
}

type Redis struct {
	Addr     string
	DB       int
	Password string
}

type Zap struct {
	Level       string `yaml:"level"`
	Path        string `yaml:"path"`
	PathDb      string `yaml:"pathDb"`
	Format      string `yaml:"format"`
	Prefix      string `yaml:"prefix"`
	EncodeLevel string `yaml:"encodeLevel"`
}

/*
Filename : 日志文件的位置；
MaxSize ：在进行切割之前，日志文件的最大大小（以MB为单位）；
MaxBackups ：保留旧文件的最大个数；
MaxAges ：保留旧文件的最大天数；
Compress ：是否压缩/归档旧文件；
*/
type Logrotate struct {
	MaxSize    int  `yaml:"maxSize"`
	MaxBackups int  `yaml:"maxBackups"`
	MaxAges    int  `yaml:"maxAges"`
	Compress   bool `yaml:"compress"`
}

type CactiCfg struct {
	BaseURL  string `yaml:"baseURL"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	ImgPath  string `yaml:"imgPath"`
}
