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
}

type Target struct {
	Name      string `yaml:"name"`
	UpLimit   int    `yaml:"up_limit"`
	DownLimit int    `yaml:"down_limit"`
	Region    string `yaml:"region"`
	Step      int    `yaml:"step"`
}
