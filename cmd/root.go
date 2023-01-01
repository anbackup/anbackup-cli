package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	adb "github.com/zach-klippenstein/goadb"
)

var rootCmd = &cobra.Command{
	Use: "anbackup-cli",
	Short: `
	####### #     # ######  ####### ####### #     # #     # #######          ####### #       #    
	#     # ##    # #    #  #     # #       #    #  #     # #     #  ####### #       #       #    
	####### # ### # ####### ####### #       #####   #     # #######          #       #       #    
	#     # #    ## #     # #     # #       #    #  #     # #                #       #       #    
	#     # #     # ####### #     # ####### #     # ####### #                ####### ####### #    
	`,
	Args:    cobra.MinimumNArgs(1),
	Version: "1.2.0",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var (
	log    = logrus.New()
	client *adb.Adb
	device *adb.Device
	isRoot = true
)

func init() {
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	cobra.OnInitialize(func() {
		client, _ = adb.NewWithConfig(adb.ServerConfig{})
		host := rootCmd.Flag("host").Value.String()
		port, _ := strconv.Atoi(rootCmd.Flag("port").Value.String())
		if host != "" && port != 0 {
			log.Infof("Connect %s:%d", host, port)
			client.Connect(host, port)
		}
		di, err := client.ListDevices()
		if err != nil {
			log.Fatal(err)
		}
		var deviceIndex = 0
		if len(di) < 1 {
			log.Fatal("Please connect the device")
		}
		if len(di) > 1 {
			for i, di2 := range di {
				log.Infof("%d : %s", i+1, di2.Serial)
			}
			log.Info("Please select the device to operate：")
			fmt.Scanf("%d", &deviceIndex)
			deviceIndex--
		}
		// 切换root
		s, err := exec.LookPath("adb")
		if err != nil {
			log.Fatal(err)
		}
		b, _ := exec.Command(s, "-s", di[deviceIndex].Serial, "root").CombinedOutput()
		if !(strings.Contains(string(b), "running") || strings.Contains(string(b), "adbd as root")) {
			log.Warn(string(b))
		}
		device = client.Device(adb.DeviceWithSerial(di[deviceIndex].Serial))
		s2, _ := device.RunCommand("whoami")
		isRoot = strings.Contains(s2, "root")
	})
	rootCmd.PersistentFlags().StringP("path", "p", "", "operation path")
	rootCmd.PersistentFlags().IP("host", nil, "adb connect ip")
	rootCmd.PersistentFlags().Int("port", 0, "adb connect port")
}
