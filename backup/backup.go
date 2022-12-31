package backup

import (
	"anbackup-cli/config"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	adb "github.com/zach-klippenstein/goadb"
)

type Config struct {
	OutPath          string
	Packages         []*config.PackageConfig
	PackageName      string
	Package          *config.PackageConfig
	Count            int
	FailApkCount     int
	FailAppDataCount int
	Device           *adb.Device
	Log              *logrus.Logger
}

func New(c *Config) *Config {
	err := os.MkdirAll(c.OutPath, 7777)
	if err != nil {
		c.Log.Fatal(err)
	}
	_, err = c.Device.RunCommand(`mkdir /storage/emulated/0/anbackup`)
	if err != nil {
		c.Log.Fatal(err)
	}
	return c
}

func (c *Config) BackupApk() error {
	c.Log.Info(c.PackageName, " Backup apk file")
	s, err := c.Device.RunCommand("pm path " + c.PackageName)
	if err != nil {
		return err
	}
	s = strings.TrimSpace(strings.Replace(s, "package:", "", -1))
	s2 := strings.Split(s, "\n")
	for i, v := range s2 {
		rc, err := c.Device.OpenRead(v)
		if err != nil {
			return err
		}
		defer rc.Close()
		de, err := c.Device.Stat(v)
		if err != nil {
			return err
		}
		apkName := c.PackageName + strconv.Itoa(i) + ".apk"
		err = c.SaveFile(rc, de.Size, c.OutPath+"/"+apkName)
		if err != nil {
			return err
		}
	}
	c.Package.Apks = len(s2)
	return nil
}

func (c *Config) BackupDataFile() error {
	c.Log.Info(c.PackageName, " Backup app data")
	c.Log.Info(c.PackageName, " Zip app data to temp dir...")
	command := `su -c 'cd /data/data/` + c.PackageName + ` && tar -czf  /storage/emulated/0/anbackup/` + c.PackageName + `.tar.gz . && sleep 5s'`
	s, err := c.Device.RunCommand(command)
	if err != nil {
		return err
	}
	if s != "" {
		c.Log.Error(s)
	}
	zipPath := "/sdcard/anbackup/" + c.PackageName + ".tar.gz"
	rc, err := c.Device.OpenRead(zipPath)
	if err != nil {
		return err
	}
	defer rc.Close()
	de, err := c.Device.Stat(zipPath)
	if err != nil {
		return err
	}
	return c.SaveFile(rc, de.Size, c.OutPath+"/"+c.PackageName+".tar.gz")
}

func (c *Config) DeleteDataFile() error {
	s, err := c.Device.RunCommand("rm -r /storage/emulated/0/anbackup/" + c.PackageName + ".tar.gz")
	if err != nil {
		return err
	}
	if s == "" {
		c.Log.Info(c.PackageName, " Remove temp app data")
	} else {
		c.Log.Error(err)
	}
	return nil
}

func (c *Config) Start() error {
	var err error
	c.Count = len(c.Packages)
	for _, v := range c.Packages {
		c.Log.Info("Start backup ", v.PackageName)
		c.Package = v
		c.PackageName = v.PackageName
		if v.Apk {
			err = c.BackupApk()
			if err != nil {
				c.FailApkCount++
				v.Apk = false
				c.Log.Error("Backup apk error ", err)
			}
		}
		if v.AppData {
			err = c.BackupDataFile()
			if err != nil {
				c.FailAppDataCount++
				v.AppData = false
				c.Log.Error("Backup data file ", v, "error ", err)
			}
			err = c.DeleteDataFile()
			if err != nil {
				c.Log.Error("Delete data file  ", v, "error ", err)
			}
		}
	}
	return nil
}

func (c *Config) SaveFile(rc io.ReadCloser, filesize int32, filename string) error {
	buf := make([]byte, 8192)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	for {
		b, err := rc.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("\033[2K\r")
				return nil
			} else {
				return err
			}
		}
		fi, err := f.Stat()
		if err != nil {
			return err
		}
		fmt.Printf("\rSave %s ...  %d / %d", filename, fi.Size(), filesize)
		_, err = f.Write(buf[:b])
		if err != nil {
			return err
		}
	}
}
