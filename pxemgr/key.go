package pxemgr

import (
	"os"
	"path"
)

const (
	vmlinuzFile      = "coreos_production_pxe.vmlinuz"
	initrdFile       = "coreos_production_pxe_image.cpio.gz"
)

func (mgr *pxeManagerT) pxeKernelImage(coreOSversion string) (*os.File, error) {
	return os.Open(path.Join(mgr.imagesCacheDir+"/"+coreOSversion, vmlinuzFile))
}

func (mgr *pxeManagerT) pxeInitRD(coreOSversion string) (*os.File, error) {
	return os.Open(path.Join(mgr.imagesCacheDir+"/"+coreOSversion, initrdFile))
}