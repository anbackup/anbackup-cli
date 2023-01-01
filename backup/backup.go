package backup

import (
	"anbackup-cli/config"
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	adb "github.com/zach-klippenstein/goadb"
)

type Config struct {
	BasePath         string
	BackupConfig     *config.Config
	PackageName      string
	Package          *config.PackageConfig
	Count            int
	FailApkCount     int
	FailAppDataCount int
	Device           *adb.Device
	Log              *logrus.Logger
}

func New(c *Config) *Config {
	err := os.MkdirAll(c.BasePath, 7777)
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
		apkName := c.PackageName + strconv.Itoa(i) + ".apk"
		err = c.SaveFile(rc, c.BasePath+"/"+apkName)
		if err != nil {
			return err
		}
	}
	c.Package.Apks = len(s2)
	return nil
}

func (c *Config) BackupDataFile() error {
	c.Log.Info(c.PackageName, " Backup app data")
	packagePath := "/data/data/" + c.PackageName
	cmd := "find " + packagePath + "/. -type f  -not -empty && sleep 3"
	s, err := c.Device.RunCommand(cmd)
	if s == "" || err != nil {
		return err
	}
	s2 := strings.Split(s, "\n")
	thread := 0
	wg := new(sync.WaitGroup)
	for _, v := range s2 {
		if v == "" {
			continue
		}
		go func(v string) {
			thread++
			wg.Add(1)
			defer func() {
				wg.Done()
				thread--
			}()
			var getFile = func() error {
				rc, err := c.Device.OpenRead(v)

				if err != nil {
					return err
				}
				defer rc.Close()
				localFilename := strings.Replace(v, packagePath, c.BasePath+"/"+c.PackageName, -1)
				s := strings.Split(localFilename, "/")
				err = os.MkdirAll(strings.Join(s[:len(s)-1], "/"), os.ModePerm)
				if err != nil {
					return err
				}
				err = c.SaveFile(rc, localFilename)
				if err != nil {
					return err
				}
				return nil
			}

			for {
				if err := getFile(); err != nil {
					if strings.Contains(err.Error(), "ServerNotAvailable") {
						time.Sleep(5 * time.Second)
						continue
					}
					c.Log.Error(err)
					break
				}
				break
			}

		}(v)
		if thread >= 128 {
			wg.Wait()
		}
	}
	wg.Wait()
	if err := c.archiveAppData(); err != nil {
		return err
	}
	if err := os.RemoveAll(c.BasePath + "/" + c.PackageName); err != nil {
		c.Log.Error(err)
	}
	return nil
}

func (c *Config) archiveAppData() error {
	c.Log.Info(c.PackageName, " Zip app data...")
	dir := c.BasePath + "/" + c.PackageName
	f, err := os.Create(dir + ".tar.gz")
	if err != nil {
		return err
	}
	defer f.Close()
	gw := gzip.NewWriter(f)
	defer gw.Close()
	w := tar.NewWriter(gw)
	defer w.Close()

	recursionDir(dir+"/", func(f string) error {
		fi2, err := os.Stat(f)
		if err != nil {
			c.Log.Error(err)
		}
		h, err := tar.FileInfoHeader(fi2, "")
		if err != nil {
			c.Log.Error(err)
		}
		h.Name = strings.Replace(f, dir+"//", "./", -1)
		if err := w.WriteHeader(h); err != nil {
			return err
		}
		if fi2.IsDir() {
			return nil
		}
		f2, err := os.Open(f)
		if err != nil {
			return err
		}
		io.Copy(w, f2)
		w.Flush()
		return nil
	})

	return nil
}

func recursionDir(path string, callBack func(f string) error) error {
	fi, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, fi2 := range fi {
		if fi2.IsDir() {
			if err := recursionDir(path+"/"+fi2.Name(), callBack); err != nil {
				return err
			}
		}
		if err := callBack(path + "/" + fi2.Name()); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) BackupContacts() error {
	c.Log.Info("Start backup contacts")
	s, err := c.Device.RunCommand("content query --uri content://com.android.contacts/data --projection display_name:data1 && sleep 5s")
	if err != nil {
		return err
	}
	if strings.Contains(s, "No result found") {
		return errors.New("contact data not found, so skipping backup")
	}
	f, err := os.Create(c.BasePath + "/" + "contacs.txt")
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(s))
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) Start() {
	var err error
	if c.BackupConfig.Contacts {
		err = c.BackupContacts()
		if err != nil {
			c.Log.Error("Backup contacts error ", err)
			c.BackupConfig.Contacts = false
		}
	}
	c.Count = len(c.BackupConfig.Packages)
	for _, v := range c.BackupConfig.Packages {
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
				c.Log.Error("Backup data file ", c.PackageName, "error ", err)
			}
		}
	}
}

func (c *Config) SaveFile(rc io.ReadCloser, filename string) error {
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
		fmt.Printf("\r Save %s ...  %d ", c.PackageName, fi.Size())
		_, err = f.Write(buf[:b])
		if err != nil {
			return err
		}
	}
}
