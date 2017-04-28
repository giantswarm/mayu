// The client package is a client implementation of the mayu network API.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"

	mayuerror "github.com/giantswarm/mayu/error"
	"github.com/giantswarm/mayu/hostmgr"
	"github.com/giantswarm/mayu/httputil"
	"github.com/giantswarm/mayu/pxemgr"
	"gopkg.in/yaml.v2"
)

const contentType = "application/json"

// Client implements the network API. Check the corresponding methods.
type Client struct {
	// Scheme defines the protocol scheme. This is either http or https.
	Scheme string

	// Host is used to connect to mayu over network.
	Host string

	// Port is used to connect to mayu over network.
	Port uint16
}

// New creates a new configured client to interact with mayu over its network
// API.
func New(scheme, host string, port uint16) (*Client, error) {
	client := &Client{
		Scheme: scheme,
		Host:   host,
		Port:   port,
	}

	return client, nil
}

func (c *Client) BootComplete(serial string, host hostmgr.Host) error {
	data, err := json.Marshal(host)

	if err != nil {
		return mayuerror.MaskAny(err)
	}

	resp, err := httputil.Put(fmt.Sprintf("%s://%s:%d/admin/host/%s/boot_complete", c.Scheme, c.Host, c.Port, serial), "application/json", bytes.NewBuffer(data))
	if err != nil {
		return mayuerror.MaskAny(err)
	}
	defer resp.Body.Close()
	return nil
}

