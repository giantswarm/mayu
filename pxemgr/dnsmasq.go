package pxemgr

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"text/template"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
)

type DNSmasqConfiguration struct {
	Executable string
	Template   string
	TFTPRoot   string
	PXEPort    int

	Logger micrologger.Logger
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
	dnsmasq.conf.Logger.Log("level", "info", "component", "dnsmasq", "message", "starting Dnsmasq server")
	cmd := exec.Command(dnsmasq.conf.Executable, dnsmasq.args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return microerror.Mask(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return microerror.Mask(err)
	}

	pipeLogger := func(rdr io.Reader) {
		scanner := bufio.NewScanner(rdr)
		for scanner.Scan() {
			dnsmasq.conf.Logger.Log("level", "info", "component", "dnsmasq", "message", scanner.Text())
		}
	}
	go pipeLogger(stdout)
	go pipeLogger(stderr)

	cmd.SysProcAttr = genPlatformSysProcAttr()
	dnsmasq.cmd = cmd
	err = cmd.Start()
	if err != nil {
		dnsmasq.conf.Logger.Log("level", "error", "component", "dnsmasq", "message", "failed to start dns command", "stack", err)
		return microerror.Mask(err)
	}
	go func(cmd *exec.Cmd) {
		err := cmd.Wait()
		if err != nil {
			dnsmasq.conf.Logger.Log("level", "error", "component", "dnsmasq", "message", "failed to start dns command", "stack", err)
		}
	}(cmd)

	return nil
}

func (dnsmasq *DNSmasqInstance) Restart() error {
	dnsmasq.conf.Logger.Log("level", "info", "component", "dnsmasq", "message", "restarting Dnsmasq server")

	if dnsmasq.cmd != nil {
		dnsmasq.cmd.Process.Kill()
	}
	err := dnsmasq.Start()
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func (dnsmasq *DNSmasqInstance) updateConf(net Network) error {
	dnsmasq.conf.Logger.Log("level", "info", "component", "dnsmasq", "message", "updating Dnsmasq configuration")

	tmpl, err := template.ParseFiles(dnsmasq.conf.Template)
	if err != nil {
		return microerror.Mask(err)
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
		return microerror.Mask(err)
	}
	defer file.Close()

	err = tmpl.Execute(file, tmplArgs)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}
