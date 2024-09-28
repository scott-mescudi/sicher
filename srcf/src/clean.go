package src

import (
	"os"
	"path/filepath"
)

func Clean(srcFiles map[string]bool, dstDir string){
	var dstFiles []string
	filepath.WalkDir(dstDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if path != dstDir{
			dstFiles = append(dstFiles, path)
		}
		
		return nil
	})

	for _, v := range dstFiles{
		s := filepath.Base(v)
		if _, ok := srcFiles[s]; !ok{
			os.RemoveAll(v)
		}
	}
	
}