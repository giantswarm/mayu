package pxemgr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/giantswarm/mayu-infopusher/machinedata"
	"github.com/giantswarm/mayu/hostmgr"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
	"github.com/giantswarm/microerror"
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
	if mgr.coreosAutologin {
		extraFlags += "coreos.autologin"
		glog.V(2).Infoln("adding coreos.autologin to kernel args")
	}

	// for ignition we use only 1phase installation without mayu-infopusher
	kernel := fmt.Sprintf("kernel %s/images/vmlinuz coreos.first_boot=1 initrd=initrd.cpio.gz coreos.config.url=%s?uuid=${uuid}&serial=${serial} systemd.journald.max_level_console=debug verbose log_buf_len=10M "+extraFlags+"\n", mgr.pxeURL(), mgr.ignitionURL())
	initrd := fmt.Sprintf("initrd %s/images/initrd.cpio.gz\n", mgr.pxeURL())
	// console=ttyS0,115200n8
	buffer.WriteString("#!ipxe\n")
	buffer.WriteString("dhcp\n")
	buffer.WriteString(kernel)
	buffer.WriteString(initrd)
	buffer.WriteString("boot\n")

	w.Write(buffer.Bytes())
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
			err = host.Commit("updated host InternalAddr")
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}

		if host.Profile == "" {
			host.Profile = mgr.getNextProfile()
			if host.Profile == "" {
				host.Profile = defaultProfileName
			}
			host.CoreOSVersion = mgr.profileCoreOSVersion(host.Profile)
			host.EtcdClusterToken = mgr.profileEtcdClusterToken(host.Profile)

			err = host.Commit("updated host profile and metadata")
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}

		if host.EtcdClusterToken == "" {
			host.EtcdClusterToken = mgr.cluster.Config.DefaultEtcdClusterToken
			err = host.Commit("set default etcd discovery token")
			if err != nil {
				return nil, microerror.Mask(err)
			}
		}
		if host.InternalAddr != nil {
			host.Hostname = strings.Replace(host.InternalAddr.String(), ".", "-", 4)
		}
	}
	return host, nil
}

func (mgr *pxeManagerT) profileCoreOSVersion(profileName string) string {
	for _, v := range mgr.config.Profiles {
		if v.Name == profileName {
			return v.CoreOSVersion
		}
	}
	return mgr.config.DefaultCoreOSVersion
}

func (mgr *pxeManagerT) profileEtcdClusterToken(profileName string) string {
	for _, v := range mgr.config.Profiles {
		if v.Name == profileName {
			return v.EtcdClusterToken
		}
	}
	return ""
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
		mgr.logger.Log("level", "error", "msg", fmt.Sprintf("empty serial. %+v\n", hostData))

		w.WriteHeader(400)
		w.Write([]byte("no serial ? :/"))
		return
	}

	host, err := mgr.maybeCreateHost(hostData.Serial)
	if err != nil {
		mgr.logger.Log("level", "error", "msg", fmt.Sprintf("failed to create machine host %+v\n", hostData))
	}
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	mgr.logger.Log("level", "info", "msg", fmt.Sprintf("got host %+v\n", host))

	host.State = hostmgr.Installing
	host.Hostname = strings.Replace(host.InternalAddr.String(), ".", "-", 4)
	host.Commit("updated host state to installing")
	mgr.cluster.Update()

	buf := &bytes.Buffer{}

	mgr.logger.Log("level", "info", "msg", "generating a ignition config")

	if err := mgr.WriteIgnitionConfig(*host, buf); err != nil {
		w.WriteHeader(500)
		w.Write([]byte("generating ignition config failed: " + err.Error()))

		mgr.logger.Log("level", "error", "msg", fmt.Sprintf("generating ignition config failed: %s", err))
		return
	}
	if _, err := buf.WriteTo(w); err != nil {
		mgr.logger.Log("level", "error", "msg", fmt.Sprintf("failed to write response: %s", err))
	}
}

