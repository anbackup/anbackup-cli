package restore

import (
	"anbackup-cli/config"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	adb "github.com/zach-klippenstein/goadb"
)

type Config struct {
	RestoreConfig    *config.Config
	BasePath         string
	PackageName      string
	Package          *config.PackageConfig
	Count            int
	FailApkCount     int
	FailAppDataCount int
	Log              *logrus.Logger
	Device           *adb.Device
}

func New(c *Config) *Config {
	_, err := c.Device.RunCommand(`mkdir /sdcard/anbackup`)
	if err != nil {
		c.Log.Fatal(err)
	}
	return c
}

func (c *Config) Start() {
	var err error
	c.Count = len(c.RestoreConfig.Packages)
	for _, v := range c.RestoreConfig.Packages {
		c.PackageName = v.PackageName
		c.Package = v
		c.Log.Info("Start restore ", c.PackageName)
		if v.Apk {
			if err = c.InstallApk(); err != nil {
				c.FailApkCount++
				c.Log.Error(err)
			}
		}
		if v.AppData {
			if err = c.InstallAppData(); err != nil {
				c.FailAppDataCount++
				c.Log.Error(err)
			}
		}
	}
	if c.RestoreConfig.Contacts {
		if err = c.RestoreContacts(); err != nil {
			c.Log.Error(err)
		}
	}
}

func (c *Config) InstallApk() error {
	deviceApkPath := "/sdcard/anbackup/"
	for i := 0; i < c.Package.Apks; i++ {
		apkname := c.PackageName + strconv.Itoa(i) + ".apk"
		err := c.UploadFile(c.BasePath+"/"+apkname, deviceApkPath+apkname)
		if err != nil {
			return err
		}
	}

	// 安装apk
	c.Log.Info(c.PackageName, " Install apk ")
	cmd := "pm install "
	for i := 0; i < c.Package.Apks; i++ {
		cmd += deviceApkPath + c.PackageName + strconv.Itoa(i) + ".apk "
	}
	s, err := c.Device.RunCommand(cmd)
	if err != nil {
		return err
	}
	if strings.Contains(strings.ToLower(s), "success") {
		c.Log.Info(c.PackageName, " Install apk success")
	} else {
		return errors.New(s)
	}

	// 移除临时apk文件
	s, err = c.Device.RunCommand("rm -r " + deviceApkPath + "*")
	if err != nil {
		return err
	}
	if s != "" {
		c.Log.Warn(err)
	}
	return nil
}

func (c *Config) InstallAppData() error {

	deviceAppDataFilename := "/sdcard/anbackup/" + c.PackageName + ".tar.gz"
	err := c.UploadFile(c.BasePath+"/"+c.PackageName+".tar.gz", deviceAppDataFilename)
	if err != nil {
		return err
	}

	// 获取app权限组
	s, err := c.Device.RunCommand(`su -c 'cd /data/data/ && ls -l | grep "` + c.PackageName + `" '`)
	if err != nil {
		return err
	}

	if s == "" {
		c.Log.Warn(c.PackageName, " Get app data dir prem error")
		return nil
	}

	r := regexp.MustCompile(`u\d+_a\d+`)
	premGroup := r.FindString(s)

	// 解压应用数据
	c.Log.Info(c.PackageName, " Restore app data...")
	s, err = c.Device.RunCommand("su -c 'cd /sdcard/anbackup/ && tar -xzf ./" + c.PackageName + ".tar.gz -C /data/data/" + c.PackageName + " && sleep 5s'")
	if err != nil {
		return err
	}
	if s != "" {
		c.Log.Warn(s)
	}

	// 修改数据文件夹用户组
	c.Log.Info(c.PackageName, " Fix file permissions ", premGroup)
	s, err = c.Device.RunCommand("su -c 'chown -R " + premGroup + ":" + premGroup + " /data/data/" + c.PackageName + "/.'")
	if err != nil {
		return err
	}
	if s != "" {
		c.Log.Warn(s)
	}

	// 移除临时app data文件
	s, err = c.Device.RunCommand("rm -r " + deviceAppDataFilename)
	if err != nil {
		return err
	}
	if s != "" {
		c.Log.Warn(err)
	}
	return nil
}

