package hostmgr

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
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

	"github.com/coreos/etcd/client"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"golang.org/x/net/context"
)

const clusterConfFile = "cluster.json"

type Cluster struct {
	Config ClusterConfig

	baseDir string
	// an cached host is identified by its serial number
	hostsCache    map[string]*cachedHost
	cachedModTime time.Time
	mu            *sync.Mutex

	logger micrologger.Logger
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

func OpenCluster(baseDir string, logger micrologger.Logger) (*Cluster, error) {
	cluster := &Cluster{logger: logger}

	err := loadJson(cluster, path.Join(baseDir, clusterConfFile))
	if err != nil {
		return nil, microerror.Mask(err)
	}

	cluster.baseDir = baseDir
	cluster.mu = new(sync.Mutex)
	cluster.hostsCache = map[string]*cachedHost{}
	return cluster, nil
}

// NewCluster creates a new cluster based on the cluster directory.
func NewCluster(baseDir string, logger micrologger.Logger) (*Cluster, error) {
	if !fileExists(baseDir) {
		err := os.Mkdir(baseDir, 0755)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	c := &Cluster{
		baseDir:    baseDir,
		mu:         new(sync.Mutex),
		Config:     ClusterConfig{},
		hostsCache: map[string]*cachedHost{},
		logger:     logger,
	}

	err := c.Commit("initial commit")
	if err != nil {
		return nil, microerror.Mask(err)
	}
	return c, nil
}

// CreateNewHost creates a new host with the given serial.
func (c *Cluster) CreateNewHost(serial string) (*Host, error) {
	serial = strings.ToLower(serial)
	hostDir := path.Join(c.baseDir, serial)
	newHost, err := createHost(serial, hostDir)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	machineID := genMachineID()
	newHost.MachineID = machineID
	if newHost.InternalAddr != nil {
		newHost.Hostname = strings.Replace(newHost.InternalAddr.String(), ".", "-", 4)
	}
	_ = c.logger.Log("level", "info", "message", fmt.Sprintf("hostname for  '%s' is %s", newHost.InternalAddr.String(), newHost.Hostname))
	_ = newHost.Save()

	return newHost, nil
}

func (c *Cluster) Commit(msg string) error {
	err := c.save()
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

// Update refreshs the internal host cache based on information within the
// cluster directory.
func (c *Cluster) Update() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.cacheHosts()
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

// HostWithSerial returns the host object given by serial based on the internal
// cache. In case the host could not be found, host is nil and false is
// returned as second return value.
func (c *Cluster) HostWithSerial(serial string) (*Host, bool) {
	if err := c.Update(); err != nil {
		_ = c.logger.Log("level", "error", "message", "error getting the serial number using the internal cache", "stack", err)
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
		_ = c.logger.Log("level", "error", "message", "error getting the list of hosts based on the internal cache: %#v", "stack", err)
		return hosts
	}

	for _, cachedHost := range c.hostsCache {
		hosts = append(hosts, cachedHost.get())
	}
	return hosts
}

func (c *Cluster) GenerateEtcdDiscoveryToken() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", microerror.Mask(err)
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
			return microerror.Maskf(err, "Unable to read custom CA file: ")
		}
		customCA.AppendCertsFromPEM(pemData)
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:    customCA,
				MinVersion: tls.VersionTLS12,
			},
			Proxy: http.ProxyFromEnvironment,
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
		return microerror.Mask(err)
	}
	kapi := client.NewKeysAPI(etcdClient)

	_, err = kapi.Set(context.Background(), path.Join("_etcd", "registry", token), "", &client.SetOptions{
		PrevExist: client.PrevNoExist,
		Dir:       true,
	})
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = kapi.Set(context.Background(), path.Join("_etcd", "registry", token, "_config", "size"), strconv.Itoa(size), &client.SetOptions{
		PrevExist: client.PrevNoExist,
	})
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (c *Cluster) FetchEtcdDiscoveryToken(etcdDiscoveryUrl string, size int) (string, error) {
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/new", etcdDiscoveryUrl), strings.NewReader(fmt.Sprintf("size=%d", size)))
	if err != nil {
		return "", microerror.Mask(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", microerror.Mask(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", microerror.Mask(err)
	}

	token := strings.TrimPrefix(string(body), etcdDiscoveryUrl+"/")
	return token, nil
}

func Has(host *Host, exists bool) bool {
	return exists
}

func (cached *cachedHost) get() *Host {
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
		return microerror.Mask(err)
	}

	modTime := baseDirFileInfo.ModTime()

	fis, err := ioutil.ReadDir(c.baseDir)
	if err != nil {
		return microerror.Mask(err)
	}

	newCache := map[string]*cachedHost{}

	for _, fi := range fis {
		if fi.IsDir() && !strings.HasPrefix(fi.Name(), ".") {
			hostConfPath := path.Join(c.baseDir, fi.Name(), hostConfFile)
			if fileExists(hostConfPath) {
				host, err := HostFromDir(path.Join(c.baseDir, fi.Name()))
				if err != nil {
					_ = c.logger.Log("level", "warning", "message", fmt.Sprintf("unable to process '%s'", hostConfPath), "stack", err)
				}
				newCache[strings.ToLower(fi.Name())] = &cachedHost{
					host:        host,
					lastModTime: host.lastModTime,
				}
			}
		}
	}

	c.hostsCache = newCache
	c.cachedModTime = modTime
	return nil
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}
