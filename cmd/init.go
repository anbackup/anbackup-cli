package cmd

import (
	"anbackup-cli/config"
	"strings"

	"github.com/spf13/cobra"
	adb "github.com/zach-klippenstein/goadb"
)

var (
	all            bool
	disableApk     bool
	disableAppData bool
)

func init() {
	initCmd.Flags().BoolVar(&all, "all", false, "load all package")
	initCmd.Flags().BoolVar(&disableApk, "disable-apk", false, "disable backup apk")
	initCmd.Flags().BoolVar(&disableAppData, "disable-app-data", false, "disable backup app data")
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the backup specification file",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Get device info")
		di, err := device.DeviceInfo()
		if err != nil {
			log.Fatal(err)
		}
		log.Info("Get packages")
		packages, err := getPackages(device, all)
		if err != nil {
			log.Fatal(err)
		}
		if !isRoot {
			disableAppData = true
			log.Warn("This device is not rooted so all app data will not be backed up")
		}
		pc := []*config.PackageConfig{}

		for _, v := range packages {
			if v == "" {
				continue
			}
			pc = append(pc, &config.PackageConfig{
				PackageName: v,
				Apk:         !disableApk,
				AppData:     !disableAppData,
			})
		}
		c := config.Config{
			DeviceInfo:  di.DeviceInfo,
			IsRoot:      isRoot,
			AddressBook: true,
			Message:     true,
			CallRecords: true,
			Packages:    pc,
		}
		err = c.Save("config.json")
		if err != nil {
			log.Fatal(err)
		}
		log.Info("Initialize success")
	},
}

func getPackages(device *adb.Device, all bool) (packages []string, err error) {
	cmd := "pm list packages -3"
	if all {
		cmd = "pm list packages"
	}
	s, err := device.RunCommand(cmd)
	if err != nil {
		return nil, err
	}
	s2 := strings.Replace(s, "package:", "", -1)
	return strings.Split(s2, "\n"), nil
}
