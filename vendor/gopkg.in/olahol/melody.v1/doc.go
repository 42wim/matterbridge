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
//  	r := gin.Default()
//  	m := melody.New()
//  	r.GET("/ws", func(c *gin.Context) {
//  		m.HandleRequest(c.Writer, c.Request)
//  	})
//  	m.HandleMessage(func(s *melody.Session, msg []byte) {
//  		m.Broadcast(msg)
//  	})
//  	r.Run(":5000")
//  }
package melody
