package fswatch

import (
	"os"
	"time"
)

const (
	NONE = iota
	CREATED
	DELETED
	MODIFIED
	NOEXIST
	NOPERM
	INVALID
)

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
			if wi.LastEvent == NOPERM {
				return false
			} else {
				wi.LastEvent = NOPERM
				return true
			}
		} else {
			wi.LastEvent = INVALID
			return false
		}
	}

	if fi.ModTime().After(wi.StatInfo.ModTime()) {
		switch wi.LastEvent {
		case NONE, CREATED, NOPERM, INVALID:
			wi.LastEvent = MODIFIED
		case DELETED, NOEXIST:
			wi.LastEvent = CREATED
		}
	}
	return true
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
