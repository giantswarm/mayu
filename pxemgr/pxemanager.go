package pxemgr

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/giantswarm/mayu/hostmgr"
	"github.com/giantswarm/mayu/logging"
	"github.com/golang/glog"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type PXEManagerConfiguration struct {
	ConfigFile               string
	UseInternalEtcdDiscovery bool
	EtcdQuorumSize           int
	EtcdDiscoveryUrl         string
	EtcdEndpoint             string
	DNSmasqExecutable        string
	DNSmasqTemplate          string
	TFTPRoot                 string
	NoTLS                    bool
	TLSCertFile              string
	TLSKeyFile               string
	HTTPPort                 int
	HTTPBindAddress          string
	YochuPath                string
	StaticHTMLPath           string
	TemplateSnippets         string
	LastStageCloudconfig     string
	IgnitionConfig           string
	UseIgnition              bool
	FirstStageScript         string
	ImagesCacheDir           string
	Version                  string
}

type pxeManagerT struct {
	noTLS                    bool
	httpPort                 int
	httpBindAddress          string
	tlsCertFile              string
	tlsKeyFile               string
	yochuPath                string
	staticHTMLPath           string
	templateSnippets         string
	lastStageCloudconfig     string
	firstStageScript         string
	ignitionConfig           string
	useIgnition              bool
	imagesCacheDir           string
	useInternalEtcdDiscovery bool
	defaultEtcdQuorumSize    int
	etcdDiscoveryUrl         string
	etcdEndpoint             string
	version                  string

	config  *configuration
	cluster *hostmgr.Cluster
	DNSmasq *DNSmasqInstance

	mu *sync.Mutex

	router *mux.Router
}

func PXEManager(c PXEManagerConfiguration, cluster *hostmgr.Cluster) (*pxeManagerT, error) {
	conf, err := loadConfig(c.ConfigFile)
	if err != nil {
		glog.Fatalln(err)
	}

	if conf.DefaultCoreOSVersion == "" {
		glog.Fatalf("No default_coreos_version specified in %s\n", c.ConfigFile)
	}

	if c.EtcdDiscoveryUrl != "" && c.UseInternalEtcdDiscovery {
		glog.Fatalln("External etcd discovery url is set and internal etcd discovery is activated. Please choose only one.")
	} else if c.EtcdDiscoveryUrl == "" && !c.UseInternalEtcdDiscovery {
		glog.Fatalln("The internal etcd discovery is deactivated and no external discovery url is given")
	} else if c.UseInternalEtcdDiscovery && c.EtcdEndpoint == "" {
		glog.Fatalln("The internal etcd discovery is activated but no etcd endpoint is given")
	}

	c.EtcdDiscoveryUrl = strings.TrimRight(c.EtcdDiscoveryUrl, "/")

	mgr := &pxeManagerT{
		noTLS:                    c.NoTLS,
		httpPort:                 c.HTTPPort,
		httpBindAddress:          c.HTTPBindAddress,
		tlsCertFile:              c.TLSCertFile,
		tlsKeyFile:               c.TLSKeyFile,
		yochuPath:                c.YochuPath,
		staticHTMLPath:           c.StaticHTMLPath,
		templateSnippets:         c.TemplateSnippets,
		lastStageCloudconfig:     c.LastStageCloudconfig,
		ignitionConfig:           c.IgnitionConfig,
		useIgnition:              c.UseIgnition,
		firstStageScript:         c.FirstStageScript,
		imagesCacheDir:           c.ImagesCacheDir,
		useInternalEtcdDiscovery: c.UseInternalEtcdDiscovery,
		defaultEtcdQuorumSize:    c.EtcdQuorumSize,
		etcdDiscoveryUrl:         c.EtcdDiscoveryUrl,
		etcdEndpoint:             c.EtcdEndpoint,
		version:                  c.Version,

		config:  &conf,
		cluster: cluster,
		DNSmasq: NewDNSmasq("/tmp/dnsmasq.mayu", DNSmasqConfiguration{
			Executable: c.DNSmasqExecutable,
			Template:   c.DNSmasqTemplate,
			TFTPRoot:   c.TFTPRoot,
			NoTLS:      c.NoTLS,
			HTTPPort:   c.HTTPPort,
		}),
		mu: new(sync.Mutex),
	}

	// check for deprecated EtcdDiscoveryUrl
	if mgr.cluster.Config.EtcdDiscoveryURL != "" && mgr.cluster.Config.DefaultEtcdClusterToken == "" {
		// transform discovery url to token
		parts := strings.Split(mgr.cluster.Config.EtcdDiscoveryURL, "/")
		token := parts[len(parts)-1]
		baseUrl := strings.Join(parts[:len(parts)-1], "/")

		if mgr.useInternalEtcdDiscovery {
			// convert token to internal etcd discovery
			err := mgr.cluster.StoreEtcdDiscoveryToken(mgr.etcdEndpoint, token, mgr.defaultEtcdQuorumSize)
			if err != nil {
				glog.Fatalf("Can't store discovery token in etcd.", baseUrl, mgr.etcdDiscoveryUrl)
			}

			glog.Warningf("Transferred etcd token to internal discovery. Note that your machines still have the old discovery url in their cloud-config and that you need to transfer the current member data yourself.")
		} else if mgr.etcdDiscoveryUrl != baseUrl {
			glog.Fatalf("Deprecated EtcdDiscoveryURL ('%s') in your cluster.json differs from the --etcd-discovery parameter ('%s').", baseUrl, mgr.etcdDiscoveryUrl)
		}
		mgr.cluster.Config.EtcdDiscoveryURL = ""
		mgr.cluster.Config.DefaultEtcdClusterToken = token
		mgr.cluster.Commit(fmt.Sprintf("Convert deprecated etcd discovery url to default etcd token '%s'", token))
	}

	if mgr.cluster.Config.DefaultEtcdClusterToken == "" {
		var (
			token string
			err   error
		)
		if mgr.useInternalEtcdDiscovery {
			mgr.etcdDiscoveryUrl = mgr.thisHost() + "/etcd"
			token, err = mgr.cluster.GenerateEtcdDiscoveryToken()
			if err != nil {
				glog.Fatalf("Failed to generate etcd cluster token: %s", err)
			}
			err := mgr.cluster.StoreEtcdDiscoveryToken(mgr.etcdEndpoint, token, mgr.defaultEtcdQuorumSize)
			if err != nil {
				glog.Fatalf("Failed to store etcd cluster token in etcd: %s", err)
			}
		} else {
			token, err = mgr.cluster.FetchEtcdDiscoveryToken(mgr.etcdDiscoveryUrl, mgr.defaultEtcdQuorumSize)
			if err != nil {
				glog.Fatalf("Failed to fetch etcd cluster token from external registry: %s", err)
			}
		}
		mgr.cluster.Config.DefaultEtcdClusterToken = token
		mgr.cluster.Commit(fmt.Sprintf("Set default etcd cluster to '%s'", token))
	}

	return mgr, nil
}

