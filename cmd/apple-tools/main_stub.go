//go:build !darwin

package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "apple-tools is only available on macOS")
	os.Exit(1)
}
