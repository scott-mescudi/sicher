package main

import (
	"fmt"

	"runtime"
	"sync"

	"gobackup/src"
)

func main() {
	srcDir := "srcf"
	dstdir := "dstf"
	var wg sync.WaitGroup
	src.Clean(srcDir, dstdir, &wg)
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

