package tools

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.design/x/clipboard"
	"gopkg.in/yaml.v3"
)

var (
	config *Config
	cfn    = "config.yaml" // config file name
)

type Config struct {
	Kuboard  Kuboard  `yaml:"kuboard"`
	Selected Selected `yaml:"selected"`
}

type Kuboard struct {
	Host     string   `yaml:"host"`
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
	Token    string   `yaml:"token"`
	Clusters []string `yaml:"clusters"`
}

type Selected struct {
	Cluster   string `yaml:"cluster"`
	Namespace string `yaml:"namespace"`
}

func newConfig() *Config {
	return &Config{
		Kuboard: Kuboard{
			Host:     "kuboard.example.com",
			Username: "admin",
			Password: "123456",
			Token:    "",
		},
	}
}

func InitConfig() {
	path := HomeDir() + "/" + cfn
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		writeDefaultConfig()
	}
	launchConfig()

	err := clipboard.Init()
	if err != nil {
		panic(err)
	}
}

func GetConfig() *Config {
	if config != nil {
		return config
	}
	return launchConfig()
}
func launchConfig() *Config {
	dir := HomeDir()
	viper.SetConfigName("config")
	viper.AddConfigPath(dir)
	config := newConfig()
	viper.ReadInConfig()
	if err := viper.Unmarshal(config); err != nil {
		fmt.Printf("Error reading config file, %s", err)
	}
	return config
}

func HomeDir() string {
	home, _ := os.UserHomeDir()
	var dir string = home + "/.kus"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
	return dir
}

func ConfigPath() string {
	return HomeDir() + "/" + cfn
}

func (conf *Config) UpdateSelectedCtx(cluster, ns string) {
	conf.Selected.Cluster = cluster
	conf.Selected.Namespace = ns
	write(conf)
}

func (conf *Config) UpdateToken(token string) {
	conf.Kuboard.Token = token
	yaml.Marshal(conf)
	write(conf)
}

func writeDefaultConfig() {
	conf := newConfig()
	write(conf)
}
func write(conf *Config) {
	out, err := yaml.Marshal(conf)
	if err != nil {
		zap.S().Errorln(err)
	}
	err = os.WriteFile(HomeDir()+"/"+cfn, out, 0644)
	if err != nil {
		zap.S().Errorln(err)
	}
}
