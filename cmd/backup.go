package cmd

import (
	"anbackup-cli/backup"
	"anbackup-cli/config"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	backupCmd.Flags().StringP("path", "p", "./backups/"+time.Now().Format("2006-1-2-15-04-05"), "backup output path")
	backupCmd.Flags().StringP("config", "c", "./config.json", "backup config path")
	rootCmd.AddCommand(backupCmd)
}

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup apk and app data to PC",
	Run: func(cmd *cobra.Command, args []string) {
		configPath := cmd.Flag("config").Value.String()
		path := cmd.Flag("path").Value.String()

		// 读取配置文件
		b2, err := os.ReadFile(configPath)
		if err != nil {
			log.Fatal("can not load config.json,you need to run init to initialize", err)
		}
		c := config.Config{}
		err = json.Unmarshal(b2, &c)
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
		var b = backup.New(&backup.Config{
			BasePath:     path,
			Log:          log,
			Device:       device,
			BackupConfig: &c,
		})
		b.Start()

		// 剔除未备份的应用
		for i := 0; i < len(c.Packages); i++ {
			if !c.Packages[i].Apk && !c.Packages[i].AppData {
				c.Packages = append(c.Packages[:i], c.Packages[i+1:]...)
				i--
			}
		}

		// 保存config.json
		err = c.Save(path + "/config.json")
		if err != nil {
			log.Fatal(err)
		}
		log.Info("Total processing ", b.Count)
		log.Info("Fail backup apk count ", b.FailApkCount)
		log.Info("Fail backup app data count ", b.FailAppDataCount)
		log.Info("Backup complete")
	},
}
