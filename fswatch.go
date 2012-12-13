package fswatch

import (
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
