package pxemgr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/mayu-infopusher/machinedata"
	"github.com/giantswarm/mayu/hostmgr"
)

const (
	defaultProfileName = "default"

	kvmStaticSerial = "0123456789"

	vmwareIdentifier = "VMware"
)

func (mgr *pxeManagerT) ipxeBootScript(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	buffer := bytes.NewBufferString("")
	extraFlags := ""
	if mgr.consoleTTY {
		extraFlags += " console=ttyS0"
		_ = mgr.logger.Log("level", "info", "message", "adding 'console=ttyS0' to kernel args")
	}

	if mgr.flatcarAutologin {
		extraFlags += " flatcar.autologin"
		_ = mgr.logger.Log("level", "info", "message", "adding flatcar.autologin to kernel args")
	}

	if mgr.systemdShell {
		extraFlags += " rd.shell"
		_ = mgr.logger.Log("level", "info", "message", "adding rd.shell to kernel args")
	}

	// for ignition we use only 1phase installation without mayu-infopusher
	kernel := fmt.Sprintf("kernel %s/images/vmlinuz flatcar.first_boot=1 initrd=initrd.cpio.gz flatcar.config.url=%s?uuid=${uuid}&serial=${serial} systemd.journald.max_level_console=debug verbose log_buf_len=10M "+extraFlags+"\n", mgr.pxeURL(), mgr.ignitionURL())
	initrd := fmt.Sprintf("initrd %s/images/initrd.cpio.gz\n", mgr.pxeURL())
	// console=ttyS0,115200n8
	buffer.WriteString("#!ipxe\n")
	buffer.WriteString("dhcp\n")
	buffer.WriteString(kernel)
	buffer.WriteString(initrd)
	buffer.WriteString("boot\n")

	_, _ = w.Write(buffer.Bytes())
}

func (mgr *pxeManagerT) maybeCreateHost(serial string) (*hostmgr.Host, error) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	host, exists := mgr.cluster.HostWithSerial(serial)
	if !exists {
		var err error
		host, err = mgr.cluster.CreateNewHost(serial)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		if host.InternalAddr == nil {
			host.InternalAddr = mgr.getNextInternalIP()
		}
		// generate addresses for the extra NICs
		host.AdditionalAddrs = make(map[string]net.IP)
		for i, nic := range mgr.config.Network.ExtraNICs {
			host.AdditionalAddrs[nic.InterfaceName] = mgr.getNextAdditionalIP(i)
		}
		if host.Profile == "" {
			host.Profile = mgr.getNextProfile()
			if host.Profile == "" {
				host.Profile = defaultProfileName
			}
			host.FlatcarVersion = mgr.config.DefaultFlatcarVersion
		}
		if host.EtcdClusterToken == "" {
			host.EtcdClusterToken = mgr.cluster.Config.DefaultEtcdClusterToken
		}
		if host.InternalAddr != nil {
			host.Hostname = strings.Replace(host.InternalAddr.String(), ".", "-", 4)
		}

		err = host.Save()
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}
	return host, nil
}

func (mgr *pxeManagerT) ignitionGenerator(w http.ResponseWriter, r *http.Request) {
	uuid := r.URL.Query().Get("uuid")
	serial := r.URL.Query().Get("serial")

	hostData := &machinedata.HostData{}

	// If there is no reliable serial then use uuid for identification of machine.
	// Case 1: serial from kvm vm is static and not unique so we need to use uuid.
	// Case 2: serial sent by ipxe from vmware machines is truncated and not unique so we need to use uuid.
	if serial == "" || serial == kvmStaticSerial || strings.Contains(serial, vmwareIdentifier) {
		hostData.Serial = uuid
	} else {
		hostData.Serial = serial
	}

	if hostData.Serial == "" {
		_ = mgr.logger.Log("level", "error", "message", fmt.Sprintf("empty serial. %+v\n", hostData))
		w.WriteHeader(400)
		_, _ = w.Write([]byte("no serial ? :/"))
		return
	}

	host, err := mgr.maybeCreateHost(hostData.Serial)
	if err != nil {
		_ = mgr.logger.Log("level", "error", "message", fmt.Sprintf("failed to create machine host %+v\n", hostData))
	}
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	_ = mgr.logger.Log("level", "info", "message", fmt.Sprintf("got host %+v\n", host))

	host.State = hostmgr.Installing
	host.Hostname = strings.Replace(host.InternalAddr.String(), ".", "-", 4)
	_ = host.Save()

	_ = mgr.cluster.Update()

	buf := &bytes.Buffer{}
	_ = mgr.logger.Log("level", "info", "message", "generating a ignition config")

	if err := mgr.WriteIgnitionConfig(*host, buf); err != nil {
		w.WriteHeader(500)
		_, _ = w.Write([]byte("generating ignition config failed: " + err.Error()))

		_ = mgr.logger.Log("level", "error", "message", "generating ignition config failed", "stack", err)
		return
	}
	if _, err := buf.WriteTo(w); err != nil {
		_ = mgr.logger.Log("level", "error", "message", "failed to write response", "stack", err)
	}
}

