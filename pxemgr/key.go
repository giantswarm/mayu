package pxemgr

import (
	"fmt"
	"os"
	"path"
)

const (
	vmlinuzFile      = "coreos_production_pxe.vmlinuz"
	initrdFile       = "coreos_production_pxe_image.cpio.gz"
	installImageFile = "coreos_production_image.bin.bz2"
	qemuImageFile    = "coreos_production_qemu_usr_image.squashfs"
	qemuKernelFile   = "coreos_production_qemu.vmlinuz"
)

func (mgr *pxeManagerT) pxeInstallImage(coreOSversion string) (*os.File, error) {
	return os.Open(path.Join(mgr.imagesCacheDir+"/"+coreOSversion, installImageFile))
}

func (mgr *pxeManagerT) pxeKernelImage(coreOSversion string) (*os.File, error) {
	return os.Open(path.Join(mgr.imagesCacheDir+"/"+coreOSversion, vmlinuzFile))
}

func (mgr *pxeManagerT) pxeInitRD(coreOSversion string) (*os.File, error) {
	return os.Open(path.Join(mgr.imagesCacheDir+"/"+coreOSversion, initrdFile))
}

func (mgr *pxeManagerT) qemuImage(coreOSversion string) (*os.File, error) {
	return os.Open(path.Join(mgr.imagesCacheDir+"/qemu/"+coreOSversion, qemuImageFile))
}

func (mgr *pxeManagerT) qemuImageSHA(coreOSversion string) (*os.File, error) {
	return os.Open(path.Join(mgr.imagesCacheDir+"/qemu/"+coreOSversion, fmt.Sprintf("%s.sha256", qemuImageFile)))
}

func (mgr *pxeManagerT) qemuKernel(coreOSversion string) (*os.File, error) {
	return os.Open(path.Join(mgr.imagesCacheDir+"/qemu/"+coreOSversion, qemuKernelFile))
}

func (mgr *pxeManagerT) qemuKernelSHA(coreOSversion string) (*os.File, error) {
	return os.Open(path.Join(mgr.imagesCacheDir+"/qemu/"+coreOSversion, fmt.Sprintf("%s.sha256", qemuKernelFile)))
}
