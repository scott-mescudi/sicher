package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
	"runtime"
)

func main() {
	dir := "folder2"
	os.Mkdir(dir, 0200)
	var wg sync.WaitGroup
	files := []string{
		"diamond.exe",
		"diamond2.exe",
		"diamond3.exe",
		"diamond4.exe",
		"diamond5.exe",
		"main.txt",
	}

	start := time.Now()
	errCh := make(chan error)

	for _, i := range files {
		fp := filepath.Join(dir, i)
		wg.Add(1)

		go func(i,fp string) {
			defer wg.Done()
			if err := copyFile(i, fp, 104_857_600); err != nil{
				errCh <-  fmt.Errorf("failed to copy %s: %v", i, err)
			}
		}(i,fp)
	}

	go func() {
		wg.Wait()
		close(errCh) 
	}()

	for err := range errCh{
		if err != nil{
			fmt.Println(err)
		}
	}

	
	elapsed := time.Since(start)
	printMemUsage()
	fmt.Printf("Time for test1: %v\n",elapsed)
	
}

func copyFile(srcFilePath, dstFilePath string, chunkSize int) (error){
	file, err := os.Open(srcFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	
	f, err := os.Create(dstFilePath)
	if err != nil{
		return err
	}
	defer f.Close()

	buf := make([]byte, chunkSize)


	for {
		bytesRead, err  := file.Read(buf)
		if err != nil{
			if err != io.EOF{
				return err
			}
			break
		}

		_, err =  f.Write(buf[:bytesRead])
		if err != nil{
			return err
		}

		if bytesRead < chunkSize {
			break
		}
	}

	return nil
}


func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) float64 {
	return float64(b) / 1024 / 1024
}