package config

import (
	"github.com/kelseyhightower/envconfig"
)

type c struct {
	STAN           stan
	Mongo          mongo
	LogLevel       string `envconfig:"LOGLEVEL"`
	DomainsFileURL string `envconfig:"DOMAINSFILEURL"`
}

type stan struct {
	URL       string `envconfig:"STAN_URL"`
	ClusterID string `envconfig:"STAN_CLUSTERID"`
}

type mongo struct {
	URL string `envconfig:"MONGO_URL"`
}

var Env c

func init() {
	envconfig.MustProcess("", &Env)
}
