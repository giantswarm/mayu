// Copyright 2015 CoreOS, Inc.
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

package util

import (
	"path/filepath"
)

func SystemdUnitsPath() string {
	return filepath.Join("etc", "systemd", "system")
}

func SystemdRuntimeUnitsPath() string {
	return filepath.Join("run", "systemd", "system")
}

func SystemdRuntimeUnitWantsPath(unitName string) string {
	return filepath.Join("run", "systemd", "system", unitName+".wants")
}

func NetworkdUnitsPath() string {
	return filepath.Join("etc", "systemd", "network")
}

func SystemdDropinsPath(unitName string) string {
	return filepath.Join("etc", "systemd", "system", unitName+".d")
}

func SystemdRuntimeDropinsPath(unitName string) string {
	return filepath.Join("run", "systemd", "system", unitName+".d")
}

func NetworkdDropinsPath(unitName string) string {
	return filepath.Join("etc", "systemd", "network", unitName+".d")
}
