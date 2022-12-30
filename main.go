package main

import (
	"anbackup-cli/cmd"
	"fmt"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("There was an error while running, please make sure you have adb installed on your device")
			fmt.Println(err)
		}
	}()
	cmd.Execute()
}
