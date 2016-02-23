package config

import "flag"

// Config configuration for DnsWatcher
type Config struct {
	Hostname     string
	Email        string
	User         string
	Password     string
	SMTPHostname string
}

// NewConfig return new config instance
func NewConfig() Interface {
	return &Config{}
}

// Init init the configuration
func (c *Config) Init() {
	flag.StringVar(&c.Hostname, "host", "", "Hostname to look at")
	flag.StringVar(&c.Email, "email", "", "Email")
	flag.StringVar(&c.User, "user", "", "Smtp User ")
	flag.StringVar(&c.Password, "password", "", "Smtp password")
	flag.StringVar(&c.SMTPHostname, "smtp", "", "Smtp hostname")

	flag.Parse()
}
