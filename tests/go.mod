module tests

go 1.21.1

require (
	github.com/roadrunner-server/config/v4 v4.4.5
	github.com/roadrunner-server/endure/v2 v2.4.2
	github.com/roadrunner-server/goridge/v3 v3.8.1
	github.com/roadrunner-server/resetter/v4 v4.0.10
	github.com/roadrunner-server/rpc/v4 v4.2.4
	github.com/roadrunner-server/sdk/v4 v4.5.1
	github.com/roadrunner-server/server/v4 v4.4.1
	github.com/stretchr/testify v1.8.4
	go.uber.org/zap v1.26.0
)

replace (
	github.com/roadrunner-server/resetter/v4 => ../
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/google/uuid v1.3.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml/v2 v2.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/roadrunner-server/errors v1.3.0 // indirect
	github.com/roadrunner-server/tcplisten v1.4.0 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/spf13/afero v1.10.0 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.16.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sync v0.3.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)