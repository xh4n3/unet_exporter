package sdk

type Config struct {
	Global  *Global   `yaml:"global"`
	Targets []*Target `yaml:"targets"`
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

type Target struct {
	Name             string         `yaml:"name"`
	Region           string         `yaml:"region"`
	DefaultBandwidth int            `yaml:"default_bandwidth"`
	VariedLimits     []*VariedLimit `yaml:"varied_limits"`
	RaiseRatio       int            `yaml:"raise_ratio"`
	QueryEndpoint    string         `yaml:"query_endpoint"`
}

type VariedLimit struct {
	Name      string `yaml:"name"`
	UpLimit   int    `yaml:"up_limit"`
	DownLimit int    `yaml:"down_limit"`
	WeekDays  []int  `yaml:"weekdays"`
	Hours     []int  `yaml:"hours"`
}
