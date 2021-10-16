// Package ddp implements the MeteorJS DDP protocol over websockets. Fallback
// to long polling is NOT supported (and is not planned on ever being supported
// by this library). We will try to model the library after `net/http` - right
// now the library is bare bones and doesn't provide the plug-ability of http.
// However, that's the goal for the package eventually.
package ddp
