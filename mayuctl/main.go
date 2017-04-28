// mayuctl is the command line implementation of a mayu client. With this
// tool you can interact with mayu over network. This provides some
// convenience so you don't need to deal with the API and network calls
// yourself. For more detailed usage information check the help usage:
// mayuctl --help
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/giantswarm/mayu/client"
	"github.com/spf13/cobra"
)

const (
	defaultHost string = "localhost"
	defaultPort uint16 = 4080
)

type Flags struct {
	// Host is used to connect to mayu over network. By default host is
	// localhost. Overwrite this via command line argument --host.
	Host string

	// Port is used to connect to mayu over network. By default port is 4080.
	// Overwrite this via command line argument --port.
	Port uint16

	// NoTLS is used to disable the TLS support. By default TLS is enabled.
	// Overwrite this via command line argument --no-tls.
	NoTLS bool

	// Debug is used to enable debug logging. By default debug logging is
	// disabled. Overwrite this via command line argument --debug.
	Debug bool

	// Verbose is used to enable verbose logging. By default verbose logging is
	// disabled. Overwrite this via command line argument --verbose.
	Verbose bool
}

var (
	globalFlags Flags

	mayu *client.Client

	mainCmd = &cobra.Command{
		Use:   "mayuctl",
		Short: "Manage a mayu cluster",
		Long:  "Manage a mayu cluster",
		Run:   mainRun,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			scheme := "https"
			if globalFlags.NoTLS {
				scheme = "http"
			}

			var err error
			mayu, err = client.New(scheme, globalFlags.Host, globalFlags.Port)
			assert(err)
		},
	}

	projectVersion string
	projectBuild   string
)

func init() {
	mainCmd.PersistentFlags().StringVar(&globalFlags.Host, "host", defaultHost, "Hostname to connect to mayu service")
	mainCmd.PersistentFlags().Uint16Var(&globalFlags.Port, "port", defaultPort, "Port to connect to mayu service")
	mainCmd.PersistentFlags().BoolVar(&globalFlags.NoTLS, "no-tls", false, "Do not use tls communication")
	mainCmd.PersistentFlags().BoolVarP(&globalFlags.Debug, "debug", "d", false, "Print debug output")
	mainCmd.PersistentFlags().BoolVarP(&globalFlags.Verbose, "verbose", "v", false, "Print verbose output")
}

func assert(err error) {
	if err != nil {
		if globalFlags.Debug {
			fmt.Printf("%#v\n", err)
			os.Exit(1)
		} else {
			log.Fatal(err)
		}
	}
}

func confirm(question string) error {
	for {
		fmt.Printf("%s ", question)
		bio := bufio.NewReader(os.Stdin)
		line, _, err := bio.ReadLine()
		if err != nil {
			return err
		}

		if string(line) == "yes" {
			return nil
		}
		fmt.Println("please enter 'yes' to confirm")
	}
}

func mainRun(cmd *cobra.Command, args []string) {
	cmd.Help()
}

func main() {
	mainCmd.AddCommand(versionCmd)
	mainCmd.AddCommand(listCmd)
	mainCmd.AddCommand(statusCmd)
	mainCmd.AddCommand(setCmd)
	mainCmd.AddCommand(bootCompleteCmd)
	mainCmd.AddCommand(overrideCmd)
	mainCmd.AddCommand(configCmd)

	mainCmd.Execute()
}
