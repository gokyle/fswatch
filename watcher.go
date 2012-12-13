package fswatch

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Type Watcher represents a file system watcher. It should be initialised
// with NewWatcher or NewAutoWatcher, and started with Watcher.Start().
type Watcher struct {
	paths       map[string]*watchItem
	notify_chan chan *Notification
	add_chan    chan *watchItem
	auto_watch  bool
}

// newWatcher is the internal function for properly setting up a new watcher.
func newWatcher(dir_notify bool, paths ...string) (w *Watcher) {
	w = new(Watcher)
	w.auto_watch = dir_notify
	w.paths = make(map[string]*watchItem, 0)

	if dir_notify {
		fmt.Println("[-] calling syncAddPaths...")
		w.syncAddPaths(paths...)
	} else {
		for _, path := range paths {
			w.paths[path] = watchPath(path)
		}
	}
	return
}

// NewWatcher initialises a new Watcher with an initial set of paths. It
// does not start listening, and this Watcher will not automatically add
// files created under any directories it is watching.
func NewWatcher(paths ...string) *Watcher {
	return newWatcher(false, paths...)
}

// NewAutoWatcher initialises a new Watcher with an initial set of paths.
// It behaves the same as NewWatcher, except it will automatically add
// files created in directories it is watching, including adding any
// subdirectories.
func NewAutoWatcher(paths ...string) *Watcher {
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
				sndch <- wi.Notification()
			}
		}
	}
}

func (w *Watcher) shouldNotify(wi *watchItem) bool {
	if w.auto_watch && wi.StatInfo.IsDir() {
		go w.addPaths(wi)
		return false
	}
	return true
}

func (w *Watcher) addPaths(wi *watchItem) {
	walker := getWalker(w, wi.Path, w.add_chan)
	go filepath.Walk(wi.Path, walker)
}

// Start begins watching the files, sending notifications when files change.
// It returns a channel that notifications are sent on.
func (w *Watcher) Start() <-chan *Notification {
	if w.notify_chan != nil {
		return w.notify_chan
	}
	if w.auto_watch {
		w.add_chan = make(chan *watchItem, NotificationBufLen)
		go w.watchItemListener()
	}
	w.notify_chan = make(chan *Notification, NotificationBufLen)
	go w.watch(w.notify_chan)
	return w.notify_chan
}

// Stop listening for changes to the files.
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

func (w *Watcher) syncAddPaths(paths ...string) {
	fmt.Println("[-] syncAddPaths")
	for _, path := range paths {
		wi := watchPath(path)
		if wi == nil {
			continue
		} else if _, watching := w.paths[wi.Path]; watching {
			continue
		}
		w.paths[wi.Path] = wi
		if wi.StatInfo.IsDir() {
			w.syncAddDir(wi)
		}
	}
}

func (w *Watcher) watchItemListener() {
	defer func() {
		recover()
	}()
	for {
		wi := <-w.add_chan
		if wi == nil {
			continue
		} else if _, watching := w.paths[wi.Path]; watching {
			continue
		}
		w.paths[wi.Path] = wi
	}
}

func getWalker(w *Watcher, root string, addch chan<- *watchItem) func(string, os.FileInfo, error) error {
	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		wi := watchPath(path)
		if wi == nil {
			return nil
		} else if _, watching := w.paths[wi.Path]; !watching {
			wi.LastEvent = CREATED
			w.notify_chan <- wi.Notification()
			addch <- wi
			if !wi.StatInfo.IsDir() {
				return nil
			}
			w.addPaths(wi)
		}
		return nil
	}
	return walker
}

func (w *Watcher) syncAddDir(wi *watchItem) {
	walker := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == wi.Path {
			return nil
		}
		new_wi := watchPath(path)
		if new_wi != nil {
			fmt.Println("[sync] add ", wi.Path)
			w.paths[path] = new_wi
			if !new_wi.StatInfo.IsDir() {
				return nil
			}
			if _, watching := w.paths[new_wi.Path]; !watching {
				w.syncAddDir(new_wi)
			}
		}
		return nil
	}
	filepath.Walk(wi.Path, walker)
}

func (w *Watcher) Watching() (paths []string) {
	paths = make([]string, 0)
	for path, _ := range w.paths {
		paths = append(paths, path)
	}
	return
}
