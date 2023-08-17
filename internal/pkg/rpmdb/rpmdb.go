package rpmdb

import (
	"log"
	"os/exec"

	"github.com/gmkurtzer/elprofile/internal/pkg/util"
)

func PkgList() []string {
	//rpmqa, err := exec.Command("rpm", "-qa").Output()
	rpmqa, err := exec.Command("echo", "bash").Output()
	if err != nil {
		log.Fatal(err)
	}

	return util.BytesToArray(rpmqa)
}

func PkgInspect(pkgName, command string) []string {
	cmd, err := exec.Command("rpm", "-q", command, pkgName).Output()
	if err != nil {
		log.Fatal(err)
	}

	return util.BytesToArray(cmd)
}
