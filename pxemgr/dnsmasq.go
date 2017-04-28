package pxemgr

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"text/template"

	"github.com/golang/glog"
)

type DNSmasqConfiguration struct {
	Executable string
	Template   string
	TFTPRoot   string
	PXEPort    int
}

type DNSmasqInstance struct {
	confpath string
	args     []string

	conf DNSmasqConfiguration
	cmd  *exec.Cmd
}

func NewDNSmasq(baseFile string, conf DNSmasqConfiguration) *DNSmasqInstance {
	confFile := baseFile + ".conf"
	leaseFile := baseFile + ".lease"

	return &DNSmasqInstance{
		args:     []string{"-k", "-d", "--conf-file=" + confFile, "--dhcp-leasefile=" + leaseFile},
		confpath: confFile,
		conf:     conf,
	}
}

func (dnsmasq *DNSmasqInstance) Start() error {
	glog.V(8).Infoln("starting Dnsmasq server")
	cmd := exec.Command(dnsmasq.conf.Executable, dnsmasq.args...)

	if glog.V(8) {
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return err
		}

		pipeLogger := func(rdr io.Reader) {
			scanner := bufio.NewScanner(rdr)
			for scanner.Scan() {
				glog.V(8).Infoln(scanner.Text())
			}
		}

		go pipeLogger(stdout)
		go pipeLogger(stderr)
	}

	cmd.SysProcAttr = genPlatformSysProcAttr()
	dnsmasq.cmd = cmd
	err := cmd.Start()
	if err != nil {
		glog.Errorln(err)
		return err
	}
	go func(cmd *exec.Cmd) {
		err := cmd.Wait()
		if err != nil {
			glog.Errorln(err)
		}
	}(cmd)
	return nil
}

func (dnsmasq *DNSmasqInstance) Restart() error {
	glog.V(8).Infoln("restarting Dnsmasq server")

	if dnsmasq.cmd != nil {
		dnsmasq.cmd.Process.Kill()
	}
	return dnsmasq.Start()
}

func (dnsmasq *DNSmasqInstance) updateConf(net Network) error {
	glog.V(8).Infoln("updating Dnsmasq configuration")

	tmpl, err := template.ParseFiles(dnsmasq.conf.Template)
	if err != nil {
		return err
	}

	tmplArgs := struct {
		Network Network
		Global  DNSmasqConfiguration
	}{
		Network: net,
		Global:  dnsmasq.conf,
	}

	file, err := os.Create(dnsmasq.confpath)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, tmplArgs)
}
