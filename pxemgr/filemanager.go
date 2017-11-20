package pxemgr

import (
	"bytes"
	"github.com/golang/glog"
	"encoding/base64"
	"io/ioutil"
	"path"
	"text/template"
)

type Files map[string]string

func (mgr *pxeManagerT) RenderFiles(ctx interface{}) *Files {

	files := Files{}

	dirList, err := ioutil.ReadDir(mgr.filesDir)
	if err != nil {
		glog.Fatalf("Failed to read files dir: %s, error: %#v", mgr.filesDir, err)
	}

	for _, dir := range dirList {
		fileList, err := ioutil.ReadDir(path.Join(mgr.filesDir, dir.Name()))
		if err != nil {
			glog.Errorf("Failed to read dir: %s, error: %#v", path.Join(mgr.filesDir, dir.Name()), err)
		}
		glog.V(8).Infoln("Reading Dir: ", dir.Name())

		for _, file := range fileList {
			glog.V(8).Infoln("Reading File: ", file.Name())
			tmpl, err := template.ParseFiles(path.Join(mgr.filesDir, dir.Name(), file.Name()))
			if err != nil {
				glog.Errorf("Failed to file: %s, error: %#v", path.Join(mgr.filesDir, dir.Name(), file.Name()), err)
			}
			var data bytes.Buffer
			tmpl.Execute(&data, ctx)

			files[dir.Name()+"/"+file.Name()] = base64.StdEncoding.EncodeToString(data.Bytes())
			glog.V(8).Infoln("encoded file as: ", files[dir.Name()+"/"+file.Name()])
		}
	}
	return &files
}
