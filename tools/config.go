package tools

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/samber/lo"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.design/x/clipboard"
	"gopkg.in/yaml.v3"
)

var (
	CONFIG *Config
	CFN    = "config.yaml" // config file name
)

type Config struct {
	Kuboard  Kuboard  `yaml:"kuboard"`
	Assets   []Asset  `yaml:"assets"`
	Selected Selected `yaml:"selected"`
}

type Kuboard struct {
	Host     string `yaml:"host"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Token    string `yaml:"token"`
}

type Asset struct {
	Cluster       string         `yaml:"cluster"`
	Namespace     string         `yaml:"namespace"`
	Wingsplatform *Wingsplatform `yaml:"wingsplatform,omitempty"`
}

type Wingsplatform struct {
	Host       string `yaml:"host"`
	Login      string `yaml:"login"`
	Project    string `yaml:"project"`
	Env        string `yaml:"env"`
	Regin      string `yaml:"regin"`
	Branch     string `yaml:"branch"`
	DeployCell string `yaml:"deployCell"`
}

type Selected struct {
	Cluster   string `yaml:"cluster"`
	Namespace string `yaml:"namespace"`
}

func templateConfig() *Config {
	return &Config{
		Kuboard: Kuboard{
			Host:     "kuboard.example.com",
			Username: "admin",
			Password: "123456",
			Token:    "",
		},
		Assets: []Asset{
			{
				Cluster:   "AWS",
				Namespace: "default",
				Wingsplatform: &Wingsplatform{
					Host:       "https://wings.example.com",
					Login:      "https://wings.example2.com/api/login",
					Project:    "comm",
					Env:        "test",
					Regin:      "us-west-2",
					Branch:     "qa",
					DeployCell: "java-template-hy",
				},
			},
		},
	}
}

func InitConfig() {
	path := HomeDir() + "/" + CFN
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		templateConfig().write()
	}

	CONFIG = readConfigFile()

	err := clipboard.Init()
	if err != nil {
		panic(err)
	}
}

var muLaunchConfig sync.Mutex = sync.Mutex{}

func GetConfig() *Config {
	if CONFIG == nil {
		muLaunchConfig.Lock()
		if CONFIG == nil {
			CONFIG = readConfigFile()
		}
		defer muLaunchConfig.Unlock()
	}
	return CONFIG
}

func readConfigFile() *Config {
	dir := HomeDir()
	viper.SetConfigName("config")
	viper.AddConfigPath(dir)
	config := &Config{}
	viper.ReadInConfig()
	if err := viper.Unmarshal(config); err != nil {
		fmt.Printf("Error reading config file, %s", err)
	}

	return config
}

func (conf *Config) write() {
	out, err := yaml.Marshal(conf)
	if err != nil {
		zap.S().Errorln(err)
	}
	err = os.WriteFile(HomeDir()+"/"+CFN, out, 0644)
	if err != nil {
		zap.S().Errorln(err)
	}
}

func HomeDir() string {
	home, _ := os.UserHomeDir()
	var dir string = home + "/.kus"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
	return dir
}

func (conf *Config) UpdateSelectedCtx(cluster, ns string) {
	conf.Selected.Cluster = cluster
	conf.Selected.Namespace = ns
	conf.write()
}

func (conf *Config) UpdateToken(token string) {
	conf.Kuboard.Token = token
	yaml.Marshal(conf)
	conf.write()
}

func (conf *Config) GetSelectedAsset() *Asset {
	asset, b := lo.Find(conf.Assets, func(asset Asset) bool {
		return asset.Cluster == conf.Selected.Cluster &&
			asset.Namespace == conf.Selected.Namespace
	})
	if !b {
		return nil
	} else {
		return &asset
	}
}
