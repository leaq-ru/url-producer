package config

import (
	"github.com/kelseyhightower/envconfig"
)

type c struct {
	STAN        stan
	NATS        nats
	MongoDB     mongodb
	DomainsFile domainsFile
	LogLevel    string `envconfig:"LOGLEVEL"`
}

type stan struct {
	ClusterID string `envconfig:"STAN_CLUSTERID"`
}

type nats struct {
	URL string `envconfig:"NATS_URL"`
}

type mongodb struct {
	URL string `envconfig:"MONGODB_URL"`
}

type domainsFile struct {
	URL   string `envconfig:"DOMAINSFILE_URL"`
	URLsu string `envconfig:"DOMAINSFILE_URLSU"`
	URLrf string `envconfig:"DOMAINSFILE_URLRF"`
}

var Env c

func init() {
	envconfig.MustProcess("", &Env)
}
