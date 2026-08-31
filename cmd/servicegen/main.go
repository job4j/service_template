package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/job4j/service_template/internal/generator"
)

const usage = `servicegen creates a Go microservice from the Job4j service template.

Usage:
  servicegen new --name <service-name> --module <go-module> [--output <directory>] [--port <port>]

Example:
  servicegen new --name sharetrip-contract --module github.com/student/sharetrip-contract
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "servicegen:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing command\n\n%s", usage)
	}

	switch args[0] {
	case "new":
		return runNew(args[1:])
	case "help", "-h", "--help":
		fmt.Print(usage)
		return nil
	default:
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usage)
	}
}

func runNew(args []string) error {
	flags := flag.NewFlagSet("new", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	name := flags.String("name", "", "service name, for example sharetrip-contract")
	modulePath := flags.String("module", "", "Go module path, for example github.com/student/sharetrip-contract")
	output := flags.String("output", "", "output directory; defaults to the service name")
	port := flags.Int("port", 8080, "default HTTP port for the generated service")

	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", flags.Args())
	}

	config := generator.Config{
		Name:       *name,
		ModulePath: *modulePath,
		OutputDir:  *output,
		HTTPPort:   *port,
	}
	if err := generator.Generate(config); err != nil {
		return err
	}

	destination := config.OutputDir
	if destination == "" {
		destination = config.Name
	}

	fmt.Printf("Service %s created in %s\n", config.Name, destination)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", destination)
	fmt.Println("  make tools")
	fmt.Println("  make generate")
	fmt.Println("  make run")
	fmt.Println()
	fmt.Println("Default HTTP port: " + strconv.Itoa(config.HTTPPort))
	return nil
}
