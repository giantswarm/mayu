package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/giantswarm/mayu/fs"
	"github.com/giantswarm/mayu/hostmgr"
	"github.com/giantswarm/mayu/pxemgr"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

const (
	DefaultConfigFile               string = "/etc/mayu/config.yaml"
	DefaultClusterDirectory         string = "cluster"
	DefaultShowTemplates            bool   = false
	DefaultNoGit                    bool   = false
	DefaultNoTLS                    bool   = false
	DefaultTFTPRoot                 string = "./tftproot"
	DefaultYochuPath                string = "./yochu"
	DefaultStaticHTMLPath           string = "./static_html"
	DefaultFirstStageScript         string = "./templates/first_stage_script.sh"
	DefaultLastStageCloudconfig     string = "./templates/last_stage_cloudconfig.yaml"
	DefaultIgnitionConfig           string = "./templates/ignition/gs_install.yaml"
	DefaultDnsmasqTemplate          string = "./templates/dnsmasq_template.conf"
	DefaultTemplateSnippets         string = "./template_snippets/cloudconfig"
	DefaultIgnitionTemplateSnippets string = "./template_snippets/ignition"
	DefaultDNSMasq                  string = "/usr/sbin/dnsmasq"
	DefaultImagesCacheDir           string = "./images"
	DefaultAPIPort                  int    = 4080
	DefaultPXEPort                  int    = 4081
	DefaultHTTPBindAddress          string = "0.0.0.0"
	DefaultTLSCertFile              string = ""
	DefaultTLSKeyFile               string = ""
	DefaultUseInternalEtcdDiscovery bool   = true
	DefaultEtcdQuorumSize           int    = 3
	DefaultEtcdDiscoveryUrl         string = ""
	DefaultEtcdEndpoint             string = "http://127.0.0.1:2379"
	DefaultEtcdCA                   string = ""
)

