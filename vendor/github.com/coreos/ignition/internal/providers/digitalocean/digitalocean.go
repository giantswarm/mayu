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

// The digitalocean provider fetches a remote configuration from the
// digitalocean user-data metadata service URL.

package digitalocean

import (
	"net/url"

	"github.com/coreos/ignition/config/types"
	"github.com/coreos/ignition/config/validate/report"
	"github.com/coreos/ignition/internal/log"
	"github.com/coreos/ignition/internal/providers"
	"github.com/coreos/ignition/internal/providers/util"
	"github.com/coreos/ignition/internal/resource"

	"golang.org/x/net/context"
)

var (
	userdataUrl = url.URL{
		Scheme: "http",
		Host:   "169.254.169.254",
		Path:   "metadata/v1/user-data",
	}
)

func FetchConfig(logger *log.Logger, client *resource.HttpClient) (types.Config, report.Report, error) {
	data := resource.FetchConfig(logger, client, context.Background(), userdataUrl)
	if data == nil {
		return types.Config{}, report.Report{}, providers.ErrNoProvider
	}

	return util.ParseConfig(logger, data)
}
