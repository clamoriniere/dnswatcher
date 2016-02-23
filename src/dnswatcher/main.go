package main

import (
	"os"

	"config"
	"utils"
)

func main() {
	err := run()
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}

func run() error {
	c := config.NewConfig()
	c.Init()

	p := NewProcess(c)
	return utils.ExecProcess(p)
}