func (mgr *pxeManagerT) imagesHandler(w http.ResponseWriter, r *http.Request) {
	var img *os.File
	var err error

	flatcarVersion := mgr.config.DefaultFlatcarVersion
	_ = mgr.logger.Log("level", "info", "message", fmt.Sprintf("sending Container Linux %s image", flatcarVersion))

	if strings.HasSuffix(r.URL.Path, "/vmlinuz") {
		img, err = mgr.pxeKernelImage(flatcarVersion)
	} else if strings.HasSuffix(r.URL.Path, "/initrd.cpio.gz") {
		img, err = mgr.pxeInitRD(flatcarVersion)
	} else {
		panic(fmt.Sprintf("no handler provided for invalid URL path '%s'", r.URL.Path))
	}

	if err != nil {
		panic(err)
	}

	_ = setContentLength(w, img)
	defer img.Close()
	_, _ = io.Copy(w, img)
}

func setContentLength(w http.ResponseWriter, f *os.File) error {
	fi, err := f.Stat()
	if err != nil {
		return microerror.Mask(err)
	}
	w.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))
	return nil
}

func (mgr *pxeManagerT) hostsList(w http.ResponseWriter, r *http.Request) {
	hosts := mgr.cluster.GetAllHosts()

	w.WriteHeader(200)
	enc := json.NewEncoder(w)
	_ = enc.Encode(hosts)
}

func (mgr *pxeManagerT) bootComplete(serial string, w http.ResponseWriter, r *http.Request) {
	host, exists := mgr.cluster.HostWithSerial(serial)
	if !exists {
		w.WriteHeader(400)
		_, _ = w.Write([]byte("host doesn't exist"))
		return
	}

	_ = mgr.logger.Log("level", "info", "message", fmt.Sprintf("host '%s' just finished booting", serial))

	decoder := json.NewDecoder(r.Body)
	payload := hostmgr.Host{}
	err := decoder.Decode(&payload)
	if err != nil {
		w.WriteHeader(400)
		_, _ = w.Write([]byte("unable to parse json data in boot_complete request"))
		return
	}

	host.State = hostmgr.Running
	host.LastBoot = time.Now()
	host.FlatcarVersion = payload.FlatcarVersion

	err = host.Save()
	if err != nil {
		w.WriteHeader(500)
		_, _ = w.Write([]byte("committing updated host state=running failed"))
		return
	}
	_ = mgr.cluster.Update()
	w.WriteHeader(202)
}

func (mgr *pxeManagerT) setProviderId(serial string, w http.ResponseWriter, r *http.Request) {
	host, exists := mgr.cluster.HostWithSerial(serial)
	if !exists {
		w.WriteHeader(400)
		_, _ = w.Write([]byte("host doesn't exist"))
		return
	}

	decoder := json.NewDecoder(r.Body)
	payload := hostmgr.Host{}
	err := decoder.Decode(&payload)
	if err != nil {
		w.WriteHeader(400)
		_, _ = w.Write([]byte("unable to parse json data in set_provider_id request"))
		return
	}

	host.ProviderId = payload.ProviderId
	err = host.Save()
	if err != nil {
		w.WriteHeader(500)
		_, _ = w.Write([]byte("committing updated host provider id failed"))
		return
	}
	_ = mgr.cluster.Update()
	w.WriteHeader(202)
}

func (mgr *pxeManagerT) setIPMIAddr(serial string, w http.ResponseWriter, r *http.Request) {
	host, exists := mgr.cluster.HostWithSerial(serial)
	if !exists {
		w.WriteHeader(400)
		_, _ = w.Write([]byte("host doesn't exist"))
		return
	}

	decoder := json.NewDecoder(r.Body)
	payload := hostmgr.Host{}
	err := decoder.Decode(&payload)
	if err != nil {
		w.WriteHeader(400)
		_, _ = w.Write([]byte("unable to parse json data in set_ipmi_addr request"))
		return
	}

	host.IPMIAddr = payload.IPMIAddr
	err = host.Save()
	if err != nil {
		w.WriteHeader(500)
		_, _ = w.Write([]byte("committing updated host ipmi address failed"))
		return
	}
	_ = mgr.cluster.Update()
	w.WriteHeader(202)
}

