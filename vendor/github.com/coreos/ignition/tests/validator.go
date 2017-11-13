// Copyright 2017 CoreOS, Inc.
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

package blackbox

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"github.com/coreos/ignition/internal/exec/util"
	"github.com/coreos/ignition/tests/types"
)

func regexpSearch(t *testing.T, itemName, pattern string, data []byte) string {
	re := regexp.MustCompile(pattern)
	match := re.FindSubmatch(data)
	if len(match) < 2 {
		t.Fatalf("couldn't find %s", itemName)
	}
	return string(match[1])
}

func validatePartitions(t *testing.T, expected []*types.Partition, imageFile string) {
	for _, e := range expected {
		if e.TypeCode == "blank" || e.FilesystemType == "swap" {
			continue
		}
		sgdiskInfo, err := exec.Command(
			"sgdisk", "-i", strconv.Itoa(e.Number),
			imageFile).CombinedOutput()
		if err != nil {
			fmt.Printf("sgdisk -i %d %s died\n", e.Number, imageFile)
			bufio.NewReader(os.Stdin).ReadBytes('\n')
			t.Error("sgdisk -i", strconv.Itoa(e.Number), err)
			return
		}

		actualGUID := regexpSearch(t, "GUID", "Partition unique GUID: (?P<partition_guid>[\\d\\w-]+)", sgdiskInfo)
		actualTypeGUID := regexpSearch(t, "type GUID", "Partition GUID code: (?P<partition_code>[\\d\\w-]+)", sgdiskInfo)
		actualSectors := regexpSearch(t, "partition size", "Partition size: (?P<sectors>\\d+) sectors", sgdiskInfo)
		actualLabel := regexpSearch(t, "partition name", "Partition name: '(?P<name>[\\d\\w-_]+)'", sgdiskInfo)

		// have to align the size to the nearest sector first
		expectedSectors := align(e.Length, 512)

		if e.TypeGUID != "" && e.TypeGUID != actualTypeGUID {
			t.Error("TypeGUID does not match!", e.TypeGUID, actualTypeGUID)
		}
		if e.GUID != "" && formatUUID(e.GUID) != formatUUID(actualGUID) {
			t.Error("GUID does not match!", e.GUID, actualGUID)
		}
		if e.Label != actualLabel {
			t.Error("Label does not match!", e.Label, actualLabel)
		}
		if strconv.Itoa(expectedSectors) != actualSectors {
			t.Error(
				"Sectors does not match!", expectedSectors, actualSectors)
		}
	}
}

func formatUUID(s string) string {
	return strings.ToUpper(strings.Replace(s, "-", "", -1))
}

func validateFilesystems(t *testing.T, expected []*types.Partition, imageFile string) {
	for _, e := range expected {
		if e.FilesystemType != "" {
			filesystemType, err := util.FilesystemType(e.Device)
			if err != nil {
				t.Fatalf("couldn't determine filesystem type: %v", err)
			}
			if filesystemType != e.FilesystemType {
				t.Errorf("FilesystemType does not match, expected:%q actual:%q",
					e.FilesystemType, filesystemType)
			}
		}
		if e.FilesystemUUID != "" {
			filesystemUUID, err := util.FilesystemUUID(e.Device)
			if err != nil {
				t.Fatalf("couldn't determine filesystem uuid: %v", err)
			}
			if formatUUID(filesystemUUID) != formatUUID(e.FilesystemUUID) {
				t.Errorf("FilesystemUUID does not match, expected:%q actual:%q",
					e.FilesystemUUID, filesystemUUID)
			}
		}
		if e.FilesystemLabel != "" {
			filesystemLabel, err := util.FilesystemLabel(e.Device)
			if err != nil {
				t.Fatalf("couldn't determine filesystem label: %v", err)
			}
			if filesystemLabel != e.FilesystemLabel {
				t.Errorf("FilesystemLabel does not match, expected:%q actual:%q",
					e.FilesystemLabel, filesystemLabel)
			}
		}
	}
}

func validateFilesDirectoriesAndLinks(t *testing.T, expected []*types.Partition) {
	for _, partition := range expected {
		for _, file := range partition.Files {
			validateFile(t, partition, file)
		}
		for _, dir := range partition.Directories {
			validateDirectory(t, partition, dir)
		}
		for _, link := range partition.Links {
			validateLink(t, partition, link)
		}
		for _, node := range partition.RemovedNodes {
			path := filepath.Join(partition.MountPath, node.Directory, node.Name)
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Error("Node was expected to be removed and is present!", path)
			}
		}
	}
}

