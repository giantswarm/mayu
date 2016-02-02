package main

import (
	"fmt"

	"github.com/giantswarm/mayu/hostmgr"
	"github.com/ryanuber/columnize"
	"github.com/spf13/cobra"
)

var (
	listCmd = &cobra.Command{
		Use:   "list",
		Short: "List machines.",
		Long:  "List machines.",
		Run:   listRun,
	}
)

const (
	listHeader      = "IP | Serial | Profile | IPMI Address | ProviderId | Fleet | CoreOS | State | Last Boot"
	listScheme      = "%s | %s | %s | %s | %s | %s | %s | %s | %s"
	timestampFormat = "2006-01-02 15:04:05"
)

func listRun(cmd *cobra.Command, args []string) {
	hosts, err := mayu.List()
	assert(err)

	lines := []string{listHeader}
	for _, host := range hosts {
		lines = append(lines, fmt.Sprintf(listScheme,
			host.InternalAddr,
			host.Serial,
			host.Profile,
			host.IPMIAddr,
			host.ProviderId,
			host.FleetMetadata,
			host.CoreOSVersion,
			hostmgr.HostStateMap()[host.State],
			host.LastBoot.Format(timestampFormat),
		))
	}
	fmt.Println(columnize.SimpleFormat(lines))
}
