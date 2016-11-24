package collector

type Config struct {
	Global  *Global   `yaml:"global"`
	Targets []*Target `yaml:"targets"`
}

type Global struct {
	ApiEndpoint string `yaml:"api_endpoint"`
	PublicKey   string `yaml:"public_key"`
	PrivateKey  string `yaml:"private_key"`
}

type Target struct {
	Name      string `yaml:"name"`
	UpLimit   int    `yaml:"up_limit"`
	DownLimit int    `yaml:"down_limit"`
	Region    string `yaml:"region"`
}
