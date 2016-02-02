package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/giantswarm/mayu/hostmgr"
	"github.com/golang/glog"
)

var (
	confFile      = flag.String("config", "/etc/mayu/config.yaml", "path to the configuration file")
	clusterDir    = flag.String("cluster-directory", "cluster", "path cluster directory")
	showTemplates = flag.Bool("show-templates", false, "show the templates and quit")
	noGit         = flag.Bool("no-git", false, "disable git operations")
	showVersion   = flag.Bool("version", false, "show the version of mayu")

	conf      configuration
	tempFiles = make(chan string, 4096)
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("mayu: ")
	flag.Set("logtostderr", "true")
	flag.Parse()

	if *showVersion {
		printVersion()
		os.Exit(0)
	}

	glog.V(8).Infoln(fmt.Sprintf("starting mayu version %s", projectVersion))

	hostmgr.DisableGit = *noGit

	var err error

	conf, err = loadConfig(*confFile)
	if err != nil {
		glog.Fatalln(err)
	}

	if ok, err := conf.Validate(); !ok {
		glog.Fatalln(err)
	}

	var cluster *hostmgr.Cluster

	if fileExists(*clusterDir) {
		cluster, err = hostmgr.OpenCluster(*clusterDir)
	} else {
		cluster, err = hostmgr.NewCluster(*clusterDir, true)
	}

	if err != nil {
		glog.Fatalf("unable to get a cluster: %s\n", err)
	}

	pxeManager, err := defaultPXEManager(cluster)
	if err != nil {
		glog.Fatalf("unable to create a pxe manager: %s\n", err)
	}

	if *showTemplates {
		placeholderHost := hostmgr.Host{}

		os.Stdout.WriteString("last stage cloud config:\n")
		pxeManager.writeLastStageCC(placeholderHost, os.Stdout)

		b := bytes.NewBuffer(nil)
		pxeManager.writeLastStageCC(placeholderHost, b)
		yamlErr := validateCC(b.Bytes())
		if yamlErr != nil {
			fmt.Errorf("error found while checking generated cloud-config: %+v", yamlErr)
			os.Exit(1)
		}

		os.Exit(0)
	}

	err = pxeManager.Start()
	if err != nil {
		glog.Errorln(err)
	}
}

func cleanTempFiles() {
	close(tempFiles)
	for fname := range tempFiles {
		glog.V(5).Infoln("removing temporary file", fname)
		os.RemoveAll(fname)
	}
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}
