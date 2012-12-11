package main

import (
	"fmt"
	"github.com/gokyle/fswatch"
	"os"
	"time"
)

func init() {
	if len(os.Args) == 1 {
		fmt.Println("[+] not watching anything, exiting.")
		os.Exit(1)
	}
}

func main() {
	paths := os.Args[1:]
	w := fswatch.Watch(paths...)
	fmt.Println("listening...")

	l := w.Start()
	go func() {
		for {
			n, ok := <-l
			if !ok {
				return
			}
			var status_text string
			switch n.Event {
			case fswatch.CREATED:
				status_text = "was created"
			case fswatch.DELETED:
				status_text = "was deleted"
			case fswatch.MODIFIED:
				status_text = "was modified"
			case fswatch.PERM:
				status_text = "permissions changed"
			case fswatch.NOEXIST:
				status_text = "doesn't exist"
			case fswatch.NOPERM:
				status_text = "has invalid permissions"
			case fswatch.INVALID:
				status_text = "is invalid"
			}
			fmt.Printf("[+] %s %s\n", n.Path, status_text)
		}
	}()
	time.Sleep(60 * time.Second)
	fmt.Println("[+] stopping...")
	w.Stop()
	time.Sleep(5 * time.Second)
}
