//go:build js && wasm
// +build js,wasm

package webrtc

import "syscall/js"

type TrackRemote struct {
	// Pointer to the underlying JavaScript TrackRemote object.
	underlying js.Value
}
