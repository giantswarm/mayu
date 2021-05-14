package pxemgr

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"path"
	"text/template"

	"github.com/giantswarm/microerror"
)

// Files is map[string]string for files that we fetched from disk and then filled with data.
type Files map[string]string

func (mgr *pxeManagerT) RenderFiles(ctx interface{}) (*Files, error) {
	files := Files{}
	dirList, err := ioutil.ReadDir(mgr.filesDir)
	if err != nil {
		_ = mgr.logger.Log("level", "error", "message", fmt.Sprintf("Failed to read files dir: %s", mgr.filesDir), "stack", err)
		return nil, microerror.Mask(err)
	}

	for _, dir := range dirList {
		fileList, err := ioutil.ReadDir(path.Join(mgr.filesDir, dir.Name()))
		if err != nil {
			_ = mgr.logger.Log("level", "error", "message", fmt.Sprintf("Failed to read dir: %s", path.Join(mgr.filesDir, dir.Name())), "stack", err)
			return nil, microerror.Mask(err)
		}

		for _, file := range fileList {
			tmpl, err := template.ParseFiles(path.Join(mgr.filesDir, dir.Name(), file.Name()))
			if err != nil {
				_ = mgr.logger.Log("level", "error", "message", fmt.Sprintf("Failed to file: %s", path.Join(mgr.filesDir, dir.Name(), file.Name())), "stack", err)
				return nil, microerror.Mask(err)
			}

			var data bytes.Buffer
			err = tmpl.Execute(&data, ctx)
			if err != nil {
				_ = mgr.logger.Log("level", "error", "message", fmt.Sprintf("Failed to execute tmpl for  %s", path.Join(mgr.filesDir, dir.Name(), file.Name())), "stack", err)
				return nil, microerror.Mask(err)
			}

			files[dir.Name()+"/"+file.Name()] = base64.StdEncoding.EncodeToString(data.Bytes())
		}
	}
	return &files, nil
}