func (mgr *pxeManagerT) imagesHandler(w http.ResponseWriter, r *http.Request) {
	coreOSversion := mgr.hostCoreOSVersion(r)

	mgr.logger.Log("level", "info", "msg", fmt.Sprintf("sending Container Linux %s image", coreOSversion))

	var (
		img *os.File
		err error
	)

	if strings.HasSuffix(r.URL.Path, fmt.Sprintf("/qemu/%s.sha256", qemuImageFile)) {
		img, err = mgr.qemuImageSHA(coreOSversion)
	} else if strings.HasSuffix(r.URL.Path, fmt.Sprintf("/qemu/%s", qemuImageFile)) {
		img, err = mgr.qemuImage(coreOSversion)
	} else if strings.HasSuffix(r.URL.Path, fmt.Sprintf("/qemu/%s.sha256", qemuKernelFile)) {
		img, err = mgr.qemuKernelSHA(coreOSversion)
	} else if strings.HasSuffix(r.URL.Path, fmt.Sprintf("/qemu/%s", qemuKernelFile)) {
		img, err = mgr.qemuKernel(coreOSversion)

	} else if strings.HasSuffix(r.URL.Path, "/vmlinuz") {
		img, err = mgr.pxeKernelImage(coreOSversion)
	} else if strings.HasSuffix(r.URL.Path, "/initrd.cpio.gz") {
		img, err = mgr.pxeInitRD(coreOSversion)
	} else if strings.HasSuffix(r.URL.Path, "/install_image.bin.bz2") {
		img, err = mgr.pxeInstallImage(coreOSversion)
	} else {
		panic(fmt.Sprintf("no handler provided for invalid URL path '%s'", r.URL.Path))
	}

	if err != nil {
		panic(err)
	}

	setContentLength(w, img)
	defer img.Close()
	io.Copy(w, img)
}

func setContentLength(w http.ResponseWriter, f *os.File) error {
	fi, err := f.Stat()
	if err != nil {
		return microerror.Mask(err)
	}
	w.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))
	return nil
}

func (mgr *pxeManagerT) markFresh(serial string, w http.ResponseWriter, r *http.Request) {
	host, exists := mgr.cluster.HostWithSerial(serial)
	if !exists {
		w.WriteHeader(400)
		w.Write([]byte("host doesn't exist"))
		return
	}

	host.State = hostmgr.Configured
	host.Commit("host flagged as fresh")
	mgr.cluster.Update()

	w.WriteHeader(202)
}

func (mgr *pxeManagerT) hostsList(w http.ResponseWriter, r *http.Request) {
	hosts := mgr.cluster.GetAllHosts()

	w.WriteHeader(200)
	enc := json.NewEncoder(w)
	enc.Encode(hosts)
}

func (mgr *pxeManagerT) hostFromSerialVar(r *http.Request) (*hostmgr.Host, bool) {
	params := mux.Vars(r)
	serial := strings.ToLower(params["serial"])

	return mgr.cluster.HostWithSerial(serial)
}

func (mgr *pxeManagerT) bootComplete(serial string, w http.ResponseWriter, r *http.Request) {
	host, exists := mgr.cluster.HostWithSerial(serial)
	if !exists {
		w.WriteHeader(400)
		w.Write([]byte("host doesn't exist"))
		return
	}

	mgr.logger.Log("level", "info", "msg", fmt.Sprintf("host '%s' just finished booting", serial))

	decoder := json.NewDecoder(r.Body)
	payload := hostmgr.Host{}
	err := decoder.Decode(&payload)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("unable to parse json data in boot_complete request"))
		return
	}

	host.State = hostmgr.Running
	host.LastBoot = time.Now()
	host.CoreOSVersion = payload.CoreOSVersion

	err = host.Commit("updated state to running")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("committing updated host state=running failed"))
		return
	}
	mgr.cluster.Update()
	w.WriteHeader(202)
}

func (mgr *pxeManagerT) setProviderId(serial string, w http.ResponseWriter, r *http.Request) {
	host, exists := mgr.cluster.HostWithSerial(serial)
	if !exists {
		w.WriteHeader(400)
		w.Write([]byte("host doesn't exist"))
		return
	}

	decoder := json.NewDecoder(r.Body)
	payload := hostmgr.Host{}
	err := decoder.Decode(&payload)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("unable to parse json data in set_provider_id request"))
		return
	}

	host.ProviderId = payload.ProviderId
	err = host.Commit("updated host provider id")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("committing updated host provider id failed"))
		return
	}
	mgr.cluster.Update()
	w.WriteHeader(202)
}

