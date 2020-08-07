bcast package for Go
====================

Broadcasting on a set of channels in Go. Go channels offer different usage patterns but not ready to use broadcast pattern.
This library solves the problem in direct way. Each routine keeps member structure with own input channel and single for all
members output channel. Central dispatcher accepts broadcasts and resend them to all members.

Usage [![Go Walker](http://img.shields.io/badge/docs-API-brightgreen.svg?style=flat)](http://gowalker.org/github.com/NimbleIndustry/bcast)
-----

Firstly import package and create broadcast group. You may create any number of groups for different broadcasts:

			import (
				"github.com/grafov/bcast"
			)

			group := bcast.NewGroup() // create broadcast group
			go group.Broadcast(0) // accepts messages and broadcast it to all members

You may listen broadcasts limited time:

			bcast.Broadcast(2 * time.Minute) // if message not arrived during 2 min. function exits

Now join to the group from different goroutines:

			member1 := group.Join() // joined member1 from one routine

Either member may send message which received by all other members of the group:

			member1.Send("test message") // send message to all members

Also you may send message to group from nonmember of a group:

			group.Send("test message")

Method `Send` accepts `interface{}` type so any values may be broadcasted.

			member2 := group.Join() // joined member2 form another routine
			val := member1.Recv() // broadcasted value received

Another way to receive broadcasted messages is listen input channel of the member.

			val := <-*member1.In // each member keeps pointer to its own input channel

It may be convenient for example when `select` used.

See more examples in a test suit `bcast_test.go`.

Install
-------

`go get github.com/grafov/bcast`

The library doesn't require external packages for build. The next
package required if you want to run unit tests:

`gopkg.in/fatih/set.v0`

License
-------

Library licensed under BSD 3-clause license. See LICENSE.

Project status [![Build Status](https://img.shields.io/travis/grafov/bcast/master.svg?style=flat)](https://travis-ci.org/grafov/bcast)
--------------

WIP again. There is bug found (see #12) and some possible improvements are waiting for review (#9).

API is stable. No major changes planned, maybe small improvements.
