package main

import (
	"fmt"
	"os"

	"github.com/giantswarm/mayu/hostmgr"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

type BootCompleteFlags struct {
	UpdateVersions bool
}

var (
	bootCompleteCmd = &cobra.Command{
		Use:   "boot-complete",
		Short: "Change the state of a host to 'running' (only run on provisioned machines).",
		Long: `Change the state of a host to 'running' (only run on provisioned machines).

Update the software versions running on a host with '--update-versions'.
This includes versions of CoreOS, mayu, docker, etcd, fleet, rkt and the
Giant Swarm yochu.
`,
		Run: bootCompleteRun,
	}

	bootCompleteFlags = &BootCompleteFlags{}
)

func init() {
	bootCompleteCmd.PersistentFlags().BoolVar(&bootCompleteFlags.UpdateVersions, "update-versions", false, "Update installed software versions in the mayu catalog")
}

func bootCompleteRun(cmd *cobra.Command, args []string) {
	hostEnvironment, err := godotenv.Read(
		"/etc/os-release",
		"/etc/yochu-env",
		"/etc/mayu-env",
	)
	assert(err)

	serial, ok := hostEnvironment["SERIAL"]
	if !ok {
		fmt.Printf("Can't find serial in host environment (/etc/mayu-env)")
		os.Exit(1)
	}

	var host hostmgr.Host
	if bootCompleteFlags.UpdateVersions {
		for key, value := range hostEnvironment {
			switch key {
			case "VERSION":
				host.CoreOSVersion = value
			case "MAYU_VERSION":
				host.MayuVersion = value
			case "DOCKER_VERSION":
				host.DockerVersion = value
			case "ETCD_VERSION":
				host.EtcdVersion = value
			case "FLEET_VERSION":
				host.FleetVersion = value
			case "YOCHU_VERSION":
				host.YochuVersion = value
			case "RKT_VERSION":
				host.RktVersion = value
			}
		}
	}

	err = mayu.BootComplete(serial, host)
	assert(err)
}
