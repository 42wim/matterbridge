/*
This package allows you to automate actions on Valve's Steam network. It is a Go port of SteamKit.

To login, you'll have to create a new Client first. Then connect to the Steam network
and wait for a ConnectedCallback. Then you may call the Login method in the Auth module
with your login information. This is covered in more detail in the method's documentation. After you've
received the LoggedOnEvent, you should set your persona state to online to receive friend lists etc.

Example code

You can also find a running example in the `gsbot` package.

	package main

	import (
		"io/ioutil"
		"log"

		"github.com/Philipp15b/go-steam"
		"github.com/Philipp15b/go-steam/protocol/steamlang"
	)

	func main() {
		myLoginInfo := new(steam.LogOnDetails)
		myLoginInfo.Username = "Your username"
		myLoginInfo.Password = "Your password"

		client := steam.NewClient()
		client.Connect()
		for event := range client.Events() {
			switch e := event.(type) {
			case *steam.ConnectedEvent:
				client.Auth.LogOn(myLoginInfo)
			case *steam.MachineAuthUpdateEvent:
				ioutil.WriteFile("sentry", e.Hash, 0666)
			case *steam.LoggedOnEvent:
				client.Social.SetPersonaState(steamlang.EPersonaState_Online)
			case steam.FatalErrorEvent:
				log.Print(e)
			case error:
				log.Print(e)
			}
		}
	}


Events

go-steam emits events that can be read via Client.Events(). Although the channel has the type interface{},
only types from this package ending with "Event" and errors will be emitted.

*/
package steam
