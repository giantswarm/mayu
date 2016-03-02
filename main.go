package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/giantswarm/mayu/fs"
	"github.com/giantswarm/mayu/hostmgr"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
)

const (
	DefaultConfigFile           string = "/etc/mayu/config.yaml"
	DefaultClusterDirectory     string = "cluster"
	DefaultShowTemplates        bool   = false
	DefaultNoGit                bool   = false
	DefaultNoSecure             bool   = false
	DefaultTFTPRoot             string = "./tftproot"
	DefaultYochuPath            string = "./yochu"
	DefaultStaticHTMLPath       string = "./static_html"
	DefaultFirstStageScript     string = "./templates/first_stage_script.sh"
	DefaultLastStageCloudconfig string = "./templates/last_stage_cloudconfig.yaml"
	DefaultDnsmasqTemplate      string = "./templates/dnsmasq_template.conf"
	DefaultTemplateSnippets     string = "./template_snippets"
	DefaultDNSMasq              string = "/usr/sbin/dnsmasq"
	DefaultImagesCacheDir       string = "./images"
	DefaultHTTPPort             string = "4080"
	DefaultHTTPBindAddress      string = "0.0.0.0"
	DefaultTLSCertFile          string = ""
	DefaultTLSKeyFile           string = ""
)

type MayuFlags struct {
	debug   bool
	verbose bool
	version bool

	configFile           string
	clusterDir           string
	showTemplates        bool
	noGit                bool
	noSecure             bool
	tFTPRoot             string
	yochuPath            string
	staticHTMLPath       string
	firstStageScript     string
	lastStageCloudconfig string
	dnsmasqTemplate      string
	templateSnippets     string
	dnsMasq              string
	imagesCacheDir       string
	httpPort             string
	httpBindAddress      string
	tlsCertFile          string
	tlsKeyFile           string

	filesystem fs.FileSystem // internal filesystem abstraction to enable testing of file operations.
}

var (
	globalFlags = MayuFlags{}

	mainCmd = &cobra.Command{
		Use:   "mayu",
		Short: "Manage your bare metal machines",
		Long:  "",
		Run:   mainRun,
	}

	projectVersion string
	projectBuild   string
)

func init() {
	mainCmd.PersistentFlags().BoolVarP(&globalFlags.debug, "debug", "d", false, "Print debug output")
	mainCmd.PersistentFlags().BoolVarP(&globalFlags.verbose, "verbose", "v", false, "Print verbose output")
	mainCmd.PersistentFlags().BoolVar(&globalFlags.version, "version", false, "Show the version of Mayu")

	mainCmd.PersistentFlags().StringVar(&globalFlags.configFile, "config", DefaultConfigFile, "Path to the configuration file")
	mainCmd.PersistentFlags().StringVar(&globalFlags.clusterDir, "cluster-directory", DefaultClusterDirectory, "Path to the cluster directory")
	mainCmd.PersistentFlags().BoolVar(&globalFlags.showTemplates, "show-templates", DefaultShowTemplates, "Show the templates and quit")
	mainCmd.PersistentFlags().BoolVar(&globalFlags.noGit, "no-git", DefaultNoGit, "Disable git operations")
	mainCmd.PersistentFlags().BoolVar(&globalFlags.noSecure, "no-secure", DefaultNoSecure, "Disable tls")
	mainCmd.PersistentFlags().StringVar(&globalFlags.tFTPRoot, "tftproot", DefaultTFTPRoot, "Path to the tftproot")
	mainCmd.PersistentFlags().StringVar(&globalFlags.yochuPath, "yochu-path", DefaultYochuPath, "Path to Yochus assets (eg docker, etcd, rkt binaries)")
	mainCmd.PersistentFlags().StringVar(&globalFlags.staticHTMLPath, "static-html-path", DefaultStaticHTMLPath, "Path to Mayus binaries (eg. mayuctl, infopusher)")
	mainCmd.PersistentFlags().StringVar(&globalFlags.firstStageScript, "first-stage-script", DefaultFirstStageScript, "Install script to install CoreOS on disk in the first stage.")
	mainCmd.PersistentFlags().StringVar(&globalFlags.lastStageCloudconfig, "last-stage-cloudconfig", DefaultLastStageCloudconfig, "Final cloudconfig that is used to boot the machine")
	mainCmd.PersistentFlags().StringVar(&globalFlags.dnsmasqTemplate, "dnsmasq-template", DefaultDnsmasqTemplate, "dnsmasq config template")
	mainCmd.PersistentFlags().StringVar(&globalFlags.templateSnippets, "template-snippets", DefaultTemplateSnippets, "Cloudconfig template snippets (eg storage or network configuration)")
	mainCmd.PersistentFlags().StringVar(&globalFlags.dnsMasq, "dnsmasq", DefaultDNSMasq, "Path to dnsmasq binary")
	mainCmd.PersistentFlags().StringVar(&globalFlags.imagesCacheDir, "images-cache-dir", DefaultImagesCacheDir, "Directory for CoreOS images")
	mainCmd.PersistentFlags().StringVar(&globalFlags.httpPort, "http-port", DefaultHTTPPort, "HTTP port Mayu listens on")
	mainCmd.PersistentFlags().StringVar(&globalFlags.httpBindAddress, "http-bind-address", DefaultHTTPBindAddress, "HTTP address Mayu listens on")
	mainCmd.PersistentFlags().StringVar(&globalFlags.tlsCertFile, "tls-cert-file", DefaultTLSCertFile, "Path to tls certificate file")
	mainCmd.PersistentFlags().StringVar(&globalFlags.tlsKeyFile, "tls-key-file", DefaultTLSKeyFile, "Path to tls key file")

	globalFlags.filesystem = fs.DefaultFilesystem
}

