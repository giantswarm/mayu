package hostmgr

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"crypto/tls"
	"crypto/x509"
	"errors"
	"github.com/coreos/etcd/client"
	"github.com/golang/glog"
	"golang.org/x/net/context"
)

const clusterConfFile = "cluster.json"

type Cluster struct {
	GitStore bool
	Config   ClusterConfig

	baseDir string

	// an cached host is identified by its serial number
	hostsCache    map[string]*cachedHost
	cachedModTime time.Time
	mu            *sync.Mutex

	// indexes
	hostByInternalAddr map[string]*cachedHost
	hostByMacAddr      map[string]*cachedHost

	// z
	predefinedVals map[string]map[string]string
}

type ClusterConfig struct {
	DefaultEtcdClusterToken string

	// Deprecated
	EtcdDiscoveryURL string `json:"EtcdDiscoveryURL,omitempty"`
}

type cachedHost struct {
	lastModTime time.Time
	host        *Host
}

func OpenCluster(baseDir string) (*Cluster, error) {
	cluster := &Cluster{}

	err := loadJson(cluster, path.Join(baseDir, clusterConfFile))
	if err != nil {
		return nil, err
	}

	cluster.baseDir = baseDir
	cluster.mu = new(sync.Mutex)
	cluster.predefinedVals = map[string]map[string]string{}
	cluster.hostsCache = map[string]*cachedHost{}
	return cluster, nil
}

// NewCluster creates a new cluster based on the cluster directory. gitStore
// defines whether cluster changes should be tracked using version control or
// not.
func NewCluster(baseDir string, gitStore bool) (*Cluster, error) {
	if !fileExists(baseDir) {
		err := os.Mkdir(baseDir, 0755)
		if err != nil {
			return nil, err
		}
	}

	if gitStore && !isGitRepo(baseDir) {
		err := gitInit(baseDir)
		if err != nil {
			return nil, err
		}
	}

	c := &Cluster{
		baseDir:        baseDir,
		GitStore:       gitStore,
		mu:             new(sync.Mutex),
		Config:         ClusterConfig{},
		predefinedVals: map[string]map[string]string{},
		hostsCache:     map[string]*cachedHost{},
	}

	return c, c.Commit("initial commit")
}

// CreateNewHost creates a new host with the given serial.
func (c *Cluster) CreateNewHost(serial string) (*Host, error) {
	serial = strings.ToLower(serial)
	hostDir := path.Join(c.baseDir, strings.ToLower(serial))
	newHost, err := createHost(serial, hostDir)
	if err != nil {
		return nil, err
	}

	var cabinet, machineOnCabinet uint

	if predef, exists := c.predefinedVals[serial]; exists {
		glog.V(2).Infof("found predefined values for '%s'", serial)
		if s, exists := predef["ipmiaddr"]; exists {
			newHost.IPMIAddr = net.ParseIP(s)
			glog.V(4).Infof("setting IPMIAdddress for '%s': %s", serial, newHost.IPMIAddr.String())
		}
		if s, exists := predef["cabinet"]; exists {
			num, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				return nil, err
			}
			cabinet = uint(num)
		}
		if s, exists := predef["machineoncabinet"]; exists {
			num, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				return nil, err
			}
			machineOnCabinet = uint(num)
		}
		if s, exists := predef["internaladdr"]; exists {
			newHost.InternalAddr = net.ParseIP(s)
			glog.V(4).Infof("setting internal address for '%s': %s", serial, newHost.InternalAddr.String())
		}
		if s, exists := predef["fleettags"]; exists {
			newHost.FleetMetadata = strings.Split(s, ",")
		}
		if s, exists := predef["etcdclustertoken"]; exists {
			newHost.EtcdClusterToken = s
		}
	} else {
		glog.V(2).Infof("no predefined values for '%s'", serial)
	}

	machineID := genMachineID(cabinet, machineOnCabinet)
	newHost.Cabinet = cabinet
	newHost.MachineOnCabinet = machineOnCabinet
	newHost.MachineID = machineID
	newHost.Hostname = machineID[:16]
	newHost.Commit("updated with predefined settings")

	return newHost, c.reindex()
}

