package cmd

type Account struct {
	AwsAccessKeyID     string   `mapstructure:"aws_access_key_id"`
	AwsSecretAccessKey string   `mapstructure:"aws_secret_access_key"`
	Regions            []string `yaml:"regions"`
}

type AWSCredentials struct {
	Accounts []Account `yaml:"accounts"`
}
