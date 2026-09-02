package config

import (
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/spf13/viper"
)

// Init sets up viper to read config from YAML file, env vars, and flags.
// It mirrors the pattern used in github.com/kthcloud/podsh.
func Init(projectName string) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	if configDir, err := configDir(projectName); err == nil {
		viper.AddConfigPath(configDir)
	}
	if wd, err := os.Getwd(); err == nil {
		viper.AddConfigPath(wd)
	}
	viper.AddConfigPath(".")

	viper.SetEnvPrefix(strings.ToUpper(strings.ReplaceAll(projectName, "-", "_")))
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		slog.Default().Debug("config file not found, using defaults or environment variables")
	} else {
		slog.Default().Debug("using config file", "path", viper.ConfigFileUsed())
	}
}

func configDir(projectName string) (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := path.Join(base, projectName)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", err
	}
	return dir, nil
}