func (mgr *pxeManagerT) setIPMIAddr(serial string, w http.ResponseWriter, r *http.Request) {
	host, exists := mgr.cluster.HostWithSerial(serial)
	if !exists {
		w.WriteHeader(400)
		w.Write([]byte("host doesn't exist"))
		return
	}

	decoder := json.NewDecoder(r.Body)
	payload := hostmgr.Host{}
	err := decoder.Decode(&payload)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("unable to parse json data in set_ipmi_addr request"))
		return
	}

	host.IPMIAddr = payload.IPMIAddr
	err = host.Commit("updated host ipmi address")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("committing updated host ipmi address failed"))
		return
	}
	mgr.cluster.Update()
	w.WriteHeader(202)
}

func (mgr *pxeManagerT) setState(serial string, w http.ResponseWriter, r *http.Request) {
	host, exists := mgr.cluster.HostWithSerial(serial)
	if !exists {
		w.WriteHeader(400)
		w.Write([]byte("host doesn't exist"))
		return
	}

	decoder := json.NewDecoder(r.Body)
	payload := hostmgr.Host{}
	err := decoder.Decode(&payload)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("unable to parse json data in set_state request"))
		return
	}

	host.State = payload.State
	switch payload.State {
	case hostmgr.Configured:
		host.State = hostmgr.Configured
	case hostmgr.Running:
		host.State = hostmgr.Configured
	}

	err = host.Commit("updated host state")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("committing updated host state failed"))
		return
	}

	mgr.cluster.Update()

	err = mgr.updateDNSmasqs()
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("updated host state failed in update DNSmasq"))
		return
	}

	mgr.cluster.Update()
	w.WriteHeader(202)
}

func (mgr *pxeManagerT) override(serial string, w http.ResponseWriter, r *http.Request) {
	host, exists := mgr.cluster.HostWithSerial(serial)
	if !exists {
		w.WriteHeader(400)
		w.Write([]byte("host doesn't exist"))
		return
	}

	decoder := json.NewDecoder(r.Body)
	payload := hostmgr.Host{}
	err := decoder.Decode(&payload)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("unable to parse json data in override request"))
		return
	}

	if len(payload.Overrides) == 0 {
		w.WriteHeader(400)
		w.Write([]byte("nothing to override"))
		return
	}

	updatedVars := []string{}
	if host.Overrides == nil {
		host.Overrides = make(map[string]interface{})
	}
	for k, v := range payload.Overrides {
		host.Overrides[k] = v
		updatedVars = append(updatedVars, k)
	}

	err = host.Commit("updated host overrides: " + strings.Join(updatedVars, ", "))
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("committing updated host overrides failed"))
		return
	}
	mgr.cluster.Update()
	w.WriteHeader(202)
}

func (mgr *pxeManagerT) setEtcdClusterToken(serial string, w http.ResponseWriter, r *http.Request) {
	host, exists := mgr.cluster.HostWithSerial(serial)
	if !exists {
		w.WriteHeader(400)
		w.Write([]byte("host doesn't exist"))
		return
	}

	decoder := json.NewDecoder(r.Body)
	payload := hostmgr.Host{}
	err := decoder.Decode(&payload)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("unable to parse json data in set_etcd_cluster_token request"))
		return
	}

	host.EtcdClusterToken = payload.EtcdClusterToken
	err = host.Commit("updated host etcd cluster token")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("committing updated host etcd cluster token failed"))
		return
	}
	mgr.cluster.Update()
	w.WriteHeader(202)
}

func (mgr *pxeManagerT) welcomeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("this is the iPXE server of mayu " + mgr.version))
	return
}

func (mgr *pxeManagerT) hostCoreOSVersion(r *http.Request) string {
	coreOSversion := mgr.config.DefaultCoreOSVersion

	host, exists := mgr.hostFromSerialVar(r)
	if exists {
		if version, exist := host.Overrides["CoreOSVersion"]; exist {
			return version.(string)
		}

		if host.CoreOSVersion == "" {
			return mgr.config.DefaultCoreOSVersion
		} else {
			return host.CoreOSVersion
		}
	}

	return coreOSversion
}
