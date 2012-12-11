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
        w := fswatch.New(paths...)
        fmt.Println("listening...")

        l := w.Watch()
        go func() {
                for {
                        n, ok := <-l
                        if !ok {
                                return
                        }
                        fmt.Printf("[+] %s %s\n", n.Path, n.EventText())
                }
        }()
        time.Sleep(60 * time.Second)
        fmt.Println("[+] stopping...")
        w.Stop()
        time.Sleep(5 * time.Second)
}