func validateFile(t *testing.T, partition *types.Partition, file types.File) {
	path := filepath.Join(partition.MountPath, file.Node.Directory, file.Node.Name)
	fileInfo, err := os.Stat(path)
	if err != nil {
		t.Errorf("Error stat'ing file %s: %v", path, err)
		return
	}
	if file.Contents != "" {
		dat, err := ioutil.ReadFile(path)
		if err != nil {
			t.Error("Error when reading file", path)
			return
		}

		actualContents := string(dat)
		if file.Contents != actualContents {
			t.Error("Contents of file", path, "do not match!",
				file.Contents, actualContents)
		}
	}

	validateNode(t, fileInfo, file.Node)
}

func validateDirectory(t *testing.T, partition *types.Partition, dir types.Directory) {
	path := filepath.Join(partition.MountPath, dir.Node.Directory, dir.Node.Name)
	dirInfo, err := os.Stat(path)
	if err != nil {
		t.Errorf("Error stat'ing directory %s: %v", path, err)
		return
	}
	if !dirInfo.IsDir() {
		t.Errorf("Node at %s is not a directory!", path)
	}
	validateMode(t, path, dir.Mode)
	validateNode(t, dirInfo, dir.Node)
}

func validateLink(t *testing.T, partition *types.Partition, link types.Link) {
	linkPath := filepath.Join(partition.MountPath, link.Node.Directory, link.Node.Name)
	linkInfo, err := os.Lstat(linkPath)
	if err != nil {
		t.Error("Error stat'ing link \"" + linkPath + "\": " + err.Error())
		return
	}
	if link.Hard {
		targetPath := filepath.Join(partition.MountPath, link.Target)
		targetInfo, err := os.Stat(targetPath)
		if err != nil {
			t.Error("Error stat'ing target \"" + targetPath + "\": " + err.Error())
			return
		}
		if linkInfoSys, ok := linkInfo.Sys().(*syscall.Stat_t); ok {
			if targetInfoSys, ok := targetInfo.Sys().(*syscall.Stat_t); ok {
				if linkInfoSys.Ino != targetInfoSys.Ino {
					t.Error("Hard link and target don't have same inode value: " + linkPath + " " + targetPath)
					return
				}
			} else {
				t.Error("Stat type assertion failed, this will only work on Linux")
				return
			}
		} else {
			t.Error("Stat type assertion failed, this will only work on Linux")
			return
		}
	} else {
		if linkInfo.Mode()&os.ModeType != os.ModeSymlink {
			t.Errorf("Node at symlink path is not a symlink (it's a %s): %s", linkInfo.Mode().String(), linkPath)
			return
		}
		targetPath, err := os.Readlink(linkPath)
		if err != nil {
			t.Error("Error reading symbolic link: " + err.Error())
			return
		}
		if targetPath != link.Target {
			t.Error("Actual and expected symbolic link targets don't match")
			return
		}
	}
	validateNode(t, linkInfo, link.Node)
}

func validateMode(t *testing.T, path string, mode string) {
	if mode != "" {
		fileInfo, err := os.Stat(path)
		if err != nil {
			t.Error("Error running stat on node", path, err)
			return
		}

		// converts to a string containing the base-10 mode
		actualMode := strconv.Itoa(int(fileInfo.Mode()) & 07777)

		if mode != actualMode {
			t.Error("Node Mode does not match", path, mode, actualMode)
		}
	}
}

func validateNode(t *testing.T, nodeInfo os.FileInfo, node types.Node) {
	if nodeInfoSys, ok := nodeInfo.Sys().(*syscall.Stat_t); ok {
		if nodeInfoSys.Uid != uint32(node.User) {
			t.Error("Node has the wrong owner", node.User, nodeInfoSys.Uid)
		}

		if nodeInfoSys.Gid != uint32(node.Group) {
			t.Error("Node has the wrong group owner", node.Group, nodeInfoSys.Gid)
		}
	} else {
		t.Error("Stat type assertion failed, this will only work on Linux")
		return
	}
}
