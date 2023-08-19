package rpmdb

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/gmkurtzer/elprofile/internal/pkg/util"
)

func PkgList() ([]string, error) {
	rpmqa, err := exec.Command("rpm", "-qa", "--qf", "%{NAME}-%{VERSION}.%{ARCH}\n").Output()
	//rpmqa, err := exec.Command("echo", "systemd-libs").Output()
	if err != nil {
		return nil, err
	}

	return util.ArrayNotMatch(`.(none)`, util.BytesToArray(rpmqa)), nil
}

func rpmRun(a ...string) ([]string, error) {
	cmd, err := exec.Command("rpm", a...).Output()
	if err != nil {
		return nil, err
	}

	return util.BytesToArray(cmd), nil
}

func PkgInspect(pkgName, command string) ([]string, error) {
	return rpmRun("-q", command, pkgName)
}

func PkgRequires(pkgName string) ([]string, error) {
	var ret []string

	out, err := rpmRun("-q", "--requires", pkgName)

	for _, prov := range out {
		regex := regexp.MustCompile(" <|>|= ")
		name := regex.Split(prov, 2)

		ret = append(ret, strings.TrimSpace(name[0]))
	}

	return ret, err
}

func PkgProvides(pkgName string) ([]string, error) {
	var ret []string

	out, err := rpmRun("-q", "--provides", pkgName)

	for _, prov := range out {
		regex := regexp.MustCompile(" <|>|= ")
		name := regex.Split(prov, 2)

		ret = append(ret, strings.TrimSpace(name[0]))
	}

	return ret, err
}

func PkgFiles(pkgName string) ([]string, error) {
	query, err := rpmRun("-q", "-l", pkgName)

	ret := util.ArrayNotMatch(`/.build-id/`, query)

	return ret, err
}

func PkgVersion(pkgName string) (string, error) {
	ret, err := rpmRun("-q", "--qf", "%{VERSION}-%{RELEASE}", pkgName)
	if len(ret) == 0 {
		return "", err
	}
	return ret[0], err
}

func PkgSize(pkgName string) (uint64, error) {
	size, err := rpmRun("-q", "--qf", "%{SIZE}", pkgName)
	if len(size) == 0 {
		return 0, err
	}
	ret, err := strconv.ParseUint(size[0], 0, 64)
	return ret, err
}