type MayuFlags struct {
	debug   bool
	version bool

	configFile               string
	clusterDir               string
	showTemplates            bool
	noGit                    bool
	noTLS                    bool
	tFTPRoot                 string
	yochuPath                string
	staticHTMLPath           string
	firstStageScript         string
	lastStageCloudconfig     string
	ignitionConfig           string
	useIgnition              bool
	templateSnippets         string
	dnsmasq                  string
	dnsmasqTemplate          string
	imagesCacheDir           string
	apiPort                  int
	pxePort                  int
	bindAddress              string
	tlsCertFile              string
	tlsKeyFile               string
	useInternalEtcdDiscovery bool
	etcdQuorumSize           int
	etcdDiscoveryUrl         string
	etcdEndpoint             string
	etcdCAfile               string

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
	// make sure Mayu logs to stderr
	if err := flag.Lookup("logtostderr").Value.Set("true"); err != nil {
		panic(err)
	}

	// Map any flags registered in the standard "flag" package into the
	// top-level mayu command (eg. log flags)
	pf := mainCmd.PersistentFlags()
	flag.VisitAll(func(f *flag.Flag) {
		pf.AddFlag(pflag.PFlagFromGoFlag(f))
	})

	pf.BoolVarP(&globalFlags.debug, "debug", "d", false, "Print debug output")
	pf.BoolVar(&globalFlags.version, "version", false, "Show the version of Mayu")

	pf.StringVar(&globalFlags.configFile, "config", DefaultConfigFile, "Path to the configuration file")
	pf.StringVar(&globalFlags.clusterDir, "cluster-directory", DefaultClusterDirectory, "Path to the cluster directory")
	pf.BoolVar(&globalFlags.showTemplates, "show-templates", DefaultShowTemplates, "Show the templates and quit")
	pf.BoolVar(&globalFlags.noGit, "no-git", DefaultNoGit, "Disable git operations")
	pf.BoolVar(&globalFlags.noTLS, "no-tls", DefaultNoTLS, "Disable tls")
	pf.StringVar(&globalFlags.tFTPRoot, "tftproot", DefaultTFTPRoot, "Path to the tftproot")
	pf.StringVar(&globalFlags.yochuPath, "yochu-path", DefaultYochuPath, "Path to Yochus assets (eg docker, etcd, rkt binaries)")
	pf.StringVar(&globalFlags.staticHTMLPath, "static-html-path", DefaultStaticHTMLPath, "Path to Mayus binaries (eg. mayuctl, infopusher)")
	pf.StringVar(&globalFlags.firstStageScript, "first-stage-script", DefaultFirstStageScript, "Install script to install CoreOS on disk in the first stage.")
	pf.StringVar(&globalFlags.lastStageCloudconfig, "last-stage-cloudconfig", DefaultLastStageCloudconfig, "Final cloudconfig that is used to boot the machine")
	pf.StringVar(&globalFlags.ignitionConfig, "ignition-config", DefaultIgnitionConfig, "Final ignition config file that is used to boot the machine")
	pf.BoolVar(&globalFlags.useIgnition, "use-ignition", false, "Use ignition configuration setup")
	pf.StringVar(&globalFlags.dnsmasqTemplate, "dnsmasq-template", DefaultDnsmasqTemplate, "Dnsmasq config template")
	pf.StringVar(&globalFlags.templateSnippets, "template-snippets", DefaultTemplateSnippets, "Cloudconfig or Ignition template snippets (eg storage or network configuration)")
	pf.StringVar(&globalFlags.dnsmasq, "dnsmasq", DefaultDNSMasq, "Path to dnsmasq binary")
	pf.StringVar(&globalFlags.imagesCacheDir, "images-cache-dir", DefaultImagesCacheDir, "Directory for CoreOS images")
	pf.IntVar(&globalFlags.apiPort, "api-port", DefaultAPIPort, "API HTTP port Mayu listens on")
	pf.IntVar(&globalFlags.pxePort, "pxe-port", DefaultPXEPort, "PXE HTTP port Mayu listens on")
	pf.StringVar(&globalFlags.bindAddress, "http-bind-address", DefaultHTTPBindAddress, "HTTP address Mayu listens on")
	pf.StringVar(&globalFlags.tlsCertFile, "tls-cert-file", DefaultTLSCertFile, "Path to tls certificate file")
	pf.StringVar(&globalFlags.tlsKeyFile, "tls-key-file", DefaultTLSKeyFile, "Path to tls key file")
	pf.BoolVar(&globalFlags.useInternalEtcdDiscovery, "use-internal-etcd-discovery", DefaultUseInternalEtcdDiscovery, "Use the internal etcd discovery")
	pf.IntVar(&globalFlags.etcdQuorumSize, "etcd-quorum-size", DefaultEtcdQuorumSize, "Default quorum of the etcd clusters")
	pf.StringVar(&globalFlags.etcdDiscoveryUrl, "etcd-discovery", DefaultEtcdDiscoveryUrl, "External etcd discovery base url (eg https://discovery.etcd.io). Note: This should be the base URL of the discovery without a specific token. Mayu itself creates a token for the etcd clusters.")
	pf.StringVar(&globalFlags.etcdEndpoint, "etcd-endpoint", DefaultEtcdEndpoint, "The etcd endpoint for the internal discovery feature (you must also specify protocol).")
	pf.StringVar(&globalFlags.etcdCAfile, "etcd-cafile", DefaultEtcdCA, "The etcd CA file, if etcd is using non-trustred root CA certificate")

	globalFlags.filesystem = fs.DefaultFilesystem
}

