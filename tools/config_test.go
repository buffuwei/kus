package tools

import (
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
	c := launchConfig()
	fmt.Printf("%+v", c)
}

func TestGetConfig(t *testing.T) {
	c := GetConfig()
	fmt.Printf("%+v \n", c)
	c = GetConfig()
	fmt.Printf("%+v \n", c)
	c = GetConfig()
	fmt.Printf("%+v \n", c)
}
