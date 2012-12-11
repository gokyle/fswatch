package fswatch

import (
        "fmt"
	"os"
	"time"
)

const (
	NONE = iota
	CREATED
	DELETED
	MODIFIED
        PERM
	NOEXIST
	NOPERM
	INVALID
)

var NotificationBufLen = 16

var watch_delay time.Duration

func init() {
	del, err := time.ParseDuration("100ms")
	if err != nil {
		panic("couldn't set up fswatch: " + err.Error())
	}
	watch_delay = del
}

type watchItem struct {
	Path      string
	StatInfo  os.FileInfo
	LastEvent int
}

func watchPath(path string) (wi *watchItem) {
	wi = new(watchItem)
	wi.Path = path
	wi.LastEvent = NONE

	fi, err := os.Stat(path)
	if err == nil {
		wi.StatInfo = fi
	} else if os.IsNotExist(err) {
		wi.LastEvent = NOEXIST
	} else if os.IsPermission(err) {
		wi.LastEvent = NOPERM
	} else {
		wi.LastEvent = INVALID
	}
	return
}

func (wi *watchItem) Update() bool {
	fi, err := os.Stat(wi.Path)
	if err != nil {
                if !os.IsNotExist(err) {
                        fmt.Printf("[-] stat err: %+v\n", err)
                }
		if os.IsNotExist(err) {
			if wi.LastEvent == NOEXIST {
				return false
			} else if wi.LastEvent == DELETED {
				wi.LastEvent = NOEXIST
				return false
			} else {
				wi.LastEvent = DELETED
				return true
			}
		} else if os.IsPermission(err) {
                        fmt.Println("[-] perm event")
			if wi.LastEvent == NOPERM {
                                fmt.Println("[-] already know about bad perms")
				return false
			} else {
                                fmt.Println("[-] perms were changed")
				wi.LastEvent = NOPERM
				return true
			}
		} else {
			wi.LastEvent = INVALID
			return false
		}
	}

        if wi.LastEvent == NOEXIST {
                wi.LastEvent = CREATED
                wi.StatInfo = fi
                return true
        } else if fi.ModTime().After(wi.StatInfo.ModTime()) {
                wi.StatInfo = fi
		switch wi.LastEvent {
		case NONE, CREATED, NOPERM, INVALID:
			wi.LastEvent = MODIFIED
		case DELETED, NOEXIST:
			wi.LastEvent = CREATED
		}
                return true
	} else if fi.Mode() != wi.StatInfo.Mode() {
                wi.LastEvent = PERM
                wi.StatInfo = fi
                return true
        }
	return false
}

type Notification struct {
	Path  string
	Event int
}

func (wi *watchItem) Notification() *Notification {
	return &Notification{wi.Path, wi.LastEvent}
}

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
