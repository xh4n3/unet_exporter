package sdk

type Config struct {
	Global  *Global   `yaml:"global"`
	Targets []*Target `yaml:"targets"`
	Plugins *Plugins  `yaml:"plugins"`
}

type Global struct {
	ApiEndpoint string `yaml:"api_endpoint"`
	PublicKey   string `yaml:"public_key"`
	PrivateKey  string `yaml:"private_key"`
	MertricPort string `yaml:"mertric_port"`
	Interval    int    `yaml:"interval"`
	DryRun      bool   `yaml:"dry_run"`
	Verbose     bool   `yaml:"verbose"`
}

type Plugins struct {
	DownLimitAdvisor string `yaml:"down_limit_advisor"`
}

type Target struct {
	Name             string         `yaml:"name"`
	Region           string         `yaml:"region"`
	DefaultBandwidth int            `yaml:"default_bandwidth"`
	VariedLimits     []*VariedLimit `yaml:"varied_limits"`
}

type VariedLimit struct {
	Name      string `yaml:"name"`
	UpLimit   int    `yaml:"up_limit"`
	DownLimit int    `yaml:"down_limit"`
	UpStep    int    `yaml:"up_step"`
	DownStep  int    `yaml:"down_step"`
	WeekDays  []int  `yaml:"weekdays"`
	Hours     []int  `yaml:"hours"`
}
