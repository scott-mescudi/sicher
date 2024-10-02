package main

import (
	"context"
	"fmt"
	"os/signal"
	"sync"
	"syscall"

	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"gobackup/pkg"
)


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

type Worker struct{
	Srcfile, Dstfile string
	Buf, MaxFileSize int
}


func main() {
    config, err := LoadConfig("config.toml")
    if err != nil {
        fmt.Printf("Error loading config: %v\n", err)
        return
    }

    ctx, cancel := context.WithCancel(context.Background())

    sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)


    go func() {
        <-sigs 
        cancel() 
    }()

    
    config.StartBackup(ctx)
}


func LoadConfig(filePath string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (cf *Config) StartBackup(ctx context.Context) {
	var srcfiles = make(map[string]bool)
	var dirsToCreate = []string{}
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
					dirsToCreate = append(dirsToCreate, tpath)
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


	var dirWg sync.WaitGroup
	for _, dir := range dirsToCreate{
		dirWg.Add(1)
		go func(dir string){
            defer dirWg.Done()
            fs := filepath.Join(cf.DstDir, dir)
            os.Mkdir(fs, 0666)
        }(dir)
	}
	dirWg.Wait()

	tasks := make(chan Worker, len(srcfiles))
	var wg sync.WaitGroup
	for i := 0; i < cf.MaxWorkers; i++ {
		wg.Add(1)
		go worker(ctx, tasks, &wg)
	}

	for i := range srcfiles {
		x := strings.TrimPrefix(i, cf.SrcDir)
		dstfile := filepath.Join(cf.DstDir, x)
		srcfile := filepath.Join(i)
		tasks <- Worker{srcfile, dstfile, cf.MemUsage, cf.MaxFileSize}

	}

	close(tasks)
	wg.Wait()
}

func worker(ctx context.Context, tasks chan Worker, wg *sync.WaitGroup) {
    defer wg.Done()
    for {
        select {
        case task, ok := <-tasks:
            if !ok {
                return 
            }
            work(task.Srcfile, task.Dstfile, task.Buf, task.MaxFileSize)
        case <-ctx.Done(): 
            return 
        }
    }
}

func work(srcfile, dstfile string, buf, maxFileSize int) {
	ok, err := pkg.FileCheck(srcfile, dstfile, maxFileSize)
	if err != nil || !ok {
		return
	}

	err = pkg.CopyFile(srcfile, dstfile, buf)
	if err!= nil {
        fmt.Println(err)
    }
}