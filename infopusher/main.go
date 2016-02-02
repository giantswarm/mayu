package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/giantswarm/mayu/infopusher/machinedata"
	"github.com/golang/glog"
)

var url = flag.String("post-url", "", "url to post the host data")

func main() {
	log.SetFlags(0)
	flag.Set("logtostderr", "true")
	flag.Parse()

	if *url == "" {
		fmt.Println("you need to set the parameter post-url")
		os.Exit(1)
	}

	data, err := json.Marshal(machinedata.HostData{
		Serial:       fetchDMISerial(),
		NetDevs:      fetchNetDevs(),
		ConnectedNIC: fetchConnectedNIC(),
		IPMIAddress:  fetchIPMIAddress(),
	})

	if err != nil {
		glog.Fatalln(err)
	}

	resp, err := http.Post(*url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		glog.Fatalln(err)
	}

	io.Copy(os.Stdout, resp.Body)
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func readContents(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(b)), nil
}

func fetchDMISerial() string {
	productSerial := fetchDMIXSerial("product_serial")
	if len(productSerial) > 0 {
		return productSerial
	}
	return fetchDMIXSerial("chassis_serial")
}

func fetchDMIXSerial(x string) string {
	serialFile := "/sys/devices/virtual/dmi/id/" + x
	if fileExists(serialFile) {
		if contents, err := readContents(serialFile); err == nil {
			return strings.Replace(contents, " ", "-", -1)
		} else {
			glog.Warningln(err)
		}
	}
	return ""
}

func fetchNetDevs() []machinedata.NetDev {
	classNetDir := "/sys/class/net"
	classNet, err := os.Open(classNetDir)
	if err != nil {
		glog.Warningln(err)
		return []machinedata.NetDev{}
	}

	finfos, err := classNet.Readdir(1000)
	if err != nil {
		glog.Warningln(err)
		return []machinedata.NetDev{}
	}

	rv := []machinedata.NetDev{}

	for _, fi := range finfos {
		fullPath := path.Join(classNetDir, fi.Name())
		link, err := os.Readlink(fullPath)
		if err != nil {
			continue
		}
		if !strings.Contains(link, "virtual") {
			macAddr, _ := readContents(path.Join(fullPath, "address"))
			rv = append(rv, machinedata.NetDev{
				Name:       fi.Name(),
				MacAddress: macAddr,
			})
		}
	}

	return rv
}

func fetchConnectedNIC() string {
	c, err := readContents("/proc/net/route")
	if err != nil {
		return ""
	}
	strToInt := func(s string) int {
		i, err := strconv.Atoi(s)
		if err != nil {
			return -1
		}
		return i
	}
	lines := strings.Split(c, "\n")
	for _, line := range lines {
		columns := strings.Fields(line)
		if len(columns) > 8 {
			destination := strToInt(columns[1])
			flags := strToInt(columns[3])
			mask := strToInt(columns[7])
			nic := columns[0]

			// flags:
			//   RTF_UP   0x0001        /* route usable             */
			//   RTF_GATEWAY  0x0002    /* destination is a gateway */

			if destination == 0 && mask == 0 && (flags&1 == 1) && (flags&2 == 2) {
				return nic
			}
		}
	}
	return ""
}

func fetchIPMIAddress() (ip net.IP) {
	ipmitool, err := Asset("embedded/ipmitool")
	if err != nil {
		glog.Warningln(err)
		return
	}

	tempFile, err := ioutil.TempFile("", "ipmitool")
	if err != nil {
		glog.Warningln(err)
		return
	}

	tempFile.Write(ipmitool)
	tempFile.Chmod(0700)
	tempFile.Close()

	ipmitoolCmd := exec.Command(tempFile.Name(), "lan", "print")
	output, err := ipmitoolCmd.CombinedOutput()
	if err != nil {
		glog.Warningln(err)
		return
	}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "IP Address    ") {
			fields := strings.Fields(line)
			return net.ParseIP(fields[len(fields)-1])
		}
	}

	return
}
