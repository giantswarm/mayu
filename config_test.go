package main

import (
	"testing"

	"github.com/giantswarm/mayu/fs"
)

func TestHTTPSCertConfigValidation(t *testing.T) {
	cases := []struct {
		conf           configuration
		expectedResult bool
		expectedError  error
	}{
		{configuration{NoSecure: true, HTTPSCertFile: "", HTTPSKeyFile: ""}, true, nil},
		{configuration{NoSecure: true, HTTPSCertFile: "certfile", HTTPSKeyFile: ""}, true, nil},
		{configuration{NoSecure: true, HTTPSCertFile: "", HTTPSKeyFile: "keyfile"}, true, nil},
		{configuration{NoSecure: true, HTTPSCertFile: "certfile", HTTPSKeyFile: "keyfile"}, true, nil},
		{configuration{NoSecure: false, HTTPSCertFile: "", HTTPSKeyFile: ""}, false, ErrNotAllCertFilesProvided},
		{configuration{NoSecure: false, HTTPSCertFile: "certfile", HTTPSKeyFile: ""}, false, ErrNotAllCertFilesProvided},
		{configuration{NoSecure: false, HTTPSCertFile: "", HTTPSKeyFile: "keyfile"}, false, ErrNotAllCertFilesProvided},
		{configuration{NoSecure: false, HTTPSCertFile: "certfile", HTTPSKeyFile: "keyfile"}, true, nil},
	}

	for _, c := range cases {
		result, err := c.conf.ValidateHTTPCertificateUsage()

		if result != c.expectedResult {
			t.Errorf("expected function ValidateHTTPCertificateUsage() to return %v for configuration :%#v", c.expectedResult, c.conf)
		}

		if err != c.expectedError {
			t.Errorf("expected function ValidateHTTPCertificateUsage() to return error '%s' but got '%s' for configuration :%#v", c.expectedError, err, c.conf)
		}

	}
}

func TestHTTPCertConfigFileStatValidation(t *testing.T) {
	cases := []struct {
		conf           configuration
		expectedResult bool
		expectedError  error
	}{
		{ // TLS secured connections are turned off, so no cert files should be needed.
			configuration{
				filesystem: fs.FakeFilesystem{},
				NoSecure:   true,
			},
			true,
			nil,
		},
		{ // Both files are provided but TLS secured connections are turned off, which should be ok too.
			configuration{
				filesystem: fs.NewFakeFilesystemWithFiles([]fs.FakeFile{
					fs.NewFakeFile("cert.pem", "foobar"),
					fs.NewFakeFile("key.pem", "barbaz"),
				}),
				HTTPSCertFile: "cert.pem",
				HTTPSKeyFile:  "key.pem",
				NoSecure:      true,
			},
			true,
			nil,
		},
		{
			configuration{
				filesystem:    fs.FakeFilesystem{},
				HTTPSCertFile: "cert.pem",
				HTTPSKeyFile:  "key.pem",
			},
			false,
			ErrHTTPSCertFileNotRedable,
		},
		{ // Only a key file is provided which should result in a missing cert file error.
			configuration{
				filesystem: fs.NewFakeFilesystemWithFiles([]fs.FakeFile{
					fs.NewFakeFile("key.pem", "foobar"),
				}),
				HTTPSCertFile: "cert.pem",
				HTTPSKeyFile:  "key.pem",
			},
			false,
			ErrHTTPSCertFileNotRedable,
		},
		{ // Only a cert file is provided which should result in a missing key file error.
			configuration{
				filesystem: fs.NewFakeFilesystemWithFiles([]fs.FakeFile{
					fs.NewFakeFile("cert.pem", "foobar"),
				}),
				HTTPSCertFile: "cert.pem",
				HTTPSKeyFile:  "key.pem",
			},
			false,
			ErrHTTPSKeyFileNotReadable,
		},
		{ // Both files are provided which should be ok.
			configuration{
				filesystem: fs.NewFakeFilesystemWithFiles([]fs.FakeFile{
					fs.NewFakeFile("cert.pem", "foobar"),
					fs.NewFakeFile("key.pem", "barbaz"),
				}),
				HTTPSCertFile: "cert.pem",
				HTTPSKeyFile:  "key.pem",
			},
			true,
			nil,
		},
	}

	for _, c := range cases {
		result, err := c.conf.ValidateHTTPCertificateFileExistance()

		if result != c.expectedResult {
			t.Errorf("Expected result to be %v but got %v for configuration %#v", c.expectedResult, result, c.conf)
		}

		if err != c.expectedError {
			t.Errorf("Expected error to be '%s' but got '%s' for configuration %#v", c.expectedError, err, c.conf)
		}
	}

}
