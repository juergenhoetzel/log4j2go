package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"syscall"

	"flag"
	"github.com/juergenhoetzel/log4j2go/internal/log4j"
	"sync"
	"runtime"
)

func main() {
	sameFs := flag.Bool("samefs", false, "dont search in mountpoints")
	flag.Parse()

	numJobs := runtime.NumCPU()
	jobs := make(chan string, numJobs)
	var wg sync.WaitGroup
	for w := 1; w <= numJobs; w++ {
		wg.Add(1)
		go func () {
			for s := range jobs {
				log4j.CheckFile(s, s);
			}
			wg.Done()
		}()
	}

	for _, s := range flag.Args() {

		stat := syscall.Statfs_t{}
		var limitFs *syscall.Statfs_t
		if err := syscall.Statfs(s, &stat); err != nil {
			// FIXME: UNIX-only
			log.Printf("statfs %q failed: %v", s, err)
		} else {
			limitFs = &stat
		}

		file, err := os.Open(s)
		if err != nil {
			log.Printf("Failed to open %s: %v\n", s, err)
			continue
		}
		defer file.Close()
		fileInfo, err := file.Stat()
		if err != nil {
			log.Printf("Failed to stat %s: %v\n", s, err)
			continue
		}
		if fileInfo.IsDir() {
			err = filepath.Walk(s, func(path string, info fs.FileInfo, err error) error {
				if err != nil {
					log.Printf("Ignoring failure accessing a path %q: %v\n", path, err)
					return nil
				}
				if !info.IsDir() && (strings.HasSuffix(info.Name(), ".jar") || strings.HasSuffix(info.Name(), ".war") || strings.HasSuffix(info.Name(), ".ear")) {
					jobs <- path
				}

				if info.IsDir() && *sameFs && limitFs != nil {
					stat := syscall.Statfs_t{}
					syscall.Statfs(path, &stat)
					if stat.Fsid != limitFs.Fsid {
						log.Printf("Ignoring mountpoint %q", path)
						return filepath.SkipDir
					}
				}
				return nil
			})
			if err != nil {
				log.Printf("error walking the path %q: %v\n", s, err)
				continue
			}
		} else {
			jobs <- s
		}
	}
	close(jobs)
	wg.Wait()
}