func (c *Cluster) Commit(msg string) error {
	if err := c.save(); err != nil {
		return err
	}

	if c.GitStore {
		return gitAddCommit(c.baseDir, c.confPath(), msg)
	}
	return nil
}

// Update refreshs the internal host cache based on information within the
// cluster directory.
func (c *Cluster) Update() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.cacheHosts(); err != nil {
		return err
	}
	return c.reindex()
}

// HostWithMacAddress returns the host object given by macAddr based on the
// internal cache. In case the host could not be found, host is nil and false
// is returned as second return value.
func (c *Cluster) HostWithMacAddress(macAddr string) (*Host, bool) {
	if err := c.Update(); err != nil {
		fmt.Errorf("error getting the mac address using the internal cache %v", err)
		return nil, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	if cached, exists := c.hostByMacAddr[strings.ToLower(macAddr)]; exists {
		return cached.get(), true
	} else {
		return nil, false
	}
}

// HostWithInternalAddr returns the host object given by ipAddr based on the
// internal cache. In case the host could not be found, host is nil and false
// is returned as second return value.
func (c *Cluster) HostWithInternalAddr(ipAddr net.IP) (*Host, bool) {
	if err := c.Update(); err != nil {
		fmt.Errorf("error getting the ip address using the internal cache %v", err)
		return nil, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	if cached, exists := c.hostByInternalAddr[ipAddr.String()]; exists {
		return cached.get(), true
	} else {
		return nil, false
	}
}

// HostWithSerial returns the host object given by serial based on the internal
// cache. In case the host could not be found, host is nil and false is
// returned as second return value.
func (c *Cluster) HostWithSerial(serial string) (*Host, bool) {
	if err := c.Update(); err != nil {
		fmt.Errorf("error getting the serial number using the internal cache %v", err)
		return nil, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	if cached, exists := c.hostsCache[strings.ToLower(serial)]; exists {
		return cached.get(), true
	} else {
		return nil, false
	}
}

// GetProfileCount returns a matching of profiles and how many of them are
// known to the cluster. Imagine there is a provile name core. If there are 2
// core nodes known to the cluster, the map would look like this.
//
//   map[string]int{
//     "core": 2,
//   }
//
func (c *Cluster) GetProfileCount() map[string]int {
	count := map[string]int{}
	allHosts := c.GetAllHosts()
	for _, host := range allHosts {
		if host.Profile == "" {
			continue
		}
		if cnt, exists := count[host.Profile]; exists {
			count[host.Profile] = cnt + 1
		} else {
			count[host.Profile] = 1
		}
	}
	return count
}

// GetAllHosts returns a list of all hosts based on the internal cache.
func (c *Cluster) GetAllHosts() []*Host {
	hosts := make([]*Host, 0, len(c.hostsCache))

	if err := c.Update(); err != nil {
		fmt.Errorf("error getting the list of hosts based on the internal cache %v", err)
		return hosts
	}

	for _, cachedHost := range c.hostsCache {
		hosts = append(hosts, cachedHost.get())
	}
	return hosts
}

func (c *Cluster) FilterHostsFunc(predicate func(*Host) bool) chan *Host {
	ch := make(chan *Host)

	if err := c.Update(); err != nil {
		fmt.Errorf("error filtering the hosts %v", err)
		return ch
	}

	go func() {
		for _, cachedHost := range c.hostsCache {
			host := cachedHost.get()
			if predicate(host) {
				ch <- host
			}
		}
		close(ch)
	}()

	return ch
}

func (c *Cluster) GenerateEtcdDiscoveryToken() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	return token, nil
}

func (c *Cluster) StoreEtcdDiscoveryToken(etcdEndpoint, etcdCAFile, token string, size int) error {
	//http transport for etcd connection
	transport := client.DefaultTransport
	// read custom root CA file if https and CAfile is configured
	if strings.HasPrefix(etcdEndpoint, "https") && etcdCAFile != "" {
		customCA := x509.NewCertPool()

		pemData, err := ioutil.ReadFile(etcdCAFile)
		if err != nil {
			return errors.New("Unable to read custom CA file: " + err.Error())
		}
		customCA.AppendCertsFromPEM(pemData)
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: customCA},
			Proxy:           http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 10 * time.Second,
		}
	}

	// store in etcd
	cfg := client.Config{
		Endpoints: []string{etcdEndpoint},
		Transport: transport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	etcdClient, err := client.New(cfg)
	if err != nil {
		return err
	}
	kapi := client.NewKeysAPI(etcdClient)

	_, err = kapi.Set(context.Background(), path.Join("_etcd", "registry", token), "", &client.SetOptions{
		PrevExist: client.PrevNoExist,
		Dir:       true,
	})
	if err != nil {
		return err
	}

	_, err = kapi.Set(context.Background(), path.Join("_etcd", "registry", token, "_config", "size"), strconv.Itoa(size), &client.SetOptions{
		PrevExist: client.PrevNoExist,
	})
	return err
}

func (c *Cluster) FetchEtcdDiscoveryToken(etcdDiscoveryUrl string, size int) (string, error) {
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/new", etcdDiscoveryUrl), strings.NewReader(fmt.Sprintf("size=%d", size)))
	if err != nil {
		return "", err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	token := strings.TrimPrefix(string(body), etcdDiscoveryUrl+"/")
	return token, nil
}

func Has(host *Host, exists bool) bool {
	return exists
}

func (cached cachedHost) get() *Host {
	fi, err := os.Stat(cached.host.confPath())
	if err != nil {
		panic(err)
	}

	if fi.ModTime().After(cached.lastModTime) {
		hostDir := cached.host.hostDir.Name()
		cached.host, err = HostFromDir(hostDir)
		if err != nil {
			panic(err)
		}
		cached.lastModTime = cached.host.lastModTime
	}

	return cached.host
}

func (c *Cluster) save() error {
	return saveJson(c, c.confPath())
}

func (c *Cluster) confPath() string {
	return path.Join(c.baseDir, clusterConfFile)
}

func (c *Cluster) cacheHosts() error {
	baseDirFileInfo, err := os.Stat(c.baseDir)
	if err != nil {
		return err
	}

	modTime := baseDirFileInfo.ModTime()

	fis, err := ioutil.ReadDir(c.baseDir)
	if err != nil {
		return err
	}

	newCache := map[string]*cachedHost{}

	for _, fi := range fis {
		if fi.IsDir() && !strings.HasPrefix(fi.Name(), ".") {
			hostConfPath := path.Join(c.baseDir, fi.Name(), hostConfFile)
			if fileExists(hostConfPath) {
				host, err := HostFromDir(path.Join(c.baseDir, fi.Name()))
				if err != nil {
					glog.Warningf("unable to process '%s': %s", hostConfPath, err)
				}
				newCache[strings.ToLower(fi.Name())] = &cachedHost{
					host:        host,
					lastModTime: host.lastModTime,
				}
			} else {
				glog.V(4).Infof("file '%s' doesn't exist, skipping directory '%s'", hostConfPath, fi.Name())
			}

		}
	}

	c.hostsCache = newCache
	c.cachedModTime = modTime
	return c.reindex()
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func (c *Cluster) reindex() error {
	c.hostByInternalAddr = map[string]*cachedHost{}
	c.hostByMacAddr = map[string]*cachedHost{}

	for _, cached := range c.hostsCache {
		host := cached.get()

		c.hostByInternalAddr[host.InternalAddr.String()] = cached
		for _, macAddr := range host.MacAddresses {
			c.hostByMacAddr[strings.ToLower(macAddr)] = cached
		}
	}

	return nil
}
