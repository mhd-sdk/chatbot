//go:build mage

package main

import (
	"github.com/magefile/mage/sh"
)

func Dev() error {
	return sh.RunV("go", "run", "cmd/main.go")
}
func Build() error {
	return sh.RunV("go", "build", "cmd/main.go")
}