var (
	conf      configuration
	tempFiles = make(chan string, 4096)

	ErrNotAllCertFilesProvided = errors.New("please configure a key and cert files for TLS secured connections.")
	ErrHTTPSCertFileNotRedable = errors.New("cannot open configured certificate file for TLS secured connections.")
	ErrHTTPSKeyFileNotReadable = errors.New("cannot open configured key file for TLS secured connections.")
)

// Validate checks the configuration based on all Validate* functions
// attached to the configuration struct.
func (g MayuFlags) Validate() (bool, error) {
	if ok, err := g.ValidateHTTPCertificateUsage(); !ok {
		return ok, err
	}

	if ok, err := g.ValidateHTTPCertificateFileExistance(); !ok {
		return ok, err
	}

	return true, nil
}

// ValidateHTTPCertificateUsage checks if the fields HTTPSCertFile and HTTPSKeyFile
// of the configuration struct are set whenever the NoSecure is set to false.
// This makes sure that users are configuring the needed certificate files when
// using TLS encrypted connections.
func (g MayuFlags) ValidateHTTPCertificateUsage() (bool, error) {
	if g.noSecure == true {
		return true, nil
	}

	if g.noSecure == false && g.tlsCertFile != "" && g.tlsKeyFile != "" {
		return true, nil
	}

	return false, ErrNotAllCertFilesProvided
}

// ValidateHTTPCertificateFileExistance checks if the filenames configured
// in the fields HTTPSCertFile and HTTPSKeyFile can be stat'ed to make sure
// they actually exist.
func (g MayuFlags) ValidateHTTPCertificateFileExistance() (bool, error) {
	if g.noSecure == true {
		return true, nil
	}

	if _, err := g.filesystem.Stat(g.tlsCertFile); err != nil {
		return false, ErrHTTPSCertFileNotRedable
	}

	if _, err := g.filesystem.Stat(g.tlsKeyFile); err != nil {
		return false, ErrHTTPSKeyFileNotReadable
	}

	return true, nil
}

func mainRun(cmd *cobra.Command, args []string) {
	glog.V(8).Infoln(fmt.Sprintf("starting mayu version %s", projectVersion))

	var err error

	// hack to make some dnsmasq versions happy
	globalFlags.tFTPRoot, err = filepath.Abs(globalFlags.tFTPRoot)

	if ok, err := globalFlags.Validate(); !ok {
		glog.Fatalln(err)
	}

	hostmgr.DisableGit = globalFlags.noGit

	conf, err = loadConfig(globalFlags.configFile)
	if err != nil {
		glog.Fatalln(err)
	}

	var cluster *hostmgr.Cluster

	if fileExists(fmt.Sprintf("%s/cluster.json", globalFlags.clusterDir)) {
		cluster, err = hostmgr.OpenCluster(globalFlags.clusterDir)
	} else {
		cluster, err = hostmgr.NewCluster(globalFlags.clusterDir, true)
	}

	if err != nil {
		glog.Fatalf("unable to get a cluster: %s\n", err)
	}

	pxeManager, err := defaultPXEManager(cluster)
	if err != nil {
		glog.Fatalf("unable to create a pxe manager: %s\n", err)
	}

	if globalFlags.showTemplates {
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

func main() {
	log.SetFlags(0)
	log.SetPrefix("mayu: ")

	if globalFlags.version {
		printVersion()
		os.Exit(0)
	}

	mainCmd.Execute()
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}