func withSerialParam(serialHandler func(serial string, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		serialHandler(params["serial"], w, r)
	}
}

func (mgr *pxeManagerT) startIPXEserver() error {
	mgr.router = mux.NewRouter()

	// first stage ipxe boot script
	mgr.router.Methods("GET").PathPrefix("/ipxebootscript").HandlerFunc(mgr.ipxeBootScript)
	mgr.router.Methods("GET").PathPrefix("/first-stage-script/{serial}").HandlerFunc(mgr.firstStageScriptGenerator)

	// used by the first-stage-script:
	mgr.router.Methods("GET").PathPrefix("/hostinfo-helper").HandlerFunc(mgr.infoPusher)

	if mgr.useIgnition {
		mgr.router.Methods("POST").PathPrefix("/final-ignition-config.json").HandlerFunc(mgr.configGenerator)
	} else {
		mgr.router.Methods("POST").PathPrefix("/final-cloud-config.yaml").HandlerFunc(mgr.configGenerator)
	}

	mgr.router.Methods("PUT").PathPrefix("/admin/host/{serial}/boot_complete").HandlerFunc(withSerialParam(mgr.bootComplete))
	mgr.router.Methods("PUT").PathPrefix("/admin/host/{serial}/set_installed").HandlerFunc(withSerialParam(mgr.setInstalled))
	mgr.router.Methods("PUT").PathPrefix("/admin/host/{serial}/set_metadata").HandlerFunc(withSerialParam(mgr.setMetadata))
	mgr.router.Methods("PUT").PathPrefix("/admin/host/{serial}/mark_fresh").HandlerFunc(withSerialParam(mgr.markFresh))
	mgr.router.Methods("PUT").PathPrefix("/admin/host/{serial}/set_provider_id").HandlerFunc(withSerialParam(mgr.setProviderId))
	mgr.router.Methods("PUT").PathPrefix("/admin/host/{serial}/set_ipmi_addr").HandlerFunc(withSerialParam(mgr.setIPMIAddr))
	mgr.router.Methods("PUT").PathPrefix("/admin/host/{serial}/set_cabinet").HandlerFunc(withSerialParam(mgr.setCabinet))
	mgr.router.Methods("PUT").PathPrefix("/admin/host/{serial}/set_state").HandlerFunc(withSerialParam(mgr.setState))
	mgr.router.Methods("PUT").PathPrefix("/admin/host/{serial}/set_etcd_cluster_token").HandlerFunc(withSerialParam(mgr.setEtcdClusterToken))
	mgr.router.Methods("PUT").PathPrefix("/admin/host/{serial}/override").HandlerFunc(withSerialParam(mgr.override))

	// etcd discovery
	if mgr.useInternalEtcdDiscovery {
		etcdRouter := mgr.router.PathPrefix("/etcd").Subrouter()
		mgr.defineEtcdDiscoveryRoutes(etcdRouter)
	}

	// boring stuff
	mgr.router.Methods("GET").PathPrefix("/admin/hosts").HandlerFunc(mgr.hostsList)
	mgr.router.Methods("GET").PathPrefix("/images/{serial}").HandlerFunc(mgr.imagesHandler)

	// serve assets for yochu like etcd, fleet, docker, kubectl and rkt
	mgr.router.PathPrefix("/yochu").Handler(http.StripPrefix("/yochu", http.FileServer(http.Dir(mgr.yochuPath))))

	// add welcome handler for debugging
	mgr.router.Path("/").HandlerFunc(mgr.welcomeHandler)

	// serve static files like infopusher and mayuctl etc.
	mgr.router.PathPrefix("/").Handler(http.FileServer(http.Dir(mgr.staticHTMLPath)))

	glogWrapper := logging.NewGlogWrapper(8)
	loggedRouter := handlers.LoggingHandler(glogWrapper, mgr.router)

	glog.V(8).Infoln(fmt.Sprintf("starting iPXE server at %s:%d", mgr.httpBindAddress, mgr.httpPort))

	if mgr.noTLS {
		err := http.ListenAndServe(fmt.Sprintf("%s:%d", mgr.httpBindAddress, mgr.httpPort), loggedRouter)
		if err != nil {
			return err
		}
	} else {
		err := http.ListenAndServeTLS(fmt.Sprintf("%s:%d", mgr.httpBindAddress, mgr.httpPort), mgr.tlsCertFile, mgr.tlsKeyFile, loggedRouter)
		if err != nil {
			return err
		}
	}

	return nil
}

