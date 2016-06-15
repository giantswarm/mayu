package hostmgr

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const hostConfFile = "conf.json"

// Host represents a node within the mayu cluster.
type Host struct {
	Id                 int       `json:",omitempty"`
	ProviderId         string    `json:",omitempty"`
	Enabled            bool      `json:",omitempty"`
	Name               string    `json:",omitempty"`
	Serial             string    `json:",omitempty"`
	MacAddresses       []string  `json:",omitempty"`
	InternalAddr       net.IP    `json:",omitempty"`
	BondInterfaces     []string  `json:",omitempty"`
	Cabinet            uint      `json:",omitempty"`
	MachineOnCabinet   uint      `json:",omitempty"`
	IPMIAddr           net.IP    `json:",omitempty"`
	Hostname           string    `json:",omitempty"`
	MachineID          string    `json:",omitempty"`
	ConnectedNIC       string    `json:",omitempty`
	FleetMetadata      FleetMeta `json:",omitempty"`
	FleetDisableEngine bool      `json:",omitempty"`
	KeepDiskData       bool      `json:",omitempty"`
	LastBoot           time.Time `json:",omitempty"`
	Profile            string    `json:",omitempty"`
	EtcdClusterToken   string    `json:",omitempty"`

	Overrides map[string]interface{} `json:",omitempty"`

	State hostState

	CoreOSVersion string `json:",omitempty"`
	MayuVersion   string `json:",omitempty"`
	DockerVersion string `json:",omitempty"`
	FleetVersion  string `json:",omitempty"`
	EtcdVersion   string `json:",omitempty"`
	RktVersion    string `json:",omitempty"`
	K8sVersion    string `json:",omitempty"`
	YochuVersion  string `json:",omitempty"`

	hostDir     *os.File
	lastModTime time.Time
}

type FleetMeta []string

type IPMac struct {
	IP      net.IP
	MacAddr string
}

func (fleetMeta FleetMeta) Has(m string) bool {
	for _, v := range fleetMeta {
		if strings.ToLower(m) == strings.ToLower(v) {
			return true
		}
	}
	return false
}

func (fleetMeta FleetMeta) String() string {
	return strings.Join(fleetMeta, ",")
}

// Commit stores the given msg in git version control.
func (h *Host) Commit(msg string) error {
	h.save()
	return h.maybeGitCommit(h.Serial + ": " + msg)
}

// HostFromDir takes a path to a host directory within the cluster directory
// and loads the found configuration. Then the corresponding Host is returned.
func HostFromDir(hostdir string) (*Host, error) {
	confPath := path.Join(hostdir, hostConfFile)

	h := &Host{}
	err := loadJson(h, confPath)
	if err != nil {
		return nil, err
	}

	h.hostDir, err = os.Open(hostdir)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(confPath)
	if err != nil {
		return nil, err
	}
	h.lastModTime = fi.ModTime()

	return h, nil
}

func genMachineID(cabinet, machineOnCabinet uint) string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	b[0] = byte(0xff & cabinet)
	b[1] = byte(0xff & machineOnCabinet)

	return hex.EncodeToString(b)
}

func createHost(serial string, hostDir string) (*Host, error) {
	if !fileExists(hostDir) {
		os.Mkdir(hostDir, 0755)
	}

	hostDirFile, err := os.Open(hostDir)
	if err != nil {
		return nil, err
	}

	h := &Host{
		hostDir: hostDirFile,
		Serial:  strings.ToLower(serial),

		Enabled: true,
	}
	err = h.Commit("host created")
	return h, err
}

func (h *Host) save() error {
	if err := saveJson(h, h.confPath()); err == nil {
		if fi, err := os.Stat(h.confPath()); err == nil {
			h.lastModTime = fi.ModTime()
		} else {
			return err
		}
	} else {
		return err
	}
	return nil
}

func (h *Host) confPath() string {
	return path.Join(h.hostDir.Name(), hostConfFile)
}

func (h *Host) maybeGitCommit(msg string) error {
	absHostDir, err := filepath.Abs(h.hostDir.Name())
	if err != nil {
		return err
	}
	clusterDir := filepath.Clean(filepath.Join(absHostDir, ".."))
	if isGitRepo(clusterDir) {
		gitAddCommit(clusterDir, h.confPath(), msg)
	}
	return nil
}
