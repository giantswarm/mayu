package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/giantswarm/mayu/fs"
	"github.com/giantswarm/mayu/hostmgr"
	"github.com/giantswarm/mayu/pxemgr"
	"github.com/giantswarm/micrologger"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	globalFlags = MayuFlags{}

	mainCmd = &cobra.Command{
		Use:   "mayu",
		Short: "Manage your bare metal machines",
		Long:  "",
		Run:   mainRun,
	}

	projectVersion string = "1.1.0"
	projectBuild   string = "git"
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
	pf.BoolVar(&globalFlags.help, "help", false, "Show mayu usage")
	pf.StringVar(&globalFlags.configFile, "config", DefaultConfigFile, "Path to the configuration file")
	pf.StringVar(&globalFlags.clusterDir, "cluster-directory", DefaultClusterDirectory, "Path to the cluster directory")
	pf.BoolVar(&globalFlags.showTemplates, "show-templates", DefaultShowTemplates, "Show the templates and quit")
	pf.BoolVar(&globalFlags.noGit, "no-git", DefaultNoGit, "Disable git operations")
	pf.BoolVar(&globalFlags.noTLS, "no-tls", DefaultNoTLS, "Disable tls")
	pf.StringVar(&globalFlags.tFTPRoot, "tftproot", DefaultTFTPRoot, "Path to the tftproot")
	pf.StringVar(&globalFlags.fileServerPath, "file-server-path", DefaultFileServerPath, "Path to fileserver dir.")
	pf.StringVar(&globalFlags.ignitionConfig, "ignition-config", DefaultIgnitionConfig, "Final ignition config file that is used to boot the machine")
	pf.StringVar(&globalFlags.dnsmasqTemplate, "dnsmasq-template", DefaultDnsmasqTemplate, "Dnsmasq config template")
	pf.StringVar(&globalFlags.templateSnippets, "template-snippets", DefaultTemplateSnippets, "Cloudconfig or Ignition template snippets (eg storage or network configuration)")
	pf.StringVar(&globalFlags.dnsmasq, "dnsmasq", DefaultDNSMasq, "Path to dnsmasq binary")
	pf.StringVar(&globalFlags.imagesCacheDir, "images-cache-dir", DefaultImagesCacheDir, "Directory for Container Linux images")
	pf.StringVar(&globalFlags.filesDir, "files-dir", DefaultFilesDir, "Directory for file templates")
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
	pf.BoolVar(&globalFlags.coreosAutologin, "coreos-autologin", DefaultCoreosAutologin, "Sets kernel boot param 'coreos.autologin=1'. This is handy for debugging. Do NOT use for production!")
	globalFlags.filesystem = fs.DefaultFilesystem
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

func mainRun(cmd *cobra.Command, args []string) {
	if os.Args[1] == "version" {
		println("Mayu build: ", projectBuild)
		println("Mayu version: ", projectVersion)
		return
	}
	if globalFlags.help {
		cmd.PersistentFlags().Usage()
		return
	}

	var err error
	var logger micrologger.Logger
	{
		logger, err = micrologger.New(micrologger.Config{})
		if err != nil {
			println("ERROR: failed to init logger")
			os.Exit(1)
		}
	}

	logger.Log("level", "info", "msg", fmt.Sprintf("Starting mayu version %s", projectVersion))

	// hack to make some dnsmasq versions happy
	globalFlags.tFTPRoot, err = filepath.Abs(globalFlags.tFTPRoot)

	if ok, err := globalFlags.Validate(); !ok {
		glog.Fatalln(err)
	}

	hostmgr.DisableGit = globalFlags.noGit

	var cluster *hostmgr.Cluster

	if fileExists(fmt.Sprintf("%s/cluster.json", globalFlags.clusterDir)) {
		cluster, err = hostmgr.OpenCluster(globalFlags.clusterDir, logger)
	} else {
		cluster, err = hostmgr.NewCluster(globalFlags.clusterDir, true, logger)
	}

	if err != nil {
		logger.Log("level", "error", "msg", fmt.Sprintf("unable to get a cluster: %s", err))
		os.Exit(1)
	}

	globalFlags.templateSnippets = DefaultTemplateSnippets

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
		FileServerPath:           globalFlags.fileServerPath,
		StaticHTMLPath:           globalFlags.staticHTMLPath,
		TemplateSnippets:         globalFlags.templateSnippets,
		IgnitionConfig:           globalFlags.ignitionConfig,
		ImagesCacheDir:           globalFlags.imagesCacheDir,
		FilesDir:                 globalFlags.filesDir,
		CoreosAutologin:          globalFlags.coreosAutologin,
		Version:                  projectVersion,

		Logger: logger,
	}, cluster)
	if err != nil {
		logger.Log("level", "error", "msg", fmt.Sprintf("unable to create a pxe manager:: %s", err))
		os.Exit(1)
	}

	if globalFlags.showTemplates {
		placeholderHost := hostmgr.Host{}

		b := bytes.NewBuffer(nil)
		if err := pxeManager.WriteIgnitionConfig(placeholderHost, b); err != nil {
			logger.Log("level", "error", "msg", fmt.Sprintf("error found while checking generated ignition config %s", err))
			os.Exit(1)
		}
		os.Stdout.WriteString("ignition config:\n")
		os.Stdout.WriteString(b.String())

		os.Exit(0)
	}

	err = pxeManager.Start()
	if err != nil {
		logger.Log("level", "error", "msg", err)
		os.Exit(1)
	}
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}
