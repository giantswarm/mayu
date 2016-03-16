package main

import (
	"testing"

	"github.com/giantswarm/mayu/fs"
)

func TestHTTPSCertConfigValidation(t *testing.T) {
	cases := []struct {
		globalFlags    MayuFlags
		expectedResult bool
		expectedError  error
	}{
		{MayuFlags{noTLS: true, tlsCertFile: "", tlsKeyFile: ""}, true, nil},
		{MayuFlags{noTLS: true, tlsCertFile: "certfile", tlsKeyFile: ""}, true, nil},
		{MayuFlags{noTLS: true, tlsCertFile: "", tlsKeyFile: "keyfile"}, true, nil},
		{MayuFlags{noTLS: true, tlsCertFile: "certfile", tlsKeyFile: "keyfile"}, true, nil},
		{MayuFlags{noTLS: false, tlsCertFile: "", tlsKeyFile: ""}, false, ErrNotAllCertFilesProvided},
		{MayuFlags{noTLS: false, tlsCertFile: "certfile", tlsKeyFile: ""}, false, ErrNotAllCertFilesProvided},
		{MayuFlags{noTLS: false, tlsCertFile: "", tlsKeyFile: "keyfile"}, false, ErrNotAllCertFilesProvided},
		{MayuFlags{noTLS: false, tlsCertFile: "certfile", tlsKeyFile: "keyfile"}, true, nil},
	}

	for _, c := range cases {
		result, err := c.globalFlags.ValidateHTTPCertificateUsage()

		if result != c.expectedResult {
			t.Errorf("expected function ValidateHTTPCertificateUsage() to return %v for configuration :%#v", c.expectedResult, c.globalFlags)
		}

		if err != c.expectedError {
			t.Errorf("expected function ValidateHTTPCertificateUsage() to return error '%s' but got '%s' for configuration :%#v", c.expectedError, err, c.globalFlags)
		}

	}
}

func TestHTTPCertConfigFileStatValidation(t *testing.T) {
	cases := []struct {
		globalFlags    MayuFlags
		expectedResult bool
		expectedError  error
	}{
		{ // TLS connections are turned off, so no cert files should be needed.
			MayuFlags{
				filesystem: fs.FakeFilesystem{},
				noTLS:      true,
			},
			true,
			nil,
		},
		{ // Both files are provided but TLS connections are turned off, which should be ok too.
			MayuFlags{
				filesystem: fs.NewFakeFilesystemWithFiles([]fs.FakeFile{
					fs.NewFakeFile("cert.pem", "foobar"),
					fs.NewFakeFile("key.pem", "barbaz"),
				}),
				tlsCertFile: "cert.pem",
				tlsKeyFile:  "key.pem",
				noTLS:       true,
			},
			true,
			nil,
		},
		{
			MayuFlags{
				filesystem:  fs.FakeFilesystem{},
				tlsCertFile: "cert.pem",
				tlsKeyFile:  "key.pem",
			},
			false,
			ErrHTTPSCertFileNotRedable,
		},
		{ // Only a key file is provided which should result in a missing cert file error.
			MayuFlags{
				filesystem: fs.NewFakeFilesystemWithFiles([]fs.FakeFile{
					fs.NewFakeFile("key.pem", "foobar"),
				}),
				tlsCertFile: "cert.pem",
				tlsKeyFile:  "key.pem",
			},
			false,
			ErrHTTPSCertFileNotRedable,
		},
		{ // Only a cert file is provided which should result in a missing key file error.
			MayuFlags{
				filesystem: fs.NewFakeFilesystemWithFiles([]fs.FakeFile{
					fs.NewFakeFile("cert.pem", "foobar"),
				}),
				tlsCertFile: "cert.pem",
				tlsKeyFile:  "key.pem",
			},
			false,
			ErrHTTPSKeyFileNotReadable,
		},
		{ // Both files are provided which should be ok.
			MayuFlags{
				filesystem: fs.NewFakeFilesystemWithFiles([]fs.FakeFile{
					fs.NewFakeFile("cert.pem", "foobar"),
					fs.NewFakeFile("key.pem", "barbaz"),
				}),
				tlsCertFile: "cert.pem",
				tlsKeyFile:  "key.pem",
			},
			true,
			nil,
		},
	}

	for _, c := range cases {
		result, err := c.globalFlags.ValidateHTTPCertificateFileExistance()

		if result != c.expectedResult {
			t.Errorf("Expected result to be %v but got %v for configuration %#v", c.expectedResult, result, c.globalFlags)
		}

		if err != c.expectedError {
			t.Errorf("Expected error to be '%s' but got '%s' for configuration %#v", c.expectedError, err, c.globalFlags)
		}
	}

}
