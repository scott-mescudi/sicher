package main

import (
	"context"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"os"
	"path/filepath"
	"strings"
	"gobackup/pkg"
	"github.com/BurntSushi/toml"
	"github.com/natefinch/lumberjack"
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

type Worker struct {
	Srcfile, Dstfile string
	Buf, MaxFileSize int
}

func main() {
		logfile := &lumberjack.Logger{
		Filename:   "backup_daemon.log", 
		MaxSize:    10,                 
		MaxBackups: 3,                  
		MaxAge:     28,                  
		Compress:   true,                
	}

	// Set log output to logfile managed by lumberjack
	log.SetOutput(logfile)

	config, err := LoadConfig("config.toml")
	if err != nil {
		log.Printf("Error loading config: %v\n", err)
		return
	}

	_, err = os.Stat(config.SrcDir)
	if os.IsNotExist(err) {
		log.Printf("Cannot find source directory %v\n", config.SrcDir)
		return
	}

	_, err = os.Stat(config.DstDir)
	if os.IsNotExist(err) {
		log.Printf("Cannot find destination directory %v\n", config.DstDir)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		cancel()
	}()

	backupInterval := time.Duration(config.BackupFreq) * time.Minute

	ticker := time.NewTicker(backupInterval)
	defer ticker.Stop()

	config.StartBackup(ctx) // First backup run

	for {
		select {
		case <-ticker.C:
			config.StartBackup(ctx)
		}
	}
}

func LoadConfig(filePath string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		return nil, err
	}
	log.Printf("Configuration loaded from %v", filePath)
	return &config, nil
}

func (cf *Config) StartBackup(ctx context.Context) {
	var srcfiles = make(map[string]bool)
	var dirsToCreate = []string{}
	defer pkg.Clean(cf.SrcDir, cf.DstDir)

	log.Printf("Starting backup process from %v to %v", cf.SrcDir, cf.DstDir)

	filepath.WalkDir(cf.SrcDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Printf("Error walking through directory: %v", err)
			return err
		}

		if d.IsDir() {
			tpath := filepath.Base(path)
			if _, ok := cf.RestrictedDirs[tpath]; ok {
				log.Printf("Skipping restricted directory: %v", tpath)
				return filepath.SkipDir
			} else {
				if path != cf.SrcDir {
					dirsToCreate = append(dirsToCreate, tpath)
				}
			}
		} else {
			fpath := filepath.Base(path)
			if _, ok := cf.RestrictedFiles[fpath]; ok {
				log.Printf("Skipping restricted file: %v", fpath)
				return filepath.SkipDir
			}

			ext := filepath.Ext(path)
			if _, ok := cf.RestrictedExtensions[ext]; ok {
				log.Printf("Skipping restricted file extension: %v", ext)
				return filepath.SkipDir
			}

			if path != cf.SrcDir {
				srcfiles[path] = true
			}
		}

		return nil
	})

	var dirWg sync.WaitGroup
	for _, dir := range dirsToCreate {
		dirWg.Add(1)
		go func(dir string) {
			defer dirWg.Done()
			fs := filepath.Join(cf.DstDir, dir)
			err := os.Mkdir(fs, 0666)
			if err != nil {
				log.Printf("Error creating directory: %v", err)
			}
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
		log.Printf("Scheduled task for copying %v to %v", srcfile, dstfile)
	}

	close(tasks)
	wg.Wait()
	log.Printf("Backup completed.")
}

func worker(ctx context.Context, tasks chan Worker, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case task, ok := <-tasks:
			if !ok {
				return
			}
			log.Printf("Worker handling task: copying %v to %v", task.Srcfile, task.Dstfile)
			work(task.Srcfile, task.Dstfile, task.Buf, task.MaxFileSize)
		case <-ctx.Done():
			log.Println("Worker received cancel signal")
			return
		}
	}
}

func work(srcfile, dstfile string, buf, maxFileSize int) {
	ok, err := pkg.FileCheck(srcfile, dstfile, maxFileSize)
	if err != nil {
		log.Printf("Error checking file %v: %v", srcfile, err)
		return
	}

	if !ok {
		log.Printf("File %v is either too large or already exists in the destination.", srcfile)
		return
	}

	err = pkg.CopyFile(srcfile, dstfile, buf)
	if err != nil {
		log.Printf("Error copying file %v to %v: %v", srcfile, dstfile, err)
	}
	log.Printf("Successfully copied %v to %v", srcfile, dstfile)
}
