package cmd

import (
	"anbackup-cli/config"
	"anbackup-cli/restore"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	restoreCmd.Flags().StringP("path", "p", "", "restore backup path")
	restoreCmd.MarkFlagRequired("path")
	rootCmd.AddCommand(restoreCmd)
}

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore apk and app data to phone",
	Run: func(cmd *cobra.Command, args []string) {
		path := cmd.Flag("path").Value.String()
		b, err := os.ReadFile(path + "/config.json")
		if err != nil {
			log.Fatalf("load  %s/config.json error %v", path, err)
		}
		c := config.Config{}
		err = json.Unmarshal(b, &c)
		if err != nil {
			log.Fatal(err)
		}
		serial, err := device.Serial()
		if err != nil {
			log.Fatal(err)
		}
		if c.DeviceInfo != serial {
			log.Warn("You are backing up different device configurations, is this unsafe to continue (Enter to continue)")
			fmt.Scanln()
		}
		r := restore.New(&restore.Config{
			BasePath:      cmd.Flag("path").Value.String(),
			Log:           log,
			Device:        device,
			RestoreConfig: &c,
		})
		r.Start()
		log.Info("Total processing ", r.Count)
		log.Info("Fail restore apk count ", r.FailApkCount)
		log.Info("Fail restore app data count ", r.FailAppDataCount)
		log.Info("Restore complete")
	},
}
