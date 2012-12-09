package fswatch

import (
	"os"
	"time"
)

const (
	CREATED = iota
	DELETED
	MODIFIED
	NOPERM
	INVALID
)

type watchItem struct {
	Path      string
	last_mod  time.Time
	Exists    bool
	LastEvent int
}

type Notification struct {
	Path  string
	Event int
}

func getLastMod(path string) (mt time.Time, err error) {
	fi, err := os.Stat(path)
	if err != nil {
		return
	}
	return fi.ModTime(), nil
}

func (wi *watchItem) Update() (updated bool) {
	mt, err := getLastMod(wi.Path)
	switch {
	case os.IsNotExist(err):
		if wi.LastEvent != NOPERM {
			wi.LastEvent = NOPERM
			updated = true
		}
	case os.IsPermission(err):
		if wi.LastEvent != NOPERM {
			wi.LastEvent = NOPERM
			updated = true
		}
	default:
		if wi.LastEvent != INVALID {
			wi.LastEvent = INVALID
			updated = true
		}
	}
	if updated {
		return
	}

	if mt != wi.last_mod {
		wi.last_mod = mt
		updated = true
	}
	return
}

func (wi *watchItem) Notification() *Notification {
	return &Notification{wi.Path, wi.LastEvent}
}

type Watcher struct {
	items  map[string]*watchItem
	notify chan *Notification
	delay  time.Duration
}
