package pkg

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func Clean(srcDir, dstDir string) {
	var dstFiles []string
	var srcFiles = make(map[string]bool)
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		filepath.WalkDir(dstDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if path != dstDir {
				dstFiles = append(dstFiles, strings.TrimPrefix(path, dstDir))
			}

			return nil
		})
	}()
	
	go func() {
		defer wg.Done()
		filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if path != srcDir {
				srcFiles[strings.TrimPrefix(path, srcDir)] = true
			}

			return nil
		})
	}()

	wg.Wait()
	for _, v := range dstFiles {
		if _, ok := srcFiles[v]; !ok {
			os.RemoveAll(filepath.Join(dstDir, v))
		}
	}

}
