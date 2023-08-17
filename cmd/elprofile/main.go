package main

import (
	"fmt"
	"log"

	"github.com/gmkurtzer/elprofile/internal/pkg/rpmdb"

	"gopkg.in/yaml.v3"
)

type Packages struct {
	Rpms map[string]*RpmInfo
}

type RpmInfo struct {
	Provides []string
	Requires []string
	Files    []string
}

func main() {
	var packages Packages

	packages.Rpms = make(map[string]*RpmInfo)

	rpmlist := rpmdb.PkgList()

	for i := 0; i < len(rpmlist); i++ {
		var rpmInfo RpmInfo
		var rpmName = rpmlist[i]

		rpmInfo.Provides = rpmdb.PkgInspect(rpmName, "-provides")
		rpmInfo.Requires = rpmdb.PkgInspect(rpmName, "-requires")
		rpmInfo.Files = rpmdb.PkgInspect(rpmName, "-l")

		packages.Rpms[rpmName] = &rpmInfo

	}

	rpmYaml, err := yaml.Marshal(&packages)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(rpmYaml))

}
