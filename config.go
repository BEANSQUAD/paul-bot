package main

import (
	"github.com/lalamove/konfig"
	"github.com/lalamove/konfig/loader/klfile"
	"github.com/lalamove/konfig/parser/kptoml"
)

var configFiles = []klfile.File{
	{
		Path:   "./config/vars.toml",
		Parser: kptoml.Parser,
	},
}

func init() {
	konfig.Init(konfig.DefaultConfig())
}

// SetupConfig registers a konfig watcher to load configuration values from a
// file (e.g. API keys, persistent guild settings)
func SetupConfig() error {
	konfig.RegisterLoaderWatcher(
		klfile.New(&klfile.Config{
			Files: configFiles,
			Watch: true,
		}),
	)

	err := konfig.LoadWatch()
	return err
}
