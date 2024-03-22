package main

import (
	"buffuwei/kus/tools"
	"buffuwei/kus/view"
)

func init() {
	tools.InitLogger()
	tools.InitConfig()
}

func main() {
	view.StartApplication()
}
