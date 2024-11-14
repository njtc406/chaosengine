package config

import (
	"github.com/njtc406/chaosengine/engine/def"
)

type config struct {
	Services map[string]*def.ServiceInitConf
}
