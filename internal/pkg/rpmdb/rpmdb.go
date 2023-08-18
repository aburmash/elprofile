package rpmdb

import (
	"os/exec"
	"regexp"

	"github.com/gmkurtzer/elprofile/internal/pkg/util"
)

func PkgList() ([]string, error) {
	rpmqa, err := exec.Command("rpm", "-qa", "--qf", "%{NAME}\n").Output()
	//rpmqa, err := exec.Command("echo", "ImageMagick").Output()
	if err != nil {
		return nil, err
	}

	return util.BytesToArray(rpmqa), nil
}

func PkgInspect(pkgName, command string) ([]string, error) {
	var ret []string

	cmd, err := exec.Command("rpm", "-q", command, pkgName).Output()
	if err != nil {
		return nil, err
	}

	for _, prov := range util.BytesToArray(cmd) {
		r := regexp.MustCompile(" [<|>|=] ")
		name := r.Split(prov, 2)

		//name := strings.SplitN(prov, " (<|>|=)", 2)
		ret = append(ret, name[0])
	}

	return ret, nil
}