func (mgr *pxeManagerT) setEtcdClusterToken(serial string, w http.ResponseWriter, r *http.Request) {
	host, exists := mgr.cluster.HostWithSerial(serial)
	if !exists {
		w.WriteHeader(400)
		_, _ = w.Write([]byte("host doesn't exist"))
		return
	}

	decoder := json.NewDecoder(r.Body)
	payload := hostmgr.Host{}
	err := decoder.Decode(&payload)
	if err != nil {
		w.WriteHeader(400)
		_, _ = w.Write([]byte("unable to parse json data in set_etcd_cluster_token request"))
		return
	}

	host.EtcdClusterToken = payload.EtcdClusterToken
	err = host.Save()
	if err != nil {
		w.WriteHeader(500)
		_, _ = w.Write([]byte("committing updated host etcd cluster token failed"))
		return
	}
	_ = mgr.cluster.Update()
	w.WriteHeader(202)
}

func (mgr *pxeManagerT) welcomeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	_, _ = w.Write([]byte("this is the iPXE server of mayu " + mgr.version))
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

	currentIP := net.ParseIP(mgr.config.Network.PrimaryNIC.IPRange.Start)
	rangeEnd := net.ParseIP(mgr.config.Network.PrimaryNIC.IPRange.End)

	for ; ; ipLessThanOrEqual(currentIP, rangeEnd) {
		if IPisAvailable(currentIP) {
			return currentIP
		}
		currentIP = incIP(currentIP)
	}
}

func (mgr *pxeManagerT) getNextAdditionalIP(nicIndex int) net.IP {
	nicName := mgr.config.Network.ExtraNICs[nicIndex].InterfaceName
	assignedIPs := map[string]struct{}{}
	for _, host := range mgr.cluster.GetAllHosts() {
		if _, exists := host.AdditionalAddrs[nicName]; exists {
			assignedIPs[host.AdditionalAddrs[nicName].String()] = struct{}{}
		}
	}

	IPisAvailable := func(ip net.IP) bool {
		_, exists := assignedIPs[ip.String()]
		return !exists
	}

	currentIP := net.ParseIP(mgr.config.Network.ExtraNICs[nicIndex].IPRange.Start)
	rangeEnd := net.ParseIP(mgr.config.Network.ExtraNICs[nicIndex].IPRange.End)

	for ; ; ipLessThanOrEqual(currentIP, rangeEnd) {
		if IPisAvailable(currentIP) {
			return currentIP
		}
		currentIP = incIP(currentIP)
	}
}

// check af all hosts have properly assigned IP addresses to all Network.ExtraNICs entries
func (mgr *pxeManagerT) checkAdditionalNICAddresses() {
	hosts := mgr.cluster.GetAllHosts()
	// sort the array so we have the host ordered by the internal IP
	// this will sort in a way how hosts are listed with mayuctl
	sort.SliceStable(hosts, func(i int, j int) bool {
		return hosts[i].InternalAddr.String() < hosts[j].InternalAddr.String()
	})
	for _, h := range hosts {
		// iterate over all extra NICs
		for i, nic := range mgr.config.Network.ExtraNICs {
			if h.AdditionalAddrs == nil {
				h.AdditionalAddrs = make(map[string]net.IP)
			}
			// check if the interface has already entry on the host config
			if ip, exists := h.AdditionalAddrs[nic.InterfaceName]; !exists {
				// no entry, lets add a new clean IP from the range
				h.AdditionalAddrs[nic.InterfaceName] = mgr.getNextAdditionalIP(i)
			} else {
				// host have assigned IP on this NIC, but we need to check if the network range matches
				ipStart := net.ParseIP(nic.IPRange.Start)
				ipEnd := net.ParseIP(nic.IPRange.End)
				if ipLessThanOrEqual(ipStart, ip) &&
					ipMoreThanOrEqual(ipEnd, ip) {
					// we should be good, IP is in the range of Start and End
					// DO nothing
				} else {
					// the ip is not in the range of Start and End
					// we need to clear this ip and assign a new one
					// this is tricky, this might not work as expected
					h.AdditionalAddrs[nic.InterfaceName] = mgr.getNextAdditionalIP(i)
				}
			}
		}
		_ = h.Save()
	}
}
