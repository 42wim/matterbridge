// Copyright 2015 Ola Holmström. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package melody implements a framework for dealing with WebSockets.
//
// Example
//
// A broadcasting echo server:
//
//  func main() {
//  	m := melody.New()
//  	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
//  		m.HandleRequest(w, r)
//  	})
//  	m.HandleMessage(func(s *melody.Session, msg []byte) {
//  		m.Broadcast(msg)
//  	})
//  	http.ListenAndServe(":5000", nil)
//  }

package melody
