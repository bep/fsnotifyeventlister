package golibtemplate

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"

	"github.com/fsnotify/fsnotify"
)

func TestListEvents(t *testing.T) {
	c := qt.New(t)
	fmt.Println("GOOS is", runtime.GOOS)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	quit := make(chan struct{})

	go func() {
		for {
			select {
			case <-quit:
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				fmt.Println(event)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	tempDir := t.TempDir()
	c.Assert(watcher.Add(tempDir), qt.IsNil)

	do := func(what string, f func() error) {
		fmt.Println(what)
		c.Assert(f(), qt.IsNil, qt.Commentf("while %s", what))
		time.Sleep(1000 * time.Millisecond)
	}

	do("Create a regular file", func() error { return os.WriteFile(filepath.Join(tempDir, "regularfile.txt"), []byte("hello"), 0o644) })
	do("Write to a regular file", func() error {
		return os.WriteFile(filepath.Join(tempDir, "regularfile.txt"), []byte("hello world"), 0o644)
	})
	subDir1 := filepath.Join(tempDir, "subDir1")
	do("Create a sub directory", func() error { return os.Mkdir(subDir1, 0o755) })
	watcher.Add(subDir1)
	subDir2 := filepath.Join(tempDir, "subDir2")
	do("Create another sub directory", func() error { return os.Mkdir(subDir2, 0o755) })
	watcher.Add(subDir2)
	do("Write to a regular file in sub directory", func() error {
		return os.WriteFile(filepath.Join(subDir1, "regularfile.txt"), []byte("hello world"), 0o644)
	})
	do("Write to a regular file in the other sub directory", func() error {
		return os.WriteFile(filepath.Join(subDir2, "regularfile.txt"), []byte("hello world"), 0o644)
	})

	do("Rename regular file", func() error {
		return os.Rename(filepath.Join(tempDir, "regularfile.txt"), filepath.Join(tempDir, "regularfile2.txt"))
	})
	do("Remove regular file", func() error {
		return os.Remove(filepath.Join(tempDir, "regularfile2.txt"))
	})
	do("Rename sub directory", func() error {
		return os.Rename(subDir2, filepath.Join(tempDir, "subDir3"))
	})

	do("Remove sub directory", func() error {
		return os.RemoveAll(subDir1)
	})

	time.Sleep(2 * time.Second)
	close(quit)
}
