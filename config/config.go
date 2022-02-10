package config

import (
	"os"
	"path/filepath"
)

var HomePath = filepath.Join(os.Getenv("HOME"), "compositor")
var TempPath = filepath.Join(os.Getenv("HOME"), ".compositor", "tmp")

func init() {
	if err := os.MkdirAll(TempPath, 0755); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(HomePath, 0755); err != nil {
		panic(err)
	}
}

var MultipassHomePath = "/home/ubuntu"
