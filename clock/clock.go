package clock

import "time"

// Clock is an interface that provides the current time.
type Clock interface {
	Now() time.Time
}
