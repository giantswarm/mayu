// Copyright 2018 CoreOS, Inc.
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

package v2_2

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/coreos/ignition/config/util"
	v2_1 "github.com/coreos/ignition/config/v2_1/types"
	"github.com/coreos/ignition/config/v2_2/types"
)

func TestTranslateFromV2_1(t *testing.T) {
	type in struct {
		config v2_1.Config
	}
	type out struct {
		config types.Config
	}

	tests := []struct {
		in  in
		out out
	}{
		{
			in:  in{},
			out: out{config: types.Config{Ignition: types.Ignition{Version: types.MaxVersion.String()}}},
		},
		{
			in: in{config: v2_1.Config{
				Ignition: v2_1.Ignition{
					Config: v2_1.IgnitionConfig{
						Append: []v2_1.ConfigReference{
							{
								Source: (&url.URL{
									Scheme: "data",
									Opaque: ",file1",
								}).String(),
							},
							{
								Source: (&url.URL{
									Scheme: "data",
									Opaque: ",file2",
								}).String(),
								Verification: v2_1.Verification{
									Hash: util.StrToPtr("func2-sum2"),
								},
							},
						},
						Replace: &v2_1.ConfigReference{
							Source: (&url.URL{
								Scheme: "data",
								Opaque: ",file3",
							}).String(),
							Verification: v2_1.Verification{
								Hash: util.StrToPtr("func3-sum3"),
							},
						},
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{
					Version: types.MaxVersion.String(),
					Config: types.IgnitionConfig{
						Append: []types.ConfigReference{
							{
								Source: (&url.URL{
									Scheme: "data",
									Opaque: ",file1",
								}).String(),
							},
							{
								Source: (&url.URL{
									Scheme: "data",
									Opaque: ",file2",
								}).String(),
								Verification: types.Verification{
									Hash: util.StrToPtr("func2-sum2"),
								},
							},
						},
						Replace: &types.ConfigReference{
							Source: (&url.URL{
								Scheme: "data",
								Opaque: ",file3",
							}).String(),
							Verification: types.Verification{
								Hash: util.StrToPtr("func3-sum3"),
							},
						},
					},
				},
			}},
		},
		{
			in: in{config: v2_1.Config{
				Ignition: v2_1.Ignition{
					Timeouts: v2_1.Timeouts{
						HTTPResponseHeaders: util.IntToPtr(0),
						HTTPTotal:           util.IntToPtr(0),
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{
					Version: types.MaxVersion.String(),
					Timeouts: types.Timeouts{
						HTTPResponseHeaders: util.IntToPtr(0),
						HTTPTotal:           util.IntToPtr(0),
					},
				},
			}},
		},
		{
			in: in{config: v2_1.Config{
				Ignition: v2_1.Ignition{
					Timeouts: v2_1.Timeouts{
						HTTPResponseHeaders: util.IntToPtr(50),
						HTTPTotal:           util.IntToPtr(100),
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{
					Version: types.MaxVersion.String(),
					Timeouts: types.Timeouts{
						HTTPResponseHeaders: util.IntToPtr(50),
						HTTPTotal:           util.IntToPtr(100),
					},
				},
			}},
		},
		{
			in: in{config: v2_1.Config{
				Ignition: v2_1.Ignition{Version: v2_1.MaxVersion.String()},
				Storage: v2_1.Storage{
					Disks: []v2_1.Disk{
						{
							Device:    "/dev/sda",
							WipeTable: true,
							Partitions: []v2_1.Partition{
								{
									Label:    "ROOT",
									Number:   7,
									Size:     100,
									Start:    50,
									TypeGUID: "HI",
									GUID:     "4F68BCE3-E8CD-4DB1-96E7-FBCAF984B709",
								},
								{
									Label:    "DATA",
									Number:   12,
									Size:     1000,
									Start:    300,
									TypeGUID: "LO",
									GUID:     "3B8F8425-20E0-4F3B-907F-1A25A76F98E8",
								},
							},
						},
						{
							Device:    "/dev/sdb",
							WipeTable: true,
						},
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{Version: types.MaxVersion.String()},
				Storage: types.Storage{
					Disks: []types.Disk{
						{
							Device:    "/dev/sda",
							WipeTable: true,
							Partitions: []types.Partition{
								{
									Label:    "ROOT",
									Number:   7,
									Size:     100,
									Start:    50,
									TypeGUID: "HI",
									GUID:     "4F68BCE3-E8CD-4DB1-96E7-FBCAF984B709",
								},
								{
									Label:    "DATA",
									Number:   12,
									Size:     1000,
									Start:    300,
									TypeGUID: "LO",
									GUID:     "3B8F8425-20E0-4F3B-907F-1A25A76F98E8",
								},
							},
						},
						{
							Device:    "/dev/sdb",
							WipeTable: true,
						},
					},
				},
			}},
		},
		{
			in: in{config: v2_1.Config{
				Ignition: v2_1.Ignition{Version: v2_1.MaxVersion.String()},
				Storage: v2_1.Storage{
					Raid: []v2_1.Raid{
						{
							Name:  "fast",
							Level: "raid0",
							Devices: []v2_1.Device{
								v2_1.Device("/dev/sdc"),
								v2_1.Device("/dev/sdd"),
							},
							Spares: 2,
						},
						{
							Name:  "durable",
							Level: "raid1",
							Devices: []v2_1.Device{
								v2_1.Device("/dev/sde"),
								v2_1.Device("/dev/sdf"),
							},
							Spares: 3,
						},
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{Version: types.MaxVersion.String()},
				Storage: types.Storage{
					Raid: []types.Raid{
						{
							Name:    "fast",
							Level:   "raid0",
							Devices: []types.Device{types.Device("/dev/sdc"), types.Device("/dev/sdd")},
							Spares:  2,
						},
						{
							Name:    "durable",
							Level:   "raid1",
							Devices: []types.Device{types.Device("/dev/sde"), types.Device("/dev/sdf")},
							Spares:  3,
						},
					},
				},
			}},
		},
		{
			in: in{config: v2_1.Config{
				Ignition: v2_1.Ignition{Version: v2_1.MaxVersion.String()},
				Storage: v2_1.Storage{
					Filesystems: []v2_1.Filesystem{
						{
							Name: "filesystem-0",
							Mount: &v2_1.Mount{
								Device: "/dev/disk/by-partlabel/ROOT",
								Format: "btrfs",
								Create: &v2_1.Create{
									Force:   true,
									Options: []v2_1.CreateOption{"-L", "ROOT"},
								},
								Label:          util.StrToPtr("ROOT"),
								Options:        []v2_1.MountOption{"--nodiscard"},
								UUID:           util.StrToPtr("8A7A6E26-5E8F-4CCA-A654-46215D4696AC"),
								WipeFilesystem: true,
							},
						},
						{
							Name: "filesystem-1",
							Mount: &v2_1.Mount{
								Device:         "/dev/disk/by-partlabel/DATA",
								Format:         "ext4",
								Label:          util.StrToPtr("DATA"),
								Options:        []v2_1.MountOption{"-b", "1024"},
								UUID:           util.StrToPtr("8A7A6E26-5E8F-4CCA-A654-DEADBEEF0101"),
								WipeFilesystem: false,
							},
						},
						{
							Name: "filesystem-2",
							Path: func(p string) *string { return &p }("/foo"),
						},
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{Version: types.MaxVersion.String()},
				Storage: types.Storage{
					Filesystems: []types.Filesystem{
						{
							Name: "filesystem-0",
							Mount: &types.Mount{
								Device: "/dev/disk/by-partlabel/ROOT",
								Format: "btrfs",
								Create: &types.Create{
									Force:   true,
									Options: []types.CreateOption{"-L", "ROOT"},
								},
								Label:          util.StrToPtr("ROOT"),
								Options:        []types.MountOption{"--nodiscard"},
								UUID:           util.StrToPtr("8A7A6E26-5E8F-4CCA-A654-46215D4696AC"),
								WipeFilesystem: true,
							},
						},
						{
							Name: "filesystem-1",
							Mount: &types.Mount{
								Device:         "/dev/disk/by-partlabel/DATA",
								Format:         "ext4",
								Label:          util.StrToPtr("DATA"),
								Options:        []types.MountOption{"-b", "1024"},
								UUID:           util.StrToPtr("8A7A6E26-5E8F-4CCA-A654-DEADBEEF0101"),
								WipeFilesystem: false,
							},
						},
						{
							Name: "filesystem-2",
							Path: util.StrToPtr("/foo"),
						},
					},
				},
			}},
		},
		{
			in: in{config: v2_1.Config{
				Ignition: v2_1.Ignition{Version: v2_1.MaxVersion.String()},
				Storage: v2_1.Storage{
					Files: []v2_1.File{
						{
							Node: v2_1.Node{
								Filesystem: "filesystem-0",
								Path:       "/opt/file1",
								User:       v2_1.NodeUser{ID: util.IntToPtr(500)},
								Group:      v2_1.NodeGroup{ID: util.IntToPtr(501)},
							},
							FileEmbedded1: v2_1.FileEmbedded1{
								Mode: 0664,
								Contents: v2_1.FileContents{
									Source: (&url.URL{
										Scheme: "data",
										Opaque: ",file1",
									}).String(),
									Verification: v2_1.Verification{
										Hash: util.StrToPtr("foobar"),
									},
								},
							},
						},
						{
							Node: v2_1.Node{
								Filesystem: "filesystem-0",
								Path:       "/opt/file2",
								User:       v2_1.NodeUser{ID: util.IntToPtr(502)},
								Group:      v2_1.NodeGroup{ID: util.IntToPtr(503)},
							},
							FileEmbedded1: v2_1.FileEmbedded1{
								Mode: 0644,
								Contents: v2_1.FileContents{
									Source: (&url.URL{
										Scheme: "data",
										Opaque: ",file2",
									}).String(),
									Compression: "gzip",
								},
							},
						},
						{
							Node: v2_1.Node{
								Filesystem: "filesystem-1",
								Path:       "/opt/file3",
								User:       v2_1.NodeUser{ID: util.IntToPtr(1000)},
								Group:      v2_1.NodeGroup{ID: util.IntToPtr(1001)},
							},
							FileEmbedded1: v2_1.FileEmbedded1{
								Mode: 0400,
								Contents: v2_1.FileContents{
									Source: (&url.URL{
										Scheme: "data",
										Opaque: ",file3",
									}).String(),
								},
							},
						},
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{Version: types.MaxVersion.String()},
				Storage: types.Storage{
					Files: []types.File{
						{
							Node: types.Node{
								Filesystem: "filesystem-0",
								Path:       "/opt/file1",
								User:       &types.NodeUser{ID: util.IntToPtr(500)},
								Group:      &types.NodeGroup{ID: util.IntToPtr(501)},
							},
							FileEmbedded1: types.FileEmbedded1{
								Mode: util.IntToPtr(0664),
								Contents: types.FileContents{
									Source: (&url.URL{
										Scheme: "data",
										Opaque: ",file1",
									}).String(),
									Verification: types.Verification{
										Hash: util.StrToPtr("foobar"),
									},
								},
							},
						},
						{
							Node: types.Node{
								Filesystem: "filesystem-0",
								Path:       "/opt/file2",
								User:       &types.NodeUser{ID: util.IntToPtr(502)},
								Group:      &types.NodeGroup{ID: util.IntToPtr(503)},
							},
							FileEmbedded1: types.FileEmbedded1{
								Mode: util.IntToPtr(0644),
								Contents: types.FileContents{
									Source: (&url.URL{
										Scheme: "data",
										Opaque: ",file2",
									}).String(),
									Compression: "gzip",
								},
							},
						},
						{
							Node: types.Node{
								Filesystem: "filesystem-1",
								Path:       "/opt/file3",
								User:       &types.NodeUser{ID: util.IntToPtr(1000)},
								Group:      &types.NodeGroup{ID: util.IntToPtr(1001)},
							},
							FileEmbedded1: types.FileEmbedded1{
								Mode: util.IntToPtr(0400),
								Contents: types.FileContents{
									Source: (&url.URL{
										Scheme: "data",
										Opaque: ",file3",
									}).String(),
								},
							},
						},
					},
				},
			}},
		},
		{
			in: in{config: v2_1.Config{
				Ignition: v2_1.Ignition{Version: v2_1.MaxVersion.String()},
				Storage: v2_1.Storage{
					Directories: []v2_1.Directory{
						{
							Node: v2_1.Node{
								Filesystem: "filesystem-1",
								Path:       "/opt/dir1",
								User:       v2_1.NodeUser{ID: util.IntToPtr(500)},
								Group:      v2_1.NodeGroup{ID: util.IntToPtr(501)},
							},
							DirectoryEmbedded1: v2_1.DirectoryEmbedded1{
								Mode: 0664,
							},
						},
						{
							Node: v2_1.Node{
								Filesystem: "filesystem-2",
								Path:       "/opt/dir2",
								User:       v2_1.NodeUser{ID: util.IntToPtr(502)},
								Group:      v2_1.NodeGroup{ID: util.IntToPtr(503)},
							},
							DirectoryEmbedded1: v2_1.DirectoryEmbedded1{
								Mode: 0644,
							},
						},
						{
							Node: v2_1.Node{
								Filesystem: "filesystem-2",
								Path:       "/opt/dir3",
								User:       v2_1.NodeUser{ID: util.IntToPtr(1000)},
								Group:      v2_1.NodeGroup{ID: util.IntToPtr(1001)},
							},
							DirectoryEmbedded1: v2_1.DirectoryEmbedded1{
								Mode: 0400,
							},
						},
					},
				}},
			},
			out: out{config: types.Config{
				Ignition: types.Ignition{Version: types.MaxVersion.String()},
				Storage: types.Storage{
					Directories: []types.Directory{
						{
							Node: types.Node{
								Filesystem: "filesystem-1",
								Path:       "/opt/dir1",
								User:       &types.NodeUser{ID: util.IntToPtr(500)},
								Group:      &types.NodeGroup{ID: util.IntToPtr(501)},
							},
							DirectoryEmbedded1: types.DirectoryEmbedded1{
								Mode: util.IntToPtr(0664),
							},
						},
						{
							Node: types.Node{
								Filesystem: "filesystem-2",
								Path:       "/opt/dir2",
								User:       &types.NodeUser{ID: util.IntToPtr(502)},
								Group:      &types.NodeGroup{ID: util.IntToPtr(503)},
							},
							DirectoryEmbedded1: types.DirectoryEmbedded1{
								Mode: util.IntToPtr(0644),
							},
						},
						{
							Node: types.Node{
								Filesystem: "filesystem-2",
								Path:       "/opt/dir3",
								User:       &types.NodeUser{ID: util.IntToPtr(1000)},
								Group:      &types.NodeGroup{ID: util.IntToPtr(1001)},
							},
							DirectoryEmbedded1: types.DirectoryEmbedded1{
								Mode: util.IntToPtr(0400),
							},
						},
					},
				},
			}},
		},
		{
			in: in{config: v2_1.Config{
				Ignition: v2_1.Ignition{Version: v2_1.MaxVersion.String()},
				Storage: v2_1.Storage{
					Links: []v2_1.Link{
						{
							Node: v2_1.Node{
								Filesystem: "filesystem-1",
								Path:       "/opt/link1",
								User:       v2_1.NodeUser{ID: util.IntToPtr(500)},
								Group:      v2_1.NodeGroup{ID: util.IntToPtr(501)},
							},
							LinkEmbedded1: v2_1.LinkEmbedded1{
								Hard:   false,
								Target: "/opt/file1",
							},
						},
						{
							Node: v2_1.Node{
								Filesystem: "filesystem-2",
								Path:       "/opt/link2",
								User:       v2_1.NodeUser{ID: util.IntToPtr(502)},
								Group:      v2_1.NodeGroup{ID: util.IntToPtr(503)},
							},
							LinkEmbedded1: v2_1.LinkEmbedded1{
								Hard:   true,
								Target: "/opt/file2",
							},
						},
						{
							Node: v2_1.Node{
								Filesystem: "filesystem-2",
								Path:       "/opt/link3",
								User:       v2_1.NodeUser{ID: util.IntToPtr(1000)},
								Group:      v2_1.NodeGroup{ID: util.IntToPtr(1001)},
							},
							LinkEmbedded1: v2_1.LinkEmbedded1{
								Hard:   true,
								Target: "/opt/file3",
							},
						},
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{Version: types.MaxVersion.String()},
				Storage: types.Storage{
					Links: []types.Link{
						{
							Node: types.Node{
								Filesystem: "filesystem-1",
								Path:       "/opt/link1",
								User:       &types.NodeUser{ID: util.IntToPtr(500)},
								Group:      &types.NodeGroup{ID: util.IntToPtr(501)},
							},
							LinkEmbedded1: types.LinkEmbedded1{
								Hard:   false,
								Target: "/opt/file1",
							},
						},
						{
							Node: types.Node{
								Filesystem: "filesystem-2",
								Path:       "/opt/link2",
								User:       &types.NodeUser{ID: util.IntToPtr(502)},
								Group:      &types.NodeGroup{ID: util.IntToPtr(503)},
							},
							LinkEmbedded1: types.LinkEmbedded1{
								Hard:   true,
								Target: "/opt/file2",
							},
						},
						{
							Node: types.Node{
								Filesystem: "filesystem-2",
								Path:       "/opt/link3",
								User:       &types.NodeUser{ID: util.IntToPtr(1000)},
								Group:      &types.NodeGroup{ID: util.IntToPtr(1001)},
							},
							LinkEmbedded1: types.LinkEmbedded1{
								Hard:   true,
								Target: "/opt/file3",
							},
						},
					},
				},
			}},
		},
		{
			in: in{v2_1.Config{
				Systemd: v2_1.Systemd{
					Units: []v2_1.Unit{
						{
							Name:     "test1.service",
							Enable:   true,
							Contents: "test1 contents",
							Dropins: []v2_1.Dropin{
								{
									Name:     "conf1.conf",
									Contents: "conf1 contents",
								},
								{
									Name:     "conf2.conf",
									Contents: "conf2 contents",
								},
							},
						},
						{
							Name:     "test2.service",
							Mask:     true,
							Contents: "test2 contents",
						},
						{
							Name:    "test3.service",
							Enabled: util.BoolToPtr(false),
						},
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{Version: types.MaxVersion.String()},
				Systemd: types.Systemd{
					Units: []types.Unit{
						{
							Name:     "test1.service",
							Enable:   true,
							Contents: "test1 contents",
							Dropins: []types.SystemdDropin{
								{
									Name:     "conf1.conf",
									Contents: "conf1 contents",
								},
								{
									Name:     "conf2.conf",
									Contents: "conf2 contents",
								},
							},
						},
						{
							Name:     "test2.service",
							Mask:     true,
							Contents: "test2 contents",
						},
						{
							Name:    "test3.service",
							Enabled: util.BoolToPtr(false),
						},
					},
				},
			}},
		},
		{
			in: in{v2_1.Config{
				Networkd: v2_1.Networkd{
					Units: []v2_1.Networkdunit{
						{
							Name:     "test1.network",
							Contents: "test1 contents",
						},
						{
							Name:     "test2.network",
							Contents: "test2 contents",
						},
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{Version: types.MaxVersion.String()},
				Networkd: types.Networkd{
					Units: []types.Networkdunit{
						{
							Name:     "test1.network",
							Contents: "test1 contents",
						},
						{
							Name:     "test2.network",
							Contents: "test2 contents",
						},
					},
				},
			}},
		},
		{
			in: in{v2_1.Config{
				Ignition: v2_1.Ignition{Version: v2_1.MaxVersion.String()},
				Passwd: v2_1.Passwd{
					Users: []v2_1.PasswdUser{
						{
							Name:              "user 1",
							PasswordHash:      util.StrToPtr("password 1"),
							SSHAuthorizedKeys: []v2_1.SSHAuthorizedKey{"key1", "key2"},
						},
						{
							Name:              "user 2",
							PasswordHash:      util.StrToPtr("password 2"),
							SSHAuthorizedKeys: []v2_1.SSHAuthorizedKey{"key3", "key4"},
							Create: &v2_1.Usercreate{
								UID:          util.IntToPtr(123),
								Gecos:        "gecos",
								HomeDir:      "/home/user 2",
								NoCreateHome: true,
								PrimaryGroup: "wheel",
								Groups:       []v2_1.UsercreateGroup{"wheel", "plugdev"},
								NoUserGroup:  true,
								System:       true,
								NoLogInit:    true,
								Shell:        "/bin/zsh",
							},
						},
						{
							Name:              "user 3",
							PasswordHash:      util.StrToPtr("password 3"),
							SSHAuthorizedKeys: []v2_1.SSHAuthorizedKey{"key5", "key6"},
							UID:               util.IntToPtr(123),
							Gecos:             "gecos",
							HomeDir:           "/home/user 2",
							NoCreateHome:      true,
							PrimaryGroup:      "wheel",
							Groups:            []v2_1.PasswdUserGroup{"wheel", "plugdev"},
							NoUserGroup:       true,
							System:            true,
							NoLogInit:         true,
							Shell:             "/bin/zsh",
						},
						{
							Name:              "user 4",
							PasswordHash:      util.StrToPtr("password 4"),
							SSHAuthorizedKeys: []v2_1.SSHAuthorizedKey{"key7", "key8"},
							Create:            &v2_1.Usercreate{},
						},
					},
					Groups: []v2_1.PasswdGroup{
						{
							Name:         "group 1",
							Gid:          func(i int) *int { return &i }(1000),
							PasswordHash: "password 1",
							System:       true,
						},
						{
							Name:         "group 2",
							PasswordHash: "password 2",
						},
					},
				},
			}},
			out: out{config: types.Config{
				Ignition: types.Ignition{Version: types.MaxVersion.String()},
				Passwd: types.Passwd{
					Users: []types.PasswdUser{
						{
							Name:              "user 1",
							PasswordHash:      util.StrToPtr("password 1"),
							SSHAuthorizedKeys: []types.SSHAuthorizedKey{"key1", "key2"},
						},
						{
							Name:              "user 2",
							PasswordHash:      util.StrToPtr("password 2"),
							SSHAuthorizedKeys: []types.SSHAuthorizedKey{"key3", "key4"},
							Create: &types.Usercreate{
								UID:          func(i int) *int { return &i }(123),
								Gecos:        "gecos",
								HomeDir:      "/home/user 2",
								NoCreateHome: true,
								PrimaryGroup: "wheel",
								Groups:       []types.UsercreateGroup{"wheel", "plugdev"},
								NoUserGroup:  true,
								System:       true,
								NoLogInit:    true,
								Shell:        "/bin/zsh",
							},
						},
						{
							Name:              "user 3",
							PasswordHash:      util.StrToPtr("password 3"),
							SSHAuthorizedKeys: []types.SSHAuthorizedKey{"key5", "key6"},
							UID:               util.IntToPtr(123),
							Gecos:             "gecos",
							HomeDir:           "/home/user 2",
							NoCreateHome:      true,
							PrimaryGroup:      "wheel",
							Groups:            []types.Group{"wheel", "plugdev"},
							NoUserGroup:       true,
							System:            true,
							NoLogInit:         true,
							Shell:             "/bin/zsh",
						},
						{
							Name:              "user 4",
							PasswordHash:      util.StrToPtr("password 4"),
							SSHAuthorizedKeys: []types.SSHAuthorizedKey{"key7", "key8"},
							Create:            &types.Usercreate{},
						},
					},
					Groups: []types.PasswdGroup{
						{
							Name:         "group 1",
							Gid:          func(i int) *int { return &i }(1000),
							PasswordHash: "password 1",
							System:       true,
						},
						{
							Name:         "group 2",
							PasswordHash: "password 2",
						},
					},
				},
			}},
		},
	}

	for i, test := range tests {
		config := TranslateFromV2_1(test.in.config)
		assert.Equal(t, test.out.config, config, "#%d: bad config", i)
	}
}
