package pxemgr

import (
	"os"
	"path"
)

const (
	vmlinuzFile = "flatcar_production_pxe.vmlinuz"
	initrdFile  = "flatcar_production_pxe_image.cpio.gz"
)

func (mgr *pxeManagerT) pxeKernelImage(flatcarVersion string) (*os.File, error) {
	return os.Open(path.Join(mgr.imagesCacheDir+"/"+flatcarVersion, vmlinuzFile))
}

func (mgr *pxeManagerT) pxeInitRD(flatcarVersion string) (*os.File, error) {
	return os.Open(path.Join(mgr.imagesCacheDir+"/"+flatcarVersion, initrdFile))
}
