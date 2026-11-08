package templates

import "Domicile/internal/assets"

// AssetVersion is appended to static asset URLs as ?v= so browsers fetch a
// fresh copy after every app (re)start instead of reusing a cached file.
var AssetVersion = assets.Version
