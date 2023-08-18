package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/gmkurtzer/elprofile/internal/pkg/rpmdb"
	"github.com/gmkurtzer/elprofile/internal/pkg/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
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

var (
	rootCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "elprofile [flags]",
		Short:                 "EL Profile and Compliance Checker",
		Long:                  "Enterprise Linux system profiling and checker (from OpenELA).",
		RunE:                  CobraRunE,
		SilenceUsage:          true,
		SilenceErrors:         true,
	}

	ArgGenerate bool
	ArgRequires bool
	ArgProvides bool
	ArgFiles    bool
	ArgQuiet    bool
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&ArgGenerate, "generate", "g", false, "Generate Profile")
	rootCmd.PersistentFlags().BoolVarP(&ArgRequires, "requires", "r", false, "Only test requires")
	rootCmd.PersistentFlags().BoolVarP(&ArgProvides, "provides", "p", false, "Only test provides")
	rootCmd.PersistentFlags().BoolVarP(&ArgFiles, "files", "f", false, "Only test files")
	rootCmd.PersistentFlags().BoolVarP(&ArgQuiet, "quiet", "q", false, "Only show issues")

}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(255)
	}
}

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	var packages Packages

	var testfile string

	if len(args) == 1 {
		testfile = args[0]
	}

	if !ArgFiles && !ArgProvides && !ArgRequires {
		ArgRequires = true
		ArgProvides = true
		ArgFiles = true
	}

	if ArgGenerate {

		fmt.Fprintf(os.Stderr, "Generating local system profile...\n")

		packages.Rpms = make(map[string]*RpmInfo)

		rpmlist, err := rpmdb.PkgList()
		if err != nil {
			return errors.Wrap(err, "could not generate package listing")
		}

		for i := 0; i < len(rpmlist); i++ {
			var rpmInfo RpmInfo
			var rpmName = rpmlist[i]
			var err error

			rpmInfo.Provides, err = rpmdb.PkgInspect(rpmName, "-provides")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not obtain provides for: %s\n", rpmName)
			}
			rpmInfo.Requires, err = rpmdb.PkgInspect(rpmName, "-requires")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not obtain requires for: %s\n", rpmName)
			}
			rpmInfo.Files, err = rpmdb.PkgInspect(rpmName, "-l")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not obtain file list for: %s\n", rpmName)
			}

			packages.Rpms[rpmName] = &rpmInfo
		}

		rpmYaml, err := yaml.Marshal(&packages)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(rpmYaml))
	} else if testfile != "" {
		var packages Packages
		var err error

		fmt.Fprintf(os.Stderr, "Comparing against template: %s\n", testfile)

		rpmYaml, err := ioutil.ReadFile(testfile)
		if err != nil {
			return errors.Wrap(err, "could not open profile file")
		}

		err = yaml.Unmarshal(rpmYaml, &packages)
		if err != nil {
			return errors.Wrap(err, "could not unmarshal profile")
		}

		for rpmName := range packages.Rpms {

			missingProvides := make(map[string]bool)
			missingRequires := make(map[string]bool)

			if ArgRequires {
				var err error

				requires, err := rpmdb.PkgInspect(rpmName, "--requires")
				if err != nil {
					if !ArgQuiet {
						fmt.Fprintf(os.Stderr, "%-40s %s\n", rpmName, "NOT INSTALLED")
					}
					continue
				}

				requireMap := util.ArrayToMap(requires)

				LookingFor := util.ArrayMatch(`.*`, packages.Rpms[rpmName].Requires)
				LookingFor = util.ArrayNotMatch(`=`, LookingFor)

				for _, req := range LookingFor {
					if _, ok := requireMap[req]; !ok {
						missingRequires[req] = true
						//fmt.Printf("%-20s %-25s %s\n", "MISSING REQUIRES:", rpmName, req)
					}
				}
			}

			if ArgProvides {
				var err error

				provides, err := rpmdb.PkgInspect(rpmName, "--provides")
				if err != nil {
					if !ArgQuiet {
						fmt.Fprintf(os.Stderr, "%-40s %s\n", rpmName, "NOT INSTALLED")
					}
					continue
				}

				providesMap := util.ArrayToMap(provides)

				LookingFor := util.ArrayMatch(`.*`, packages.Rpms[rpmName].Provides)
				LookingFor = util.ArrayNotMatch(`=`, LookingFor)

				for _, req := range LookingFor {
					if _, ok := providesMap[req]; !ok {
						missingProvides[req] = true
						//fmt.Printf("%-20s %-25s %s\n", "MISSING PROVIDES:", rpmName, req)
					}
				}
			}

			for req := range missingRequires {
				if _, ok := missingProvides[req]; !ok {
					fmt.Printf("%-40s %-18s %s\n", rpmName, "MISSING REQUIRES:", req)
				}
			}

			for req := range missingProvides {
				if _, ok := missingRequires[req]; !ok {
					fmt.Printf("%-40s %-18s %s\n", rpmName, "MISSING PROVIDES:", req)
				}
			}

			if ArgFiles {
				var err error

				files, err := rpmdb.PkgInspect(rpmName, "-l")
				if err != nil {
					if !ArgQuiet {
						fmt.Fprintf(os.Stderr, "%-40s %s\n", rpmName, "NOT INSTALLED")
					}
					continue
				}

				filesMap := util.ArrayToMap(files)

				LookingFor := util.ArrayMatch(`.*`, packages.Rpms[rpmName].Files)
				LookingFor = util.ArrayNotMatch(`/.build-id/`, LookingFor)

				for _, req := range LookingFor {
					if _, ok := filesMap[req]; !ok {
						fmt.Printf("%-40s %-18s %s\n", rpmName, "MISSING FILES:", req)
					}
				}
			}
		}

	} else {
		return errors.New("check usage")
	}

	return nil
}
