# gitter
Gitter API in Go
https://developer.gitter.im

#### Install

`go get github.com/sromku/go-gitter`

- [Initialize](#initialize)
- [Users](#users)
- [Rooms](#rooms)
- [Messages](#messages)
- [Stream](#stream)
- [Faye (Experimental)](#faye-experimental)
- [Debug](#debug)
- [App Engine](#app-engine)

##### Initialize
``` Go
api := gitter.New("YOUR_ACCESS_TOKEN")
```

##### Users

- Get current user

	``` Go
	user, err := api.GetUser()
	```

##### Rooms

- Get all rooms
	``` Go
	rooms, err := api.GetRooms()
	```

- Get room by id
	``` Go
	room, err := api.GetRoom("roomID")
	```

- Get rooms of some user
	``` Go
	rooms, err := api.GetRooms("userID")
	```

- Join room
	``` Go
	room, err := api.JoinRoom("roomID", "userID")
	```
	
- Leave room
	``` Go
	room, err := api.LeaveRoom("roomID", "userID")
	```

- Get room id
	``` Go
	id, err := api.GetRoomId("room/uri")
	```

- Search gitter rooms
	``` Go
	rooms, err := api.SearchRooms("search/string")
	```
##### Messages

- Get messages of room
	``` Go
	messages, err := api.GetMessages("roomID", nil)
	```

- Get one message
	``` Go
	message, err := api.GetMessage("roomID", "messageID")
	```

- Send message
	``` Go
	err := api.SendMessage("roomID", "free chat text")
	```

##### Stream

Create stream to the room and start listening to incoming messages

``` Go
stream := api.Stream(room.Id)
go api.Listen(stream)

for {
    event := <-stream.Event
    switch ev := event.Data.(type) {
    case *gitter.MessageReceived:
        fmt.Println(ev.Message.From.Username + ": " + ev.Message.Text)
    case *gitter.GitterConnectionClosed:
        // connection was closed
    }
}
```

Close stream connection

``` Go
stream.Close()
```

##### Faye (Experimental)

``` Go
faye := api.Faye(room.ID)
go faye.Listen()

for {
    event := <-faye.Event
    switch ev := event.Data.(type) {
    case *gitter.MessageReceived:
        fmt.Println(ev.Message.From.Username + ": " + ev.Message.Text)
    case *gitter.GitterConnectionClosed: //this one is never called in Faye
        // connection was closed
    }
}
```

##### Debug

You can print the internal errors by enabling debug to true

``` Go
api.SetDebug(true, nil)
```

You can also define your own `io.Writer` in case you want to persist the logs somewhere.
For example keeping the errors on file

``` Go
logFile, err := os.Create("gitter.log")
api.SetDebug(true, logFile)
```

##### App Engine

Initialize app engine client and continue as usual

``` Go
c := appengine.NewContext(r)
client := urlfetch.Client(c)

api := gitter.New("YOUR_ACCESS_TOKEN")
api.SetClient(client)
```

[Documentation](https://godoc.org/github.com/sromku/go-gitter)
