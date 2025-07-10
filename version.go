package main

import (
	"fmt"
	"runtime"
)

var (
	// 这些变量会在编译时通过 -ldflags 注入
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

func printVersion() {
	fmt.Printf("tReader %s\n", Version)
	fmt.Printf("Git Commit: %s\n", GitCommit)
	fmt.Printf("Build Time: %s\n", BuildTime)
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
