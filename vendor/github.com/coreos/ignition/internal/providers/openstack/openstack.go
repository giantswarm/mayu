// Copyright 2016 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// The OpenStack provider fetches configurations from the userdata available in
// both the config-drive as well as the network metadata service. Whichever
// responds first is the config that is used.
// NOTE: This provider is still EXPERIMENTAL.

package openstack

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/coreos/ignition/config"
	"github.com/coreos/ignition/config/types"
	"github.com/coreos/ignition/config/validate/report"
	"github.com/coreos/ignition/internal/log"
	"github.com/coreos/ignition/internal/resource"

	"golang.org/x/net/context"
)

const (
	configDrivePath         = "/dev/disk/by-label/config-2"
	configDriveUserdataPath = "/openstack/latest/user_data"
)

var (
	metadataServiceUrl = url.URL{
		Scheme: "http",
		Host:   "169.254.169.254",
		Path:   "openstack/latest/user_data",
	}
)

func FetchConfig(logger *log.Logger, client *resource.HttpClient) (types.Config, report.Report, error) {
	var data []byte
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	dispatch := func(name string, fn func() ([]byte, error)) {
		raw, err := fetchConfigFromConfigDrive(logger, ctx)
		if err != nil {
			switch err {
			case context.Canceled:
			case context.DeadlineExceeded:
				logger.Err("timed out while fetching config from %s", name)
			default:
				logger.Err("failed to fetch config from %s: %v", name, err)
			}
			return
		}

		data = raw
		cancel()
	}

	go dispatch("config drive", func() ([]byte, error) {
		return fetchConfigFromConfigDrive(logger, ctx)
	})
	go dispatch("metadata service", func() ([]byte, error) {
		return fetchConfigFromMetadataService(logger, client, ctx)
	})

	<-ctx.Done()
	if ctx.Err() == context.DeadlineExceeded {
		logger.Info("neither config drive nor metadata service were available in time. Continuing without a config...")
	}

	return config.Parse(data)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return (err == nil)
}

func fetchConfigFromConfigDrive(logger *log.Logger, ctx context.Context) ([]byte, error) {
	for !fileExists(configDrivePath) {
		logger.Debug("config drive (%q) not found. Waiting...", configDrivePath)
		select {
		case <-time.After(time.Second):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	logger.Debug("creating temporary mount point")
	mnt, err := ioutil.TempDir("", "ignition-configdrive")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.Remove(mnt)

	cmd := exec.Command("/usr/bin/mount", "-o", "ro", "-t", "auto", configDrivePath, mnt)
	if err := logger.LogCmd(cmd, "mounting config drive"); err != nil {
		return nil, err
	}
	defer logger.LogOp(
		func() error { return syscall.Unmount(mnt, 0) },
		"unmounting %q at %q", configDrivePath, mnt,
	)

	if !fileExists(filepath.Join(mnt, configDriveUserdataPath)) {
		return nil, nil
	}

	return ioutil.ReadFile(filepath.Join(mnt, configDriveUserdataPath))
}

func fetchConfigFromMetadataService(logger *log.Logger, client *resource.HttpClient, ctx context.Context) ([]byte, error) {
	return resource.FetchConfig(logger, client, context.Background(), metadataServiceUrl), nil
}
