package src
import(
	"os"
	"path/filepath"
	"sync"
)

func Clean(srcDir, dstDir string, wg *sync.WaitGroup){
	var srcFiles = make(map[string]bool)
	var dstFiles []string
	
	wg.Add(2)
	go func(){
		defer wg.Done()
		filepath.WalkDir(dstDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if path != dstDir{
				dstFiles = append(dstFiles, path)
			}
			
			return nil
		})
	}()

	go func(){
		defer wg.Done()
		filepath.WalkDir(srcDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if path != dstDir{
				srcFiles[filepath.Base(path)] = true
			}

			return nil
		})
	}()
	wg.Wait()

	for _, v := range dstFiles{
		s := filepath.Base(v)
		if _, ok := srcFiles[s]; !ok{
			os.RemoveAll(v)
		}
	}
	
}