// SetMetadata sets fleet metadata given by value for a node given by serial.
func (c *Client) SetMetadata(serial, value string) error {
	data, err := json.Marshal(hostmgr.Host{
		FleetMetadata: strings.Split(value, ","),
	})
	if err != nil {
		return mayuerror.MaskAny(err)
	}

	resp, err := httputil.Put(fmt.Sprintf("%s://%s:%d/admin/host/%s/set_metadata", c.Scheme, c.Host, c.Port, serial), contentType, bytes.NewBuffer(data))
	if err != nil {
		return mayuerror.MaskAny(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode > 399 {
		return mayuerror.MaskAny(fmt.Errorf("invalid status code '%d'", resp.StatusCode))
	}

	return nil
}

// SetProviderId sets the provider ID given by value for a node given by serial.
func (c *Client) SetProviderId(serial, value string) error {
	data, err := json.Marshal(hostmgr.Host{
		ProviderId: value,
	})
	if err != nil {
		return mayuerror.MaskAny(err)
	}

	resp, err := httputil.Put(fmt.Sprintf("%s://%s:%d/admin/host/%s/set_provider_id", c.Scheme, c.Host, c.Port, serial), contentType, bytes.NewBuffer(data))
	if err != nil {
		return mayuerror.MaskAny(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode > 399 {
		return mayuerror.MaskAny(fmt.Errorf("invalid status code '%d'", resp.StatusCode))
	}

	return nil
}

// SetIPMIAddr sets the IPMI address given by value for a node given by serial.
func (c *Client) SetIPMIAddr(serial, value string) error {
	data, err := json.Marshal(hostmgr.Host{
		IPMIAddr: net.ParseIP(value),
	})
	if err != nil {
		return mayuerror.MaskAny(err)
	}

	resp, err := httputil.Put(fmt.Sprintf("%s://%s:%d/admin/host/%s/set_ipmi_addr", c.Scheme, c.Host, c.Port, serial), contentType, bytes.NewBuffer(data))
	if err != nil {
		return mayuerror.MaskAny(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode > 399 {
		return mayuerror.MaskAny(fmt.Errorf("invalid status code '%d'", resp.StatusCode))
	}

	return nil
}

// SetEtcdClusterToken sets the etcd cluster token given by value for a node given by serial.
func (c *Client) SetEtcdClusterToken(serial, value string) error {
	data, err := json.Marshal(hostmgr.Host{
		EtcdClusterToken: value,
	})
	if err != nil {
		return mayuerror.MaskAny(err)
	}

	resp, err := httputil.Put(fmt.Sprintf("%s://%s:%d/admin/host/%s/set_etcd_cluster_token", c.Scheme, c.Host, c.Port, serial), contentType, bytes.NewBuffer(data))
	if err != nil {
		return mayuerror.MaskAny(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode > 399 {
		return mayuerror.MaskAny(fmt.Errorf("invalid status code '%d'", resp.StatusCode))
	}

	return nil
}

// SetState sets the machine state for a node given by serial.
func (c *Client) SetState(serial, value string) error {
	state, err := hostmgr.HostState(value)
	if err != nil {
		return mayuerror.MaskAny(err)
	}

	data, err := json.Marshal(hostmgr.Host{
		State: state,
	})
	if err != nil {
		return mayuerror.MaskAny(err)
	}

	resp, err := httputil.Put(fmt.Sprintf("%s://%s:%d/admin/host/%s/set_state", c.Scheme, c.Host, c.Port, serial), contentType, bytes.NewBuffer(data))
	if err != nil {
		return mayuerror.MaskAny(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode > 399 {
		return mayuerror.MaskAny(fmt.Errorf("invalid status code '%d'", resp.StatusCode))
	}

	return nil
}

// SetCabinet sets the cabinet given by value for a node given by serial.
func (c *Client) SetCabinet(serial, value string) error {
	cabinet, err := strconv.ParseUint(value, 10, 0)
	if err != nil {
		return mayuerror.MaskAny(err)
	}

	data, err := json.Marshal(hostmgr.Host{
		Cabinet: uint(cabinet),
	})
	if err != nil {
		return mayuerror.MaskAny(err)
	}

	resp, err := httputil.Put(fmt.Sprintf("%s://%s:%d/admin/host/%s/set_cabinet", c.Scheme, c.Host, c.Port, serial), contentType, bytes.NewBuffer(data))
	if err != nil {
		return mayuerror.MaskAny(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode > 399 {
		return mayuerror.MaskAny(fmt.Errorf("invalid status code '%d'", resp.StatusCode))
	}

	return nil
}

// Override overrides a template properties such as docker_version, yochu_version, etc
func (c *Client) Override(serial, property, value string) error {
	data, err := json.Marshal(hostmgr.Host{
		Overrides: map[string]interface{}{property: value},
	})
	if err != nil {
		return mayuerror.MaskAny(err)
	}

	resp, err := httputil.Put(fmt.Sprintf("%s://%s:%d/admin/host/%s/override", c.Scheme, c.Host, c.Port, serial), contentType, bytes.NewBuffer(data))
	if err != nil {
		return mayuerror.MaskAny(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode > 399 {
		return mayuerror.MaskAny(fmt.Errorf("invalid status code '%d'", resp.StatusCode))
	}

	return nil
}

// List fetches a list of node information within the current cluster.
func (c *Client) List() ([]hostmgr.Host, error) {
	list := []hostmgr.Host{}

	resp, err := http.Get(fmt.Sprintf("%s://%s:%d/admin/hosts", c.Scheme, c.Host, c.Port))
	if err != nil {
		return list, mayuerror.MaskAny(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode > 399 {
		return nil, mayuerror.MaskAny(fmt.Errorf("invalid status code '%d'", resp.StatusCode))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return list, mayuerror.MaskAny(err)
	}

	err = json.Unmarshal(body, &list)
	if err != nil {
		return list, mayuerror.MaskAny(err)
	}

	return list, nil
}

// Status fetches status information for a node given by serial.
func (c *Client) Status(serial string) (hostmgr.Host, error) {
	var host hostmgr.Host

	resp, err := http.Get(fmt.Sprintf("%s://%s:%d/admin/hosts", c.Scheme, c.Host, c.Port))
	if err != nil {
		return host, mayuerror.MaskAny(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode > 399 {
		return host, mayuerror.MaskAny(fmt.Errorf("invalid status code '%d'", resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return host, mayuerror.MaskAny(err)
	}

	list := []hostmgr.Host{}
	err = json.Unmarshal(body, &list)
	if err != nil {
		return host, mayuerror.MaskAny(err)
	}

	for _, host = range list {
		if host.Serial == serial {
			return host, nil
		}
	}

	return host, fmt.Errorf("host %s not found.", serial)
}

func (c *Client) GetConfig() (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s://%s:%d/admin/mayu_config", c.Scheme, c.Host, c.Port))
	if err != nil {
		return "", mayuerror.MaskAny(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode > 399 {
		return "", mayuerror.MaskAny(fmt.Errorf("invalid status code '%d'", resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", mayuerror.MaskAny(err)
	}

	return string(body), nil
}

func (c *Client) SetConfig(conf pxemgr.Configuration) error {
	data, err := yaml.Marshal(conf)
	if err != nil {
		return mayuerror.MaskAny(err)
	}
	resp, err := httputil.Put(fmt.Sprintf("%s://%s:%d/admin/mayu_config", c.Scheme, c.Host, c.Port), "text/plain", bytes.NewBuffer(data))
	if err != nil {
		return mayuerror.MaskAny(err)
	}
	if resp.StatusCode > 399 {
		return mayuerror.MaskAny(fmt.Errorf("invalid status code '%d'", resp.StatusCode))
	}

	return nil
}
