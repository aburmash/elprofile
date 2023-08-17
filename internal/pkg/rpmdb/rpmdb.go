package rpmdb

import (
	"os/exec"

	"github.com/gmkurtzer/elprofile/internal/pkg/util"
)

func PkgList() ([]string, error) {
	rpmqa, err := exec.Command("rpm", "-qa", "--qf", "%{NAME}\n").Output()
	//rpmqa, err := exec.Command("echo", "bash").Output()
	if err != nil {
		return nil, err
	}

	return util.BytesToArray(rpmqa), nil
}

func PkgInspect(pkgName, command string) ([]string, error) {
	cmd, err := exec.Command("rpm", "-q", command, pkgName).Output()
	if err != nil {
		return nil, err
	}

	return util.BytesToArray(cmd), nil
}
