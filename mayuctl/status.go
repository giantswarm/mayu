package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/giantswarm/mayu/hostmgr"
	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
)

var (
	statusCmd = &cobra.Command{
		Use:   "status [serial]",
		Short: "Status of a host.",
		Long:  "Status of a host.",
		Run:   statusRun,
	}
)

const statusScheme = "%s | %s"

func statusRun(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		fmt.Println("serial missing")
		os.Exit(1)
	}

	host, err := mayu.Status(args[0])
	assert(err)

	macs := strings.Join(host.MacAddresses, ", ")
	metadata := strings.Join(host.FleetMetadata, ", ")

	lines := []string{}
	lines = append(lines, fmt.Sprintf(statusScheme, "Serial:", host.Serial))
	lines = append(lines, fmt.Sprintf(statusScheme, "IP:", host.InternalAddr))
	lines = append(lines, fmt.Sprintf(statusScheme, "IPMI:", host.IPMIAddr))
	lines = append(lines, fmt.Sprintf(statusScheme, "Provider ID:", host.ProviderId))
	lines = append(lines, fmt.Sprintf(statusScheme, "Macs:", macs))
	lines = append(lines, fmt.Sprintf("%s | %d", "Cabinet:", host.Cabinet))
	lines = append(lines, fmt.Sprintf("%s | %d", "Machine on Cabinet:", host.MachineOnCabinet))
	lines = append(lines, fmt.Sprintf(statusScheme, "Hostname:", host.Hostname))
	lines = append(lines, fmt.Sprintf(statusScheme, "MachineID:", host.MachineID))
	lines = append(lines, fmt.Sprintf(statusScheme, "ConnectedNIC:", host.ConnectedNIC))
	lines = append(lines, fmt.Sprintf(statusScheme, "Profile:", host.Profile))
	lines = append(lines, fmt.Sprintf(statusScheme, "DisableEngine:", strconv.FormatBool(host.FleetDisableEngine)))
	lines = append(lines, fmt.Sprintf(statusScheme, "State:", hostmgr.HostStateMap()[host.State]))
	lines = append(lines, fmt.Sprintf(statusScheme, "EtcdToken:", host.EtcdClusterToken))
	lines = append(lines, fmt.Sprintf(statusScheme, "Metadata:", metadata))
	lines = append(lines, fmt.Sprintf(statusScheme, "CoreOS:", host.CoreOSVersion))
	lines = append(lines, fmt.Sprintf(statusScheme, "Mayu:", host.MayuVersion))
	lines = append(lines, fmt.Sprintf(statusScheme, "Yochu:", host.YochuVersion))
	lines = append(lines, fmt.Sprintf(statusScheme, "Docker:", host.DockerVersion))
	lines = append(lines, fmt.Sprintf(statusScheme, "Etcd:", host.EtcdVersion))
	lines = append(lines, fmt.Sprintf(statusScheme, "Fleet:", host.FleetVersion))
	lines = append(lines, fmt.Sprintf(statusScheme, "Rkt:", host.RktVersion))
	lines = append(lines, fmt.Sprintf(statusScheme, "K8s:", host.K8sVersion))
	lines = append(lines, fmt.Sprintf(statusScheme, "LastBoot:", host.LastBoot))
	lines = append(lines, fmt.Sprintf(statusScheme, "Enabled:", strconv.FormatBool(host.Enabled)))
	fmt.Println(columnize.SimpleFormat(lines))
}
