// Package assets holds the cache-busting version shared by all templates.
package assets

import (
	"strconv"
	"time"
)

// Version is appended to asset URLs as ?v= so browsers and CDN caches fetch a
// fresh copy after every app (re)start instead of reusing a stale one.
var Version = strconv.FormatInt(time.Now().Unix(), 10)
