package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v3"
	"github.com/josh23french/tacacs/pkg/authorpolicy"
	"github.com/josh23french/tacacs/pkg/backends/ldap"
)

type Server struct {
    Listen string
}

type Client struct {
	Hostname string
	Secret   string
}

type AuthenPolicy struct {
	Group string
	Allow bool
}

type AcctPolicy struct {
	Command string
	Trap    string
}

type BackendsConfig struct {
	LDAP *ldap.LDAPConfig
}

type Config struct {
    Server   Server
	Clients  []Client
	Backends BackendsConfig
	Policy struct {
		Authentication []AuthenPolicy
		Authorization  []authorpolicy.AuthorPolicy
		Accounting     []AcctPolicy
	}
}

func NewConfig(filename string) *Config {
	var config Config

	fileContents, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Unable to read file: %v", err)
	}

	err = yaml.Unmarshal(fileContents, &config)
	if err != nil {
		log.Fatalf("Cannot load config: %v", err)
	}
	return &config
}