func (mgr *pxeManagerT) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mgr.router.ServeHTTP(w, r)
}

func (mgr *pxeManagerT) updateDNSmasqs() error {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	mgr.config.Network.StaticHosts = []hostmgr.IPMac{}
	mgr.config.Network.IgnoredHosts = []string{}

	ignoredHostPredicate := func(host *hostmgr.Host) bool {
		// ignore hosts that are installed or running
		return host.State == hostmgr.Installed || host.State == hostmgr.Running
	}

	for host := range mgr.cluster.FilterHostsFunc(ignoredHostPredicate) {
		for _, macAddr := range host.MacAddresses {
			mgr.config.Network.IgnoredHosts = append(mgr.config.Network.IgnoredHosts, macAddr)
		}
	}

	err := mgr.DNSmasq.updateConf(mgr.config.Network)
	if err != nil {
		return err
	}
	err = mgr.DNSmasq.Restart()
	if err != nil {
		return err
	}

	return nil
}

func (mgr *pxeManagerT) Start() error {
	err := mgr.DNSmasq.Start()
	if err != nil {
		return err
	}

	err = mgr.updateDNSmasqs()
	if err != nil {
		return err
	}

	return mgr.startIPXEserver()
}

func (mgr *pxeManagerT) getNextProfile() string {
	profileCount := mgr.cluster.GetProfileCount()

	for _, profile := range mgr.config.Profiles {
		if profileCount[profile.Name] < profile.Quantity {
			return profile.Name
		}
	}
	return ""
}

func (mgr *pxeManagerT) getNextInternalIP() net.IP {
	assignedIPs := map[string]struct{}{}
	for _, host := range mgr.cluster.GetAllHosts() {
		assignedIPs[host.InternalAddr.String()] = struct{}{}
	}

	IPisAvailable := func(ip net.IP) bool {
		_, exists := assignedIPs[ip.String()]
		return !exists
	}

	currentIP := net.ParseIP(mgr.config.Network.IPRange.Start)
	rangeEnd := net.ParseIP(mgr.config.Network.IPRange.End)

	for ; ; ipLessThanOrEqual(currentIP, rangeEnd) {
		if IPisAvailable(currentIP) {
			return currentIP
		}
		currentIP = incIP(currentIP)
	}

	panic(errors.New("unable to get a free ip"))
	return net.IP{}
}

func (mgr *pxeManagerT) thisHost() string {
	scheme := "https"
	if mgr.noTLS {
		scheme = "http"
	}

	return fmt.Sprintf("%s://%s:%d", scheme, mgr.config.Network.BindAddr, mgr.httpPort)
}

func httpError(w http.ResponseWriter, msg string, status int) {
	glog.Warningln(msg)
	http.Error(w, msg, status)
}
