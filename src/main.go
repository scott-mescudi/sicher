package main

import (
	"fmt"

	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"

	"gobackup/pkg"
)


//TODO 2: make each file a go routine with limit of 10 and 100mb per file
//TODO 3: implement worker pool for TODO 2

type Config struct {
	SrcDir      string `toml:"srcDir"`
	DstDir      string `toml:"dstDir"`
	MaxWorkers  int    `toml:"maxWorkers"`
	MemUsage    int    `toml:"memUsage"`
	BackupFreq  int    `toml:"backupFreq"`
	MaxFileSize int    `toml:"maxFileSize"`

	RestrictedDirs       map[string]bool `toml:"restrictedDirs"`
	RestrictedFiles      map[string]bool `toml:"restrictedFiles"`
	RestrictedExtensions map[string]bool `toml:"restrictedExtensions"`
}




func main() {
	config, err := LoadConfig("config.toml")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	start := time.Now()
	config.StartBackup()
	fmt.Printf("Elapsed: %v\n", time.Since(start))
}

func LoadConfig(filePath string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (cf *Config) StartBackup() {
	var srcfiles = make(map[string]bool)
	var ddddd = []string{}
	defer pkg.Clean(cf.SrcDir, cf.DstDir)

	filepath.WalkDir(cf.SrcDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			tpath := filepath.Base(path)
            if _, ok := cf.RestrictedDirs[tpath]; ok {
				return filepath.SkipDir
			}else{
				if path != cf.SrcDir{
					ddddd = append(ddddd, tpath)
				}
			}
        }else{
			fpath := filepath.Base(path)
			if _, ok := cf.RestrictedFiles[fpath]; ok {
				return filepath.SkipDir
			}

			ext := filepath.Ext(path)
			if _, ok := cf.RestrictedExtensions[ext]; ok {
				return filepath.SkipDir
			}

			if path != cf.SrcDir {
				srcfiles[path] = true
			}
		}




		return nil
	})



	for _, dir := range ddddd{
		fs := filepath.Join(cf.DstDir, dir)
		os.Mkdir(fs, 0700)
	}

	for i := range srcfiles {
		x := strings.TrimPrefix(i, cf.SrcDir)
		dstfile := filepath.Join(cf.DstDir, x)
		srcfile := filepath.Join(i)
		work(srcfile, dstfile, 1024)

	}

}

func work(srcfile, dstfile string, buf int) {
	ok, err := pkg.FileCheck(srcfile, dstfile)
	if err != nil || !ok {
		return
	}

	err = pkg.CopyFile(srcfile, dstfile, buf)
	if err!= nil {
        fmt.Println(err)
    }

}

