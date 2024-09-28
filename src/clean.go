package src

import (
	"os"
	"path/filepath"
	"strings"
)

func Clean(srcDir, dstDir string){
	var dstFiles []string
	var srcFiles = make(map[string]bool)
	
	filepath.WalkDir(dstDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path != dstDir{
			dstFiles = append(dstFiles, strings.TrimPrefix(path, dstDir))
		}
		
		return nil
	})

	filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path != srcDir{
			srcFiles[strings.TrimPrefix(path, srcDir)] = true
		}
		
		return nil
	})



	for _, v := range dstFiles{
		if _, ok := srcFiles[v]; !ok{
			os.RemoveAll(filepath.Join(dstDir,v))
		}
	}
	
}