// Code managed by Bootstrap

package vcluster_fs_syncer //nolint:revive

import (
	"context"
	"os"

	"github.com/getoutreach/gobox/pkg/cfg"
	"github.com/getoutreach/gobox/pkg/events"
	"github.com/getoutreach/gobox/pkg/log"
)

// Config tracks config needed for vcluster-fs-syncer
type Config struct {
	ListenHost string `yaml:"ListenHost"`
	HTTPPort   int    `yaml:"HTTPPort"`
	///Block(config)
	FromPath string `yaml:"fromPath"`
	ToPath   string `yaml:"toPath"`
	///EndBlock(config)
}

// MarshalLog can be used to write config to log
func (c *Config) MarshalLog(addfield func(key string, value interface{})) {
	///Block(marshalconfig)
	///EndBlock(marshalconfig)
}

func LoadConfig(ctx context.Context) *Config { //nolint: funlen
	// NOTE: Defaults should generally be set in the config
	// override jsonnet file: deployments/vcluster-fs-syncer/vcluster-fs-syncer.config.jsonnet
	c := Config{
		// Defaults to [::]/0.0.0.0 which will broadcast to all reachable
		// IPs on a server on the given port for the respective service.
		ListenHost: "",
		HTTPPort:   8000,
		///Block(defconfig)
		///EndBlock(defconfig)
	}

	// Attempt to load a local config file on top of the defaults
	if err := cfg.Load("vcluster-fs-syncer.yaml", &c); os.IsNotExist(err) {
		log.Info(ctx, "No configuration file detected. Using default settings")
	} else if err != nil {
		log.Error(ctx, "Failed to load configuration file, will use default settings", events.NewErrorInfo(err))
	}

	// Do any necessary tweaks/augmentations to your configuration here
	///Block(configtweak)
	///EndBlock(configtweak)

	log.Info(ctx, "Configuration data of the application:\n", &c)

	return &c
}
