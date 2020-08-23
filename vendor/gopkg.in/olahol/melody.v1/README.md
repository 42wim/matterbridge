# melody

[![Build Status](https://travis-ci.org/olahol/melody.svg)](https://travis-ci.org/olahol/melody)
[![Coverage Status](https://img.shields.io/coveralls/olahol/melody.svg?style=flat)](https://coveralls.io/r/olahol/melody)
[![GoDoc](https://godoc.org/github.com/olahol/melody?status.svg)](https://godoc.org/github.com/olahol/melody)

> :notes: Minimalist websocket framework for Go.

Melody is websocket framework based on [github.com/gorilla/websocket](https://github.com/gorilla/websocket)
that abstracts away the tedious parts of handling websockets. It gets out of
your way so you can write real-time apps. Features include:

* [x] Clear and easy interface similar to `net/http` or Gin.
* [x] A simple way to broadcast to all or selected connected sessions.
* [x] Message buffers making concurrent writing safe.
* [x] Automatic handling of ping/pong and session timeouts.
* [x] Store data on sessions.

## Install

```bash
go get gopkg.in/olahol/melody.v1
```

## [Example: chat](https://github.com/olahol/melody/tree/master/examples/chat)

[![Chat](https://cdn.rawgit.com/olahol/melody/master/examples/chat/demo.gif "Demo")](https://github.com/olahol/melody/tree/master/examples/chat)

Using [Gin](https://github.com/gin-gonic/gin):
```go
package main

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/olahol/melody.v1"
	"net/http"
)

func main() {
	r := gin.Default()
	m := melody.New()

	r.GET("/", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "index.html")
	})

	r.GET("/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		m.Broadcast(msg)
	})

	r.Run(":5000")
}
```

Using [Echo](https://github.com/labstack/echo):
```go
package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
	"gopkg.in/olahol/melody.v1"
	"net/http"
)

func main() {
	e := echo.New()
	m := melody.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		http.ServeFile(c.Response().(*standard.Response).ResponseWriter, c.Request().(*standard.Request).Request, "index.html")
		return nil
	})

	e.GET("/ws", func(c echo.Context) error {
		m.HandleRequest(c.Response().(*standard.Response).ResponseWriter, c.Request().(*standard.Request).Request)
		return nil
	})

	m.HandleMessage(func(s *melody.Session, msg []byte) {
		m.Broadcast(msg)
	})

	e.Run(standard.New(":5000"))
}
```

## [Example: gophers](https://github.com/olahol/melody/tree/master/examples/gophers)

[![Gophers](https://cdn.rawgit.com/olahol/melody/master/examples/gophers/demo.gif "Demo")](https://github.com/olahol/melody/tree/master/examples/gophers)

```go
package main

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/olahol/melody.v1"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type GopherInfo struct {
	ID, X, Y string
}

func main() {
	router := gin.Default()
	mrouter := melody.New()
	gophers := make(map[*melody.Session]*GopherInfo)
	lock := new(sync.Mutex)
	counter := 0

	router.GET("/", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "index.html")
	})

	router.GET("/ws", func(c *gin.Context) {
		mrouter.HandleRequest(c.Writer, c.Request)
	})

	mrouter.HandleConnect(func(s *melody.Session) {
		lock.Lock()
		for _, info := range gophers {
			s.Write([]byte("set " + info.ID + " " + info.X + " " + info.Y))
		}
		gophers[s] = &GopherInfo{strconv.Itoa(counter), "0", "0"}
		s.Write([]byte("iam " + gophers[s].ID))
		counter += 1
		lock.Unlock()
	})

	mrouter.HandleDisconnect(func(s *melody.Session) {
		lock.Lock()
		mrouter.BroadcastOthers([]byte("dis "+gophers[s].ID), s)
		delete(gophers, s)
		lock.Unlock()
	})

	mrouter.HandleMessage(func(s *melody.Session, msg []byte) {
		p := strings.Split(string(msg), " ")
		lock.Lock()
		info := gophers[s]
		if len(p) == 2 {
			info.X = p[0]
			info.Y = p[1]
			mrouter.BroadcastOthers([]byte("set "+info.ID+" "+info.X+" "+info.Y), s)
		}
		lock.Unlock()
	})

	router.Run(":5000")
}
```

### [More examples](https://github.com/olahol/melody/tree/master/examples)

## [Documentation](https://godoc.org/github.com/olahol/melody)

## Contributors

* Ola Holmström (@olahol)
* Shogo Iwano (@shiwano)
* Matt Caldwell (@mattcaldwell)
* Heikki Uljas (@huljas)
* Robbie Trencheny (@robbiet480)
* yangjinecho (@yangjinecho)

## FAQ

If you are getting a `403` when trying  to connect to your websocket you can [change allow all origin hosts](http://godoc.org/github.com/gorilla/websocket#hdr-Origin_Considerations):

```go
m := melody.New()
m.Upgrader.CheckOrigin = func(r *http.Request) bool { return true }
```
