package main

import (
	"errors"
	"github.com/giantswarm/mayu/fs"
)

const (
	DefaultConfigFile               string = "/etc/mayu/config.yaml"
	DefaultClusterDirectory         string = "cluster"
	DefaultShowTemplates            bool   = false
	DefaultNoGit                    bool   = false
	DefaultNoTLS                    bool   = false
	DefaultTFTPRoot                 string = "./tftproot"
	DefaultFileServerPath           string = "./fileserver"
	DefaultIgnitionConfig           string = "./templates/ignition.yaml"
	DefaultDnsmasqTemplate          string = "./templates/dnsmasq_template.conf"
	DefaultTemplateSnippets         string = "./templates/snippets/"
	DefaultDNSMasq                  string = "/usr/sbin/dnsmasq"
	DefaultImagesCacheDir           string = "./images"
	DefaultFilesDir                 string = "./files"
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
	DefaultCoreosAutologin          bool   = false
)

type MayuFlags struct {
	debug   bool
	version bool
	help    bool

	configFile               string
	clusterDir               string
	showTemplates            bool
	noGit                    bool
	noTLS                    bool
	tFTPRoot                 string
	fileServerPath           string
	staticHTMLPath           string
	ignitionConfig           string
	templateSnippets         string
	dnsmasq                  string
	dnsmasqTemplate          string
	imagesCacheDir           string
	filesDir                 string
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
	coreosAutologin          bool

	filesystem fs.FileSystem // internal filesystem abstraction to enable testing of file operations.
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
