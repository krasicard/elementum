package broadcast

import "github.com/elgatito/elementum/util/event"

var (
	// Closer is a global shutdown closer.
	Closer event.Event
)
