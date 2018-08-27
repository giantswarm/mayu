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

package types

import (
	"reflect"
	"testing"

	"github.com/coreos/ignition/config/shared/errors"
	"github.com/coreos/ignition/config/validate/report"
)

func TestFilesystemFormatValidate(t *testing.T) {
	type in struct {
		format FilesystemFormat
	}
	type out struct {
		rep report.Report
	}

	tests := []struct {
		in  in
		out out
	}{
		{
			in:  in{format: FilesystemFormat("ext4")},
			out: out{},
		},
		{
			in:  in{format: FilesystemFormat("btrfs")},
			out: out{},
		},
		{
			in:  in{format: FilesystemFormat("")},
			out: out{rep: report.ReportFromError(errors.ErrFilesystemInvalidFormat, report.EntryError)},
		},
	}

	for i, test := range tests {
		rep := test.in.format.Validate()
		if !reflect.DeepEqual(test.out.rep, rep) {
			t.Errorf("#%d: bad report: want %v, got %v", i, test.out.rep, rep)
		}
	}
}
