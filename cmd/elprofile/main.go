package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
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
	Version  string
	Size     uint64
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

	ArgGenerate     bool
	ArgRequires     bool
	ArgProvides     bool
	ArgFiles        bool
	ArgVersion      bool
	ArgSize         bool
	ArgAll          bool
	ArgQuiet        bool
	ArgsSizePercent float64
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&ArgGenerate, "generate", "g", false, "Generate Profile")
	rootCmd.PersistentFlags().BoolVarP(&ArgAll, "all", "a", false, "Run all comparasions")

	rootCmd.PersistentFlags().BoolVarP(&ArgRequires, "requires", "r", false, "Only compare requires")
	rootCmd.PersistentFlags().BoolVarP(&ArgProvides, "provides", "p", false, "Only compare provides")
	rootCmd.PersistentFlags().BoolVarP(&ArgFiles, "files", "f", false, "Only compare files")
	rootCmd.PersistentFlags().BoolVarP(&ArgVersion, "version", "v", false, "Only compare Versions")
	rootCmd.PersistentFlags().BoolVarP(&ArgSize, "size", "s", false, "Only compare size")

	rootCmd.PersistentFlags().Float64VarP(&ArgsSizePercent, "sizepercent", "S", 2.5, "Size percent difference to warn")

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

	if ArgAll {
		ArgRequires = true
		ArgProvides = true
		ArgFiles = true
		ArgVersion = true
		ArgSize = true
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

			rpmInfo.Provides, err = rpmdb.PkgProvides(rpmName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not obtain provides for: %s\n", rpmName)
			}
			rpmInfo.Requires, err = rpmdb.PkgRequires(rpmName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not obtain requires for: %s\n", rpmName)
			}
			rpmInfo.Files, err = rpmdb.PkgFiles(rpmName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not obtain file list for: %s\n", rpmName)
			}
			rpmInfo.Version, err = rpmdb.PkgVersion(rpmName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not obtain package version for: %s\n", rpmName)
			}
			rpmInfo.Size, err = rpmdb.PkgSize(rpmName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not obtain package size for: %s\n", rpmName)
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

				requires, err := rpmdb.PkgRequires(rpmName)
				if err != nil {
					if !ArgQuiet {
						fmt.Fprintf(os.Stderr, "%-40.40s %s\n", rpmName, "NOT INSTALLED")
					}
					continue
				}

				requireMap := util.ArrayToMap(requires)

				for _, req := range packages.Rpms[rpmName].Requires {
					if _, ok := requireMap[req]; !ok {
						missingRequires[req] = true
						fmt.Printf("%-40.40s %-18s %s\n", rpmName, "MISSING REQUIRES:", req)
					}
				}
			}

			if ArgProvides {
				var err error

				provides, err := rpmdb.PkgProvides(rpmName)
				if err != nil {
					if !ArgQuiet {
						fmt.Fprintf(os.Stderr, "%-40.40s %s\n", rpmName, "NOT INSTALLED")
					}
					continue
				}

				providesMap := util.ArrayToMap(provides)

				for _, req := range packages.Rpms[rpmName].Provides {
					if _, ok := providesMap[req]; !ok {
						missingProvides[req] = true
						fmt.Printf("%-40.40s %-18s %s\n", rpmName, "MISSING REQUIRES:", req)
					}
				}
			}

			if ArgFiles {
				var err error

				files, err := rpmdb.PkgFiles(rpmName)
				if err != nil {
					if !ArgQuiet {
						fmt.Fprintf(os.Stderr, "%-40.40s %s\n", rpmName, "NOT INSTALLED")
					}
					continue
				}

				filesMap := util.ArrayToMap(files)

				for _, req := range packages.Rpms[rpmName].Files {
					if _, ok := filesMap[req]; !ok {
						fmt.Printf("%-40.40s %-18s %s\n", rpmName, "MISSING FILES:", req)
					}
				}
			}

			if ArgVersion {
				var err error

				version, err := rpmdb.PkgVersion(rpmName)
				if err != nil {
					if !ArgQuiet {
						fmt.Fprintf(os.Stderr, "%-40.40s %s\n", rpmName, "NOT INSTALLED")
					}
					continue
				}

				if version != packages.Rpms[rpmName].Version {
					fmt.Printf("%-40.40s %-18s local=%-25s  profile=%s\n", rpmName, "VERSION MISMATCH:", version, packages.Rpms[rpmName].Version)
				}

			}

			if ArgSize {
				var err error

				size, err := rpmdb.PkgSize(rpmName)
				if err != nil {
					if !ArgQuiet {
						fmt.Fprintf(os.Stderr, "%-40.40s %s\n", rpmName, "NOT INSTALLED")
					}
					continue
				}

				//fmt.Printf("size=%f\n", float64(size)-float64(packages.Rpms[rpmName].Size))

				if math.Abs(float64(size)-float64(packages.Rpms[rpmName].Size)) > (float64(packages.Rpms[rpmName].Size) * (ArgsSizePercent / 100)) {
					fmt.Printf("%-40.40s %-18s %+d bytes\n", rpmName, "SIZE MISMATCH:", int64(size-packages.Rpms[rpmName].Size))
				}

			}
		}

	} else {
		return errors.New("check usage")
	}

	return nil
}
