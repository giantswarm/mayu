package pxemgr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/giantswarm/mayu/hostmgr"
	"github.com/giantswarm/mayu/infopusher/machinedata"
	"github.com/golang/glog"
	"github.com/gorilla/mux"
)

const (
	vmlinuzFile      = "coreos_production_pxe.vmlinuz"
	initrdFile       = "coreos_production_pxe_image.cpio.gz"
	installImageFile = "coreos_production_image.bin.bz2"
)

func (mgr *pxeManagerT) ipxeBootScript(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)

	buffer := bytes.NewBufferString("")
	buffer.WriteString("#!ipxe\n")
	buffer.WriteString(fmt.Sprintf("kernel %s/images/vmlinuz coreos.autologin maybe-install-coreos=stable console=ttyS0,115200n8 mayu=%s next-script=%s\n", mgr.thisHost(), mgr.thisHost(), mgr.thisHost()+"/first-stage-script/__SERIAL__"))
	buffer.WriteString(fmt.Sprintf("initrd %s/images/initrd.cpio.gz\n", mgr.thisHost()))
	buffer.WriteString("boot\n")

	w.Write(buffer.Bytes())
}

func (mgr *pxeManagerT) firstStageScriptGenerator(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	serial := strings.ToLower(params["serial"])
	glog.V(2).Infof("generating a first stage script for '%s'\n", serial)

	infoHelperURL := mgr.thisHost() + "/hostinfo-helper"
	cloudConfigURL := mgr.thisHost() + "/final-cloud-config.yaml"
	setInstalledURL := mgr.thisHost() + "/admin/host/__SERIAL__/set_installed"
	installImageURL := mgr.thisHost() + "/images/install_image.bin.bz2"
	host := mgr.maybeCreateHost(serial)

	ctx := struct {
		HostInfoHelperURL string
		CloudConfigURL    string
		InstallImageURL   string
		SetInstalledURL   string
		MayuURL           string
		MayuVersion       string
		MachineID         string
	}{
		HostInfoHelperURL: infoHelperURL,
		CloudConfigURL:    cloudConfigURL,
		InstallImageURL:   installImageURL,
		SetInstalledURL:   setInstalledURL,
		MayuURL:           mgr.thisHost(),
		MayuVersion:       mgr.version,
		MachineID:         host.MachineID,
	}

	tmpl, err := template.ParseFiles(mgr.firstStageScript)
	if err != nil {
		glog.Fatalln(err)
	}
	tmpl.Execute(w, ctx)
}

func (mgr *pxeManagerT) maybeCreateHost(serial string) *hostmgr.Host {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	host, exists := mgr.cluster.HostWithSerial(serial)
	if !exists {
		var err error
		host, err = mgr.cluster.CreateNewHost(serial)
		if err != nil {
			glog.Fatalln(err)
		}

		if host.InternalAddr == nil {
			host.InternalAddr = mgr.getNextInternalIP()
			err = host.Commit("updated host InternalAddr")
			if err != nil {
				glog.Fatalln(err)
			}
		}

		if host.Profile == "" {
			host.Profile = mgr.getNextProfile()
			if host.Profile != "" {
				host.FleetMetadata = mgr.profileMetadata(host.Profile)
			} else {
				host.FleetMetadata = mgr.profileMetadata("default")
			}
			err = host.Commit("updated host profile and metadata")
			if err != nil {
				glog.Fatalln(err)
			}
		}
	}
	return host
}

func (mgr *pxeManagerT) profileMetadata(profileName string) []string {
	for _, v := range mgr.config.Profiles {
		if v.Name == profileName {
			return v.Tags
		}
	}
	return []string{}
}

func (mgr *pxeManagerT) cloudConfigGenerator(w http.ResponseWriter, r *http.Request) {
	glog.V(2).Infoln("generating a final stage cloudConfig")

	hostData := &machinedata.HostData{}

	dec := json.NewDecoder(r.Body)
	err := dec.Decode(hostData)
	if err != nil {
		glog.Warningln(err)
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	if hostData.Serial == "" {
		glog.Warningf("empty serial. %+v\n", hostData)
		w.WriteHeader(400)
		w.Write([]byte("no serial ? :/"))
		return
	}

	host := mgr.maybeCreateHost(hostData.Serial)
	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	macAddresses := make([]string, len(hostData.NetDevs))
	for i, dev := range hostData.NetDevs {
		macAddresses[i] = dev.MacAddress
	}
	host.MacAddresses = macAddresses

	err = host.Commit("updated host macAddress")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("committing updated host macAddress failed"))
		return
	}

	if hostData.ConnectedNIC != "" && host.ConnectedNIC != hostData.ConnectedNIC {
		host.ConnectedNIC = hostData.ConnectedNIC
		err = host.Commit("updated host connected nic")
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("committing updated host connected nic failed"))
			return
		}
	}

	if hostData.IPMIAddress != nil {
		host.IPMIAddr = hostData.IPMIAddress
		err = host.Commit("updated host ipmi address")
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("committing updated host ipmi address failed"))
			return
		}
	}

	glog.V(2).Infof("got host %+v\n", host)

	host.State = hostmgr.Installing
	err = host.Commit("updated host state to installing")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("committing updated host state=installing failed"))
		return
	}
	mgr.cluster.Update()

	mgr.WriteLastStageCC(*host, w)
}

