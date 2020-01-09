package hostmgr

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/giantswarm/microerror"
	"net"
	"os"
	"path"
	"time"
)

const hostConfFile = "conf.json"

// Host represents a node within the mayu cluster.
type Host struct {
	Id               int               `json:",omitempty"`
	ProviderId       string            `json:",omitempty"`
	Enabled          bool              `json:",omitempty"`
	Name             string            `json:",omitempty"`
	Serial           string            `json:",omitempty"`
	MacAddresses     []string          `json:",omitempty"`
	InternalAddr     net.IP            `json:",omitempty"`
	AdditionalAddrs  map[string]net.IP `json:",omitempty"`
	IPMIAddr         net.IP            `json:",omitempty"`
	Hostname         string            `json:",omitempty"`
	MachineID        string            `json:",omitempty"`
	LastBoot         time.Time         `json:",omitempty"`
	Profile          string            `json:",omitempty"`
	EtcdClusterToken string            `json:",omitempty"`

	Overrides map[string]interface{} `json:",omitempty"`

	State hostState

	CoreOSVersion string `json:",omitempty"`

	hostDir     *os.File
	lastModTime time.Time
}

type IPMac struct {
	IP      net.IP
	MacAddr string
}

func genMachineID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

// HostFromDir takes a path to a host directory within the cluster directory
// and loads the found configuration. Then the corresponding Host is returned.
func HostFromDir(hostdir string) (*Host, error) {
	confPath := path.Join(hostdir, hostConfFile)

	h := &Host{}
	err := loadJson(h, confPath)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	h.hostDir, err = os.Open(hostdir)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	fi, err := os.Stat(confPath)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	h.lastModTime = fi.ModTime()

	return h, nil
}

func createHost(serial string, hostDir string) (*Host, error) {
	var err error
	if !fileExists(hostDir) {
		err = os.Mkdir(hostDir, 0755)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	hostDirFile, err := os.Open(hostDir)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	h := &Host{
		hostDir: hostDirFile,
		Serial:  serial,
		Enabled: true,
	}
	err = h.Save()
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return h, nil
}

func (h *Host) Save() error {
	err := saveJson(h, h.confPath())
	if err != nil {
		return microerror.Mask(err)
	}

	fi, err := os.Stat(h.confPath())
	if err != nil {
		return microerror.Mask(err)
	}

	h.lastModTime = fi.ModTime()
	return nil
}

func (h *Host) confPath() string {
	return path.Join(h.hostDir.Name(), hostConfFile)
}
