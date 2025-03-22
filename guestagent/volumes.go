package guestagent

import (
	"errors"
	"fmt"
	globalCommon "grove/common"
	"grove/config"
	guestCommon "grove/guestagent/common"
	"strings"
)

type FSBase interface {
	Format(devicePath string, timeout int) error
	CheckFormat(devicePath string) error
	Resize(devicePath string, online bool) error
}

type FSExt struct {
	FSType        string
	FormatOptions []string
}

type FSExt3 struct {
	FSExt
}

type FSExt4 struct {
	FSExt
}

type FSXFS struct {
	FSExt
}

func (f *FSExt) Format(devicePath string, timeout int) error {
	formatOptions := append(f.FormatOptions, devicePath)
	listParams := append([]string{"mkfs", "--type"}, formatOptions...)
	kwargs := map[string]interface{}{
		"log_output_on_error": true,
		"run_as_root":         true,
	}

	_, err := globalCommon.ExecuteWithTimeout(listParams, kwargs)

	return err
}

func (f *FSExt) CheckFormat(devicePath string) error {
	listParams := append([]string{"dumpe2fs", devicePath}, f.FormatOptions...)
	kwargs := map[string]interface{}{
		"log_output_on_error": true,
		"run_as_root":         true,
	}
	stdout, err := globalCommon.ExecuteWithTimeout(listParams, kwargs)
	if err != nil {
		if strings.Contains(err.Error(), "Wrong magic number") {
			logFmt := "Device '%s' did not seem to be '%s'."
			excFmt := "Device '%s' did not seem to be '%s'."
			globalCommon.LogAndRaise(logFmt, excFmt, devicePath, f.FSType)
		}
		logFmt := "Volume '%s' was not formatted."
		excFmt := "Volume '%s' was not formatted."
		globalCommon.LogAndRaise(logFmt, excFmt, devicePath, f.FSType)
	}
	if !strings.Contains(stdout, "has_journal") {
		logFmt := fmt.Sprintf("Volume '%s' was not formatted.", devicePath)
		return errors.New(logFmt)
	}
	return nil
}

func (f *FSExt) Resize(devicePath string, online bool) error {
	kwargs := map[string]interface{}{
		"log_output_on_error": true,
		"run_as_root":         true,
	}
	if !online {
		args := []string{"e2fsck", "-f", "-p", devicePath}
		_, err := globalCommon.ExecuteWithTimeout(args, kwargs)
		if err != nil {
			return err
		}
	}
	args := []string{"resize2fs", devicePath}
	_, err := globalCommon.ExecuteWithTimeout(args, kwargs)
	if err != nil {
		return err
	}
	return nil
}

func VolumeFS(fstype string, formatOptions ...string) FSBase {
	if "xfs" == fstype {
		return &FSExt{FSType: fstype, FormatOptions: formatOptions}
	}
	if "ext3" == fstype {
		return &FSExt3{FSExt{
			FormatOptions: formatOptions,
			FSType:        fstype,
		}}
	}
	if "ext4" == fstype {
		return &FSExt4{FSExt{
			FormatOptions: formatOptions,
			FSType:        fstype,
		}}
	}
	return nil
}

func (xfs *FSXFS) Format(devicePath string, timeout int) error {
	formatOptions := append(xfs.FormatOptions, devicePath)
	args := append([]string{"mkfs.xfs"}, formatOptions...)
	kwargs := map[string]interface{}{
		"log_output_on_error": true,
		"run_as_root":         true,
	}
	_, err := globalCommon.ExecuteWithTimeout(args, kwargs)
	if err != nil {
		logFmt := "Could not format device: %s."
		excFmt := "Could not format device: %s."
		globalCommon.LogAndRaise(logFmt, excFmt, devicePath, xfs.FSType)
	}
	return nil
}

func (xfs *FSXFS) CheckFormat(devicePath string) error {
	args := append(xfs.FormatOptions, devicePath)
	kwargs := map[string]interface{}{
		"log_output_on_error": true,
		"run_as_root":         true,
	}
	stdout, err := globalCommon.ExecuteWithTimeout(args, kwargs)
	if err != nil {
		logFmt := "Could not check format device: %s."
		excFmt := "Could not check format device: %s."
		globalCommon.LogAndRaise(logFmt, excFmt, devicePath, xfs.FSType)
	}
	if !strings.Contains(stdout, "not a valid XFS filesystem") {
		message := fmt.Sprintf("Volume '%s' does not appear to be formatted.", devicePath)
		return &globalCommon.GuestError{OriginalMessage: message}
	}
	return nil
}

func (xfs *FSXFS) Resize(devicePath string, online bool) error {
	kwargs := map[string]interface{}{
		"log_output_on_error": true,
		"run_as_root":         true,
	}

	argsRepair := []string{"xfs_repair", devicePath}
	_, err := globalCommon.ExecuteWithTimeout(argsRepair, kwargs)
	if err != nil {
		logFmt := "Error when check xfs_repair device: %s."
		excFmt := "Error when check xfs_repair device: %s."
		globalCommon.LogAndRaise(logFmt, excFmt, devicePath, xfs.FSType)
	}

	argsMount := []string{"mount", devicePath}
	_, err = globalCommon.ExecuteWithTimeout(argsMount, kwargs)
	if err != nil {
		logFmt := "Error when check mount device: %s."
		excFmt := "Error when check mount device: %s."
		globalCommon.LogAndRaise(logFmt, excFmt, devicePath, xfs.FSType)
	}

	argsGroupFs := []string{"xfs_growfs", devicePath}
	_, err = globalCommon.ExecuteWithTimeout(argsGroupFs, kwargs)
	if err != nil {
		logFmt := "Error when check xfs_growfs device: %s."
		excFmt := "Error when check xfs_growfs device: %s."
		globalCommon.LogAndRaise(logFmt, excFmt, devicePath, xfs.FSType)
	}

	argsUnmount := []string{"unmount", devicePath}
	_, err = globalCommon.ExecuteWithTimeout(argsUnmount, kwargs)
	if err != nil {
		logFmt := "Error when check unmount device: %s."
		excFmt := "Error when check unmount device: %s."
		globalCommon.LogAndRaise(logFmt, excFmt, devicePath, xfs.FSType)
	}
	return nil
}

type VolumeMountPoint struct {
	devicePath   string
	mountPoint   string
	volumeFsType string
	mountOptions []string
}

func (vmp *VolumeMountPoint) Mount() error {
	isExist := guestCommon.ExistsPath(vmp.mountPoint, true, true)
	if !isExist {

	}
}

type VolumeDevice struct {
	devicePath string
	volumeFs   FSBase
}

func (vd *VolumeDevice) MigrateData(sourceDir string, targetSubDir *string) error {
	// Đồng bộ dữ liệu từ source directory sang new volume, có thể tạo sub directory
	return nil
}

//func (vd *VolumeDevice) Mount(mountPoint string, writeToFstab bool) error {
//	// Mount disk và write to fstab
//	logrus.Debugf("Mounting %s at %s.", vd.devicePath, mountPoint)
//	mountPoint := Vo
//}
