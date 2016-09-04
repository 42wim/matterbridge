package main

import "github.com/pythonandchips/wray"
import "fmt"

func main() {
  wray.RegisterTransports([]wray.Transport{ &gofaye.HttpTransport{} })
  client := wray.NewFayeClient("http://localhost:5000/faye")

  params := map[string]interface{}{"hello": "from golang"}
  fmt.Println("sending")
  client.Publish("/foo", params)
}


