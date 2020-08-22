package config

import (
	"github.com/kelseyhightower/envconfig"
)

type c struct {
	STAN           stan
	NATS           nats
	Mongo          mongo
	LogLevel       string `envconfig:"LOGLEVEL"`
	DomainsFileURL string `envconfig:"DOMAINSFILEURL"`
}

type stan struct {
	ClusterID string `envconfig:"STAN_CLUSTERID"`
}

type nats struct {
	URL string `envconfig:"NATS_URL"`
}

type mongo struct {
	URL string `envconfig:"MONGO_URL"`
}

var Env c

func init() {
	envconfig.MustProcess("", &Env)
}
