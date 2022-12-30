# anbackup-cli

这是一个能用备份应用和应用数据到PC上的命令行工具。能够备份到PC,但是应用数据得通过手机用户存储中转，所以你的存储空间必须剩余有占用空间最大的应用的大小，且备份、恢复应用数据需要设备root。

```

        ####### #     # ######  ####### ####### #     # #     # #######          ####### #       #
        #     # ##    # #    #  #     # #       #    #  #     # #     #  ####### #       #       #
        ####### # ### # ####### ####### #       #####   #     # #######          #       #       #
        #     # #    ## #     # #     # #       #    #  #     # #                #       #       #
        #     # #     # ####### #     # ####### #     # ####### #                ####### ####### #

Usage:
  anbackup-cli [command]

Available Commands:
  backup      Backup apk and app data to PC
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  init        Initialize the backup specification file
  restore     Restore apk and app data to phone

Flags:
  -h, --help          help for anbackup-cli
      --host ip       adb connect ip
  -p, --path string   operation path
      --port int      adb connect port
  -v, --version       version for anbackup-cli

Use "anbackup-cli [command] --help" for more information about a command.
```
