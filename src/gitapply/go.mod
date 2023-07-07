module gitapply

go 1.19

replace (
	glb_config => ./glb_config
	init_sub => ./init_sub
	registration => ./registration
	sub_dns => ./sub_dns
	utils => ./utils
	zabbix_tools => ./zabbix_tools
)

require (
	init_sub v0.0.0-00010101000000-000000000000
	registration v0.0.0-00010101000000-000000000000
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-sql-driver/mysql v1.7.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml/v2 v2.0.6 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/spf13/afero v1.9.3 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.15.0 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	github.com/urfave/cli/v2 v2.25.3 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	glb_config v0.0.0-00010101000000-000000000000 // indirect
	golang.org/x/sys v0.3.0 // indirect
	golang.org/x/text v0.5.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	sub_dns v0.0.0-00010101000000-000000000000 // indirect
	utils v0.0.0-00010101000000-000000000000 // indirect
	zabbix_tools v0.0.0-00010101000000-000000000000 // indirect
)
