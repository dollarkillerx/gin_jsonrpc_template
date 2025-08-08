package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/feitian/internal/scaffold"
)

func main() {
	var (
		module   string
		appName  string
		outDir   string
		port     string
	)

	flag.StringVar(&module, "module", "github.com/you/yourapp", "Go module path of new project")
	flag.StringVar(&appName, "name", "app", "Application name")
	flag.StringVar(&outDir, "out", ".", "Output directory for the project")
	flag.StringVar(&port, "port", "8080", "HTTP port")
	flag.Parse()

	absOut, err := filepath.Abs(outDir)
	if err != nil { fmt.Fprintln(os.Stderr, err); os.Exit(1) }

	data := scaffold.Data{ Module: module, AppName: appName, HTTPPort: port }
	if err := scaffold.Generate(absOut, data); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("Project generated at %s\n", absOut)
	fmt.Println("Next steps:")
	fmt.Println("  1) cd", absOut)
	fmt.Println("  2) go mod tidy")
	fmt.Println("  3) go run ./cmd -c config -cPath \"./,./configs/\"")
}
