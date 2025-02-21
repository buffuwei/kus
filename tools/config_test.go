package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestConfigDir(t *testing.T) {
	home, _ := os.UserHomeDir()
	var dir string = home + "/.kus"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
}

func TestReadConfig(t *testing.T) {
	fmt.Printf("Begin \n")
	conf := readConfigFile()
	jsonStr, _ := json.MarshalIndent(conf, "", " ")
	fmt.Printf("%s\n", jsonStr)
}

func TestGetConfig(t *testing.T) {
	conf := GetConfig()
	jsonStr, _ := json.MarshalIndent(conf, "", " ")
	fmt.Printf("%s\n", jsonStr)
}

func TestNewConfig(tt *testing.T) {
	config := &Config{}
	bs, _ := json.MarshalIndent(config, "", " ")
	fmt.Printf("%s\n", string(bs))
}

func TestConfig(t *testing.T) {
	asset := GetConfig().Assets[2]
	wsp := asset.Wingsplatform
	t.Logf("start \n")
	fmt.Printf("begin\n")
	fmt.Printf("%+v\n", wsp)
	wspp := &wsp
	fmt.Printf("%+v\n", wspp)
}