var (
	ErrNotAllCertFilesProvided = errors.New("Please configure a key and cert files for TLS connections.")
	ErrHTTPSCertFileNotRedable = errors.New("Cannot open configured certificate file for TLS connections.")
	ErrHTTPSKeyFileNotReadable = errors.New("Cannot open configured key file for TLS connections.")
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
// of the configuration struct are set whenever the NoTLS is set to false.
// This makes sure that users are configuring the needed certificate files when
// using TLS encrypted connections.
func (g MayuFlags) ValidateHTTPCertificateUsage() (bool, error) {
	if g.noTLS {
		return true, nil
	}

	if !g.noTLS && g.tlsCertFile != "" && g.tlsKeyFile != "" {
		return true, nil
	}

	return false, ErrNotAllCertFilesProvided
}

// ValidateHTTPCertificateFileExistance checks if the filenames configured
// in the fields HTTPSCertFile and HTTPSKeyFile can be stat'ed to make sure
// they actually exist.
func (g MayuFlags) ValidateHTTPCertificateFileExistance() (bool, error) {
	if g.noTLS {
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
	glog.V(8).Infof("starting mayu version %s", projectVersion)

	var err error

	// hack to make some dnsmasq versions happy
	globalFlags.tFTPRoot, err = filepath.Abs(globalFlags.tFTPRoot)

	if ok, err := globalFlags.Validate(); !ok {
		glog.Fatalln(err)
	}

	hostmgr.DisableGit = globalFlags.noGit

	var cluster *hostmgr.Cluster

	if fileExists(fmt.Sprintf("%s/cluster.json", globalFlags.clusterDir)) {
		cluster, err = hostmgr.OpenCluster(globalFlags.clusterDir)
	} else {
		cluster, err = hostmgr.NewCluster(globalFlags.clusterDir, true)
	}

	if err != nil {
		glog.Fatalf("unable to get a cluster: %s\n", err)
	}
	if globalFlags.useIgnition && globalFlags.templateSnippets == DefaultTemplateSnippets {
		globalFlags.templateSnippets = DefaultIgnitionTemplateSnippets
	}

	pxeManager, err := pxemgr.PXEManager(pxemgr.PXEManagerConfiguration{
		ConfigFile:               globalFlags.configFile,
		UseInternalEtcdDiscovery: globalFlags.useInternalEtcdDiscovery,
		EtcdQuorumSize:           globalFlags.etcdQuorumSize,
		EtcdDiscoveryUrl:         globalFlags.etcdDiscoveryUrl,
		EtcdEndpoint:             globalFlags.etcdEndpoint,
		EtcdCAFile:               globalFlags.etcdCAfile,
		DNSmasqExecutable:        globalFlags.dnsmasq,
		DNSmasqTemplate:          globalFlags.dnsmasqTemplate,
		TFTPRoot:                 globalFlags.tFTPRoot,
		NoTLS:                    globalFlags.noTLS,
		APIPort:                  globalFlags.apiPort,
		PXEPort:                  globalFlags.pxePort,
		BindAddress:              globalFlags.bindAddress,
		TLSCertFile:              globalFlags.tlsCertFile,
		TLSKeyFile:               globalFlags.tlsKeyFile,
		YochuPath:                globalFlags.yochuPath,
		StaticHTMLPath:           globalFlags.staticHTMLPath,
		TemplateSnippets:         globalFlags.templateSnippets,
		LastStageCloudconfig:     globalFlags.lastStageCloudconfig,
		IgnitionConfig:           globalFlags.ignitionConfig,
		UseIgnition:              globalFlags.useIgnition,
		FirstStageScript:         globalFlags.firstStageScript,
		ImagesCacheDir:           globalFlags.imagesCacheDir,
		Version:                  projectVersion,
	}, cluster)
	if err != nil {
		glog.Fatalf("unable to create a pxe manager: %s\n", err)
	}

	if globalFlags.showTemplates {
		placeholderHost := hostmgr.Host{}

		if globalFlags.useIgnition {
			b := bytes.NewBuffer(nil)
			if err := pxeManager.WriteIgnitionConfig(placeholderHost, b); err != nil {
				fmt.Errorf("error found while checking generated ignition config: %+v", err)
				os.Exit(1)
			}
			os.Stdout.WriteString("ignition config:\n")
			os.Stdout.WriteString(b.String())
		} else {
			os.Stdout.WriteString("last stage cloud config:\n")
			pxeManager.WriteLastStageCC(placeholderHost, os.Stdout)

			b := bytes.NewBuffer(nil)
			pxeManager.WriteLastStageCC(placeholderHost, b)
			yamlErr := validateYAML(b.Bytes())
			if yamlErr != nil {
				fmt.Errorf("error found while checking generated cloud-config: %+v", yamlErr)
				os.Exit(1)
			}
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

func validateYAML(yml []byte) error {
	y := map[string]interface{}{}
	return yaml.Unmarshal(yml, &y)
}
