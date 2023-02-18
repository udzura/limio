package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/udzura/limio"
)

const burstLimit = 1000 * 1000 * 1000

var limitBPS float64 = 1024 * 1024 * 500 // 500MB

func cleanup() {
	fs, _ := filepath.Glob("./*.dest")
	for _, f := range fs {
		os.Remove(f)
	}
}

func main() {
	CopyFiles(5)
	cleanup()

	CopyFilesLimit(5)
	cleanup()
}

func CopyFiles(n int) {
	start := time.Now()
	fds := make([]io.ReadCloser, 0)
	dests := make([]io.WriteCloser, 0)

	for i := 0; i < n; i++ {
		f, err := os.Open(fmt.Sprintf("./big%d.image", i))
		if err != nil {
			panic(err)
		}
		fds = append(fds, f)
		defer f.Close()

		d, err := os.Create(fmt.Sprintf("./dest%d.dest", i))
		if err != nil {
			panic(err)
		}
		dests = append(dests, d)
		defer d.Close()
	}

	var wait sync.WaitGroup
	wait.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			n, err := io.Copy(dests[i], fds[i])
			if err != nil {
				panic(err)
			}
			dur := time.Since(start)
			log.Printf("%d bytes copied in %s (%0.2f b/s)\n", n, dur, float64(n)/float64(dur)*float64(time.Second))
			wait.Done()
		}(i)
	}

	wait.Wait()
}

func CopyFilesLimit(n int) {
	start := time.Now()
	fds := make([]io.ReadCloser, 0)
	dests := make([]io.WriteCloser, 0)

	for i := 0; i < n; i++ {
		f, err := os.Open(fmt.Sprintf("./big%d.image", i))
		if err != nil {
			panic(err)
		}
		fds = append(fds, f)
		defer f.Close()

		d, err := os.Create(fmt.Sprintf("./dest%d.dest", i))
		if err != nil {
			panic(err)
		}
		dests = append(dests, d)
		defer d.Close()
	}

	pool := limio.NewPool(limitBPS)
	var wait sync.WaitGroup
	wait.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			w := pool.GetWriteCloser(dests[i])
			n, err := io.Copy(w, fds[i])
			if err != nil {
				panic(err)
			}
			dur := time.Since(start)
			log.Printf("%d bytes copied in %s (%0.2f b/s)\n", n, dur, float64(n)/float64(dur)*float64(time.Second))
			wait.Done()
		}(i)
	}

	wait.Wait()
}
