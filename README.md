# anbackup-cli

这是一个能直接备份手机数据到PC上的命令行工具。制作他的初衷是因为在手机上备份数据时总是空间不足，又因为没有sd卡或外置存储所以就整了个直接备份到PC上的工具。此工具还可以用于换机。

## 支持备份/恢复
  - [x] apk文件
  - [x] 拆分apk文件
  - [x] 应用数据 `root`
  - [x] 通讯录
  - [ ] 短信
  - [ ] 通话记录
  - [ ] wifi
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