func (c *Config) RestoreContacts() error {
	c.Log.Info("Start restore contacts")
	accountName := time.Now().Format("2006-1-2-15-04-05")

	// 读取本地联系人文件

	b, err := os.ReadFile(c.BasePath + "/" + "contacs.txt")
	if err != nil {
		return err
	}

	contacts := strings.Split(string(b), "Row")

	contactsMap := make(map[string][]string)
	for _, v := range contacts {
		if v == "" {
			continue
		}
		display_name := ""
		data1 := ""
		match := regexp.MustCompile("display_name=(.+?),").FindStringSubmatch(v)
		if len(match) > 1 {
			display_name = match[1]
		}
		if display_name == "" {
			continue
		}
		match = regexp.MustCompile("data1=(.+?)\n").FindStringSubmatch(v)
		if len(match) > 1 {
			data1 = match[1]
		}
		if display_name == data1 {
			continue
		}
		contactsMap[display_name] = append(contactsMap[display_name], data1)
	}

	for k, v := range contactsMap {
		c.Log.Info("Start restre contact ", k)
		// 创建联系人
		cmd := fmt.Sprintf("content insert --uri content://com.android.contacts/raw_contacts --bind account_type:s:%s --bind account_name:s:%s", accountName, accountName)
		s, err := c.Device.RunCommand(cmd)
		if err != nil {
			c.Log.Warn(err)
			continue
		}
		if s != "" {
			c.Log.Warn(s)
		}
		// 查询联系人id
		cmd = fmt.Sprintf(`content query --uri content://com.android.contacts/raw_contacts  --where "account_name='%s' and display_name='' or  display_name IS NULL and deleted=0"`, accountName)
		s, err = c.Device.RunCommand(cmd)
		if err != nil {
			c.Log.Warn(err)
			continue
		}
		if s == "" {
			c.Log.Error(s)
			continue
		}
		contact_id := regexp.MustCompile("contact_id=(.+?),").FindStringSubmatch(s)[1]
		// 添加姓名
		cmd = fmt.Sprintf(`content insert --uri content://com.android.contacts/data --bind raw_contact_id:i:%s --bind mimetype:s:vnd.android.cursor.item/name --bind data1:s:%s`, contact_id, k)
		s, err = c.Device.RunCommand(cmd)
		if err != nil {
			c.Log.Warn(err)
			continue
		}
		if s != "" {
			c.Log.Error(s)
		}
		// 添加号码

		for _, v2 := range v {
			cmd = fmt.Sprintf(`content insert --uri content://com.android.contacts/data --bind raw_contact_id:i:%s --bind mimetype:s:vnd.android.cursor.item/phone_v2 --bind data1:s:%s`, contact_id, v2)
			s, err = c.Device.RunCommand(cmd)
			if err != nil {
				c.Log.Warn(err)
				continue
			}
			if s != "" {
				c.Log.Error(s)
			}
		}
	}
	return nil
}

func (c *Config) UploadFile(localFilename string, remoteFilename string) error {
	// 读取本地apk文件
	f, err := os.Open(localFilename)
	if err != nil {
		return err
	}
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	// 写入设备

	wc, err := c.Device.OpenWrite(remoteFilename, os.ModePerm, time.Now())
	if err != nil {
		return err
	}
	defer wc.Close()
	defer f.Close()

	buf := make([]byte, 8192)
	uploadSize := 0

	for {
		b, err := f.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("\033[2K\r")
				break
			} else {
				return err
			}
		}
		uploadSize += len(buf)
		fmt.Printf("\rUpload file  %s ...  %d / %d", localFilename, uploadSize, fi.Size())
		_, err = wc.Write(buf[:b])
		if err != nil {
			return err
		}
	}
	return nil
}
