//go:build js && wasm
// +build js,wasm

package webrtc

import (
	"fmt"
	"syscall/js"
)

// RTPReceiver allows an application to inspect the receipt of a TrackRemote
type RTPReceiver struct {
	// Pointer to the underlying JavaScript RTCRTPReceiver object.
	underlying js.Value
}

func (r *RTPReceiver) Track() string {
	fmt.Println("THE UNDERLYING IS ", r.underlying)
	return r.underlying.Get("track").String()
}
