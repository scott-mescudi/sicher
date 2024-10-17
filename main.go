package main

import (
	"context"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"os"
	"fmt"
	"path/filepath"
	"strings"
	"github.com/scott-mescudi/sicher/internal"
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

const (
	INFO  = "[INFO]"
	ERROR = "[ERROR]"
	WARN  = "[WARN]"
)

func logWithColor(level, message string) {
	log.Printf("%s %s\n", level, message)
}

// Struct and functions...

func main() {
	logfile := &lumberjack.Logger{
		Filename:   "sicher.log",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     28,
		Compress:   true,
	}

	// Set log output to logfile managed by lumberjack
	log.SetOutput(logfile)

	config, err := LoadConfig("sicher.toml")
	if err != nil {
		logWithColor(ERROR, fmt.Sprintf("Error loading config: %v", err))
		return
	}

	_, err = os.Stat(config.SrcDir)
	if os.IsNotExist(err) {
		logWithColor(ERROR, fmt.Sprintf("Cannot find source directory %v", config.SrcDir))
		return
	}

	_, err = os.Stat(config.DstDir)
	if os.IsNotExist(err) {
		logWithColor(ERROR, fmt.Sprintf("Cannot find destination directory %v", config.DstDir))
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

	for range time.Tick(backupInterval) {
		select {
		case <-ctx.Done():
			return
		default:
			config.StartBackup(ctx)
		}
	}
}

func LoadConfig(filePath string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		return nil, err
	}
	logWithColor(INFO, fmt.Sprintf("Configuration loaded from %v", filePath))
	
	if ok := verifyConfig(&config); !ok {
		logWithColor(ERROR, "Invalid configuration detected. Exiting...")
		return nil, fmt.Errorf("invalid configuration detected")
	}
	return &config, nil
}

func verifyConfig(config *Config) bool {
	if config.MaxWorkers <= 0 {
		logWithColor(ERROR, fmt.Sprintf("Invalid maxWorkers value. Expected a positive integer, got: %v", config.MaxWorkers))
		return false
	}

	if config.MemUsage <= 0 {
		logWithColor(ERROR, fmt.Sprintf("Invalid memUsage value. Expected a positive integer, got: %v", config.MemUsage))
		return false
	}

	if config.BackupFreq <= 0 {
		logWithColor(ERROR, fmt.Sprintf("Invalid backupFreq value. Expected a positive integer, got: %v", config.BackupFreq))
		return false
	}

	if config.MaxFileSize <= 0 {
		logWithColor(ERROR, fmt.Sprintf("Invalid maxFileSize value. Expected a positive integer, got: %v", config.MaxFileSize))
		return false
	}

	return true
}

func (cf *Config) StartBackup(ctx context.Context) {
	var srcfiles = make(map[string]bool)
	var dirsToCreate = []string{}
	defer pkg.Clean(cf.SrcDir, cf.DstDir)

	logWithColor(INFO, fmt.Sprintf("Starting backup process from %v to %v", cf.SrcDir, cf.DstDir))

	filepath.WalkDir(cf.SrcDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			logWithColor(ERROR, fmt.Sprintf("Error walking through directory: %v", err))
			return err
		}

		if d.IsDir() {
			tpath := filepath.Base(path)
			if _, ok := cf.RestrictedDirs[tpath]; ok {
				logWithColor(WARN, fmt.Sprintf("Skipping restricted directory: %v", tpath))
				return filepath.SkipDir
			} else {
				if path != cf.SrcDir {
					dirsToCreate = append(dirsToCreate, tpath)
				}
			}
		} else {
			fpath := filepath.Base(path)
			if _, ok := cf.RestrictedFiles[fpath]; ok {
				logWithColor(WARN, fmt.Sprintf("Skipping restricted file: %v", fpath))
				return nil
			}

			ext := filepath.Ext(path)
			if _, ok := cf.RestrictedExtensions[ext]; ok {
				logWithColor(WARN, fmt.Sprintf("Skipping restricted file extension: %v", ext))
				return nil
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
				logWithColor(ERROR, fmt.Sprintf("Error creating directory: %v", err))
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
		logWithColor(INFO, fmt.Sprintf("Scheduled task for copying %v to %v", srcfile, dstfile))
	}

	close(tasks)
	wg.Wait()
	logWithColor(INFO, "Backup completed.")
}

func worker(ctx context.Context, tasks chan Worker, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case task, ok := <-tasks:
			if !ok {
				return
			}
			logWithColor(INFO, fmt.Sprintf("Worker handling task: copying %v to %v", task.Srcfile, task.Dstfile))
			work(task.Srcfile, task.Dstfile, task.Buf, task.MaxFileSize)
		case <-ctx.Done():
			logWithColor(WARN, "Worker received cancel signal")
			return
		}
	}
}

func work(srcfile, dstfile string, buf, maxFileSize int) {
	ok, err := pkg.FileCheck(srcfile, dstfile, maxFileSize)
	if err != nil {
		logWithColor(ERROR, fmt.Sprintf("Error checking file %v: %v", srcfile, err))
		return
	}

	if !ok {
		logWithColor(WARN, fmt.Sprintf("File %v is either too large or already exists in the destination.", srcfile))
		return
	}

	err = pkg.CopyFile(srcfile, dstfile, buf)
	if err != nil {
		logWithColor(ERROR, fmt.Sprintf("Error copying file %v to %v: %v", srcfile, dstfile, err))
	}
	logWithColor(INFO, fmt.Sprintf("Successfully copied %v to %v", srcfile, dstfile))
}