func (mgr *pxeManagerT) imagesHandler(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/vmlinuz") {
		vmlinuz, err := mgr.getKernelImage()
		if err != nil {
			panic(err)
		}
		setContentLength(w, vmlinuz)
		defer vmlinuz.Close()
		io.Copy(w, vmlinuz)
	} else if strings.HasSuffix(r.URL.Path, "/initrd.cpio.gz") {
		initrd, err := mgr.getInitRD()
		if err != nil {
			panic(err)
		}
		setContentLength(w, initrd)
		defer initrd.Close()
		io.Copy(w, initrd)
	} else if strings.HasSuffix(r.URL.Path, "/install_image.bin.bz2") {
		img, err := mgr.getInstallImage()
		if err != nil {
			panic(err)
		}
		setContentLength(w, img)
		defer img.Close()
		io.Copy(w, img)
	} else {
		panic(fmt.Sprintf("no handler provided for invalid URL path '%s'", r.URL.Path))
	}
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("got request", r.URL)
}

func (mgr *pxeManagerT) getInstallImage() (*os.File, error) {
	return os.Open(path.Join(mgr.imagesCacheDir, installImageFile))
}

func (mgr *pxeManagerT) getKernelImage() (*os.File, error) {
	return os.Open(path.Join(mgr.imagesCacheDir, vmlinuzFile))
}

func (mgr *pxeManagerT) getInitRD() (*os.File, error) {
	return os.Open(path.Join(mgr.imagesCacheDir, initrdFile))
}

func setContentLength(w http.ResponseWriter, f *os.File) error {
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	w.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))
	return nil
}

func (mgr *pxeManagerT) markReconfigure(serial string, w http.ResponseWriter, r *http.Request) {
	host, exists := mgr.cluster.HostWithSerial(serial)
	if !exists {
		w.WriteHeader(400)
		w.Write([]byte("host doesn't exist"))
		return
	}

	host.State = hostmgr.Configured
	host.KeepDiskData = true
	host.Commit("host flagged to be reconfigured")
	mgr.cluster.Update()

	w.WriteHeader(202)
}

func (mgr *pxeManagerT) markFresh(serial string, w http.ResponseWriter, r *http.Request) {
	host, exists := mgr.cluster.HostWithSerial(serial)
	if !exists {
		w.WriteHeader(400)
		w.Write([]byte("host doesn't exist"))
		return
	}

	host.State = hostmgr.Configured
	host.KeepDiskData = false
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

func (mgr *pxeManagerT) infoPusher(w http.ResponseWriter, r *http.Request) {
	helper, err := os.Open(path.Join(mgr.staticHTMLPath, "infopusher"))
	if err != nil {
		glog.Warningln(err)
	}
	setContentLength(w, helper)
	defer helper.Close()
	io.Copy(w, helper)
}

func (mgr *pxeManagerT) hostFromSerialVar(r *http.Request) (*hostmgr.Host, bool) {
	params := mux.Vars(r)
	serial := strings.ToLower(params["serial"])

	return mgr.cluster.HostWithSerial(serial)
}

func (mgr *pxeManagerT) setInstalled(serial string, w http.ResponseWriter, r *http.Request) {
	host, exists := mgr.cluster.HostWithSerial(serial)
	glog.V(3).Infof("marking host '%s' as installed", serial)
	if !exists {
		w.WriteHeader(400)
		w.Write([]byte("host doesn't exist"))
		return
	}

	glog.V(1).Infof("host '%s' just finished installing\n", host.Serial)

	host.State = hostmgr.Installed
	err := host.Commit("updated state to installed")
	mgr.cluster.Update()
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("committing updated host state=installed failed"))
		return
	}
	w.WriteHeader(202)
	mgr.updateDNSmasqs()
}

func (mgr *pxeManagerT) bootComplete(serial string, w http.ResponseWriter, r *http.Request) {
	host, exists := mgr.cluster.HostWithSerial(serial)
	if !exists {
		w.WriteHeader(400)
		w.Write([]byte("host doesn't exist"))
		return
	}

	glog.V(1).Infof("host '%s' just finished booting\n", serial)

	decoder := json.NewDecoder(r.Body)
	payload := hostmgr.Host{}
	err := decoder.Decode(&payload)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte("unable to parse json data in request"))
		return
	}

	host.State = hostmgr.Running
	host.LastBoot = time.Now()
	host.CoreOSVersion = payload.CoreOSVersion
	host.MayuVersion = payload.MayuVersion
	host.EtcdVersion = payload.EtcdVersion
	host.FleetVersion = payload.FleetVersion
	host.DockerVersion = payload.DockerVersion
	host.RktVersion = payload.RktVersion
	host.K8sVersion = payload.K8sVersion
	host.YochuVersion = payload.YochuVersion
	glog.V(1).Infof("yochu version '%s'\n", payload.YochuVersion)

	err = host.Commit("updated state to running")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("committing updated host state=running failed"))
		return
	}
	mgr.cluster.Update()
	w.WriteHeader(202)
}

func (mgr *pxeManagerT) setMetadata(serial string, w http.ResponseWriter, r *http.Request) {
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
		w.Write([]byte("unable to parse json data in request"))
		return
	}

	host.FleetMetadata = payload.FleetMetadata
	err = host.Commit("updated host metadata")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("committing updated host metadata failed"))
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
		w.Write([]byte("unable to parse json data in request"))
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
		w.Write([]byte("unable to parse json data in request"))
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

func (mgr *pxeManagerT) setCabinet(serial string, w http.ResponseWriter, r *http.Request) {
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
		w.Write([]byte("unable to parse json data in request"))
		return
	}

	host.Cabinet = payload.Cabinet
	err = host.Commit("updated host cabinet")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("committing updated host cabinet failed"))
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
