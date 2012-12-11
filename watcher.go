package fswatch

import (
        "fmt"
        "time"
)

type Watcher struct {
	paths       map[string]*watchItem
	notify_chan chan *Notification
	auto_watch  bool
}

func newWatcher(dir_notify bool, paths ...string) (w *Watcher) {
	if len(paths) == 0 {
		return
	}
	w = new(Watcher)
	w.auto_watch = dir_notify
	w.paths = make(map[string]*watchItem, 0)

	for _, path := range paths {
		w.paths[path] = watchPath(path)
	}
	return
}

func Watch(paths ...string) *Watcher {
	return newWatcher(false, paths...)
}

func WatchDir(paths ...string) *Watcher {
	return newWatcher(true, paths...)
}

func (w *Watcher) watch(sndch chan<- *Notification) {
        defer func() {
                x := recover()
                fmt.Printf("[+] recover: %+v\n", x)
        }()
        for {
                <-time.After(watch_delay)
                for _, wi := range w.paths {
                        if wi.Update() && w.shouldNotify(wi) {
                                //fmt.Printf("notification: %+v->%+v\n", wi, wi.Notification())
                                sndch<- wi.Notification()
                        }
                }
        }
}

func (w *Watcher) shouldNotify(wi *watchItem) bool {
        if w.auto_watch && wi.StatInfo.IsDir() {
                //fmt.Println("[+] autowatch trigger")
                go w.addPaths(wi)
                return false
        }
        return true
}

func (w *Watcher) addPaths(wi *watchItem) {
        return
}

func (w *Watcher) Start() <-chan *Notification {
        if w.notify_chan != nil {
                return w.notify_chan
        }
        w.notify_chan = make(chan *Notification, NotificationBufLen)
        go w.watch(w.notify_chan)
        return w.notify_chan
}

func (w *Watcher) Stop() {
        if w.notify_chan != nil {
                close(w.notify_chan)
        }
        if w.paths != nil {
                for p, _ := range w.paths {
                        delete(w.paths, p)
                }
        }
}
