package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/giantswarm/mayu/hostmgr"
	"github.com/giantswarm/mayu/logging"
	"github.com/golang/glog"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type pxeManagerT struct {
	cluster *hostmgr.Cluster
	DNSmasq *DNSmasqInstance

	mu *sync.Mutex

	router *mux.Router
}

const defaultEtcdQuorumSize = 3

func defaultPXEManager(cluster *hostmgr.Cluster) (*pxeManagerT, error) {
	mgr := &pxeManagerT{
		cluster: cluster,
		DNSmasq: NewDNSmasq("/tmp/dnsmasq.mayu", conf),
		mu:      new(sync.Mutex),
	}

	if mgr.cluster.Config.EtcdDiscoveryURL == "" {
		mgr.cluster.GenerateEtcdDiscoveryURL(defaultEtcdQuorumSize)
		mgr.cluster.Commit("generated etcd discovery url")
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
	mgr.router.Methods("GET").PathPrefix("/ipxebootscript").HandlerFunc(ipxeBootScript)
	mgr.router.Methods("GET").PathPrefix("/first-stage-script/{serial}").HandlerFunc(mgr.firstStageScriptGenerator)

	// used by the first-stage-script:
	mgr.router.Methods("GET").PathPrefix("/hostinfo-helper").HandlerFunc(mgr.infoPusher)
	mgr.router.Methods("POST").PathPrefix("/final-cloud-config.yaml").HandlerFunc(mgr.cloudConfigGenerator)

	mgr.router.Methods("PUT").PathPrefix("/admin/host/{serial}/boot_complete").HandlerFunc(withSerialParam(mgr.bootComplete))
	mgr.router.Methods("PUT").PathPrefix("/admin/host/{serial}/set_installed").HandlerFunc(withSerialParam(mgr.setInstalled))
	mgr.router.Methods("PUT").PathPrefix("/admin/host/{serial}/set_metadata").HandlerFunc(withSerialParam(mgr.setMetadata))
	mgr.router.Methods("PUT").PathPrefix("/admin/host/{serial}/mark_fresh").HandlerFunc(withSerialParam(mgr.markFresh))
	mgr.router.Methods("PUT").PathPrefix("/admin/host/{serial}/set_provider_id").HandlerFunc(withSerialParam(mgr.setProviderId))
	mgr.router.Methods("PUT").PathPrefix("/admin/host/{serial}/set_ipmi_addr").HandlerFunc(withSerialParam(mgr.setIPMIAddr))
	mgr.router.Methods("PUT").PathPrefix("/admin/host/{serial}/set_cabinet").HandlerFunc(withSerialParam(mgr.setCabinet))

	// boring stuff
	mgr.router.Methods("GET").PathPrefix("/admin/hosts").HandlerFunc(mgr.hostsList)
	mgr.router.Methods("GET").PathPrefix("/images").HandlerFunc(imagesHandler)

	// add welcome handler for debugging
	mgr.router.Path("/").HandlerFunc(mgr.welcomeHandler)

	// serve static files like yochu, fleet, etc.
	mgr.router.PathPrefix("/").Handler(http.FileServer(http.Dir(conf.StaticHTMLPath)))

	glogWrapper := logging.NewGlogWrapper(8)
	loggedRouter := handlers.LoggingHandler(glogWrapper, mgr.router)

	glog.V(8).Infoln(fmt.Sprintf("starting iPXE server at %s:%d", conf.HTTPBindAddr, conf.HTTPPort))

	if conf.NoSecure {
		err := http.ListenAndServe(fmt.Sprintf("%s:%d", conf.HTTPBindAddr, conf.HTTPPort), loggedRouter)
		if err != nil {
			return err
		}
	} else {
		err := http.ListenAndServeTLS(fmt.Sprintf("%s:%d", conf.HTTPBindAddr, conf.HTTPPort), conf.HTTPSCertFile, conf.HTTPSKeyFile, loggedRouter)
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
	conf.Network.StaticHosts = []hostmgr.IPMac{}
	conf.Network.IgnoredHosts = []string{}

	ignoredHostPredicate := func(host *hostmgr.Host) bool {
		// ignore hosts that are installed or running
		return host.State == hostmgr.Installed || host.State == hostmgr.Running
	}

	for host := range mgr.cluster.FilterHostsFunc(ignoredHostPredicate) {
		for _, macAddr := range host.MacAddresses {
			conf.Network.IgnoredHosts = append(conf.Network.IgnoredHosts, macAddr)
		}
	}

	err := mgr.DNSmasq.updateConf(conf.Network)
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

	for _, profile := range conf.Profiles {
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

	currentIP := net.ParseIP(conf.Network.IPRange.Start)
	rangeEnd := net.ParseIP(conf.Network.IPRange.End)

	for ; ; ipLessThanOrEqual(currentIP, rangeEnd) {
		if IPisAvailable(currentIP) {
			return currentIP
		}
		currentIP = incIP(currentIP)
	}

	panic(errors.New("unable to get a free ip"))
	return net.IP{}
}
