package cmd

import (
	"fmt"
	"os"
	"strconv"

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
	Version: "1.0.0",
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
			client.Connect(host, port)
		}
		device = client.Device(adb.AnyDevice())
	})
	rootCmd.PersistentFlags().StringP("path", "p", "", "operation path")
	rootCmd.PersistentFlags().IP("host", nil, "adb connect ip")
	rootCmd.PersistentFlags().Int("port", 0, "adb connect port")
}
