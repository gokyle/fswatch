package fswatch

import (
        "fmt"
        "os"
)

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


