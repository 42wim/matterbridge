package wray

type Response struct {
  id string
  channel string
  successful bool
  clientId string
  supportedConnectionTypes []string
  messages []Message
  error error
}

type Message struct {
  Channel string
  Id string
  Data map[string]interface{}
}

func newResponse(data []interface{}) Response {
  headerData := data[0].(map[string]interface{})
  messagesData := data[1.:]
  messages := parseMessages(messagesData)
  var id string
  if headerData["id"] != nil {
    id = headerData["id"].(string)
  }
  supportedConnectionTypes := []string{}
  if headerData["supportedConnectionTypes"] != nil {
    d := headerData["supportedConnectionTypes"].([]interface{})
    for _, sct := range(d) {
      supportedConnectionTypes = append(supportedConnectionTypes, sct.(string))
    }
  }
  var clientId string
  if headerData["clientId"] != nil {
    clientId = headerData["clientId"].(string)
  }
  return Response{id: id,
                  clientId: clientId,
                  channel: headerData["channel"].(string),
                  successful: headerData["successful"].(bool),
                  messages: messages,
                  supportedConnectionTypes: supportedConnectionTypes}
}

func parseMessages(data []interface{}) []Message {
  messages := []Message{}
  for _, messageData := range(data) {
    m := messageData.(map[string]interface{})
    var id string
    if m["id"] != nil {
      id = m["id"].(string)
    }
    message := Message{Channel: m["channel"].(string),
                       Id: id,
                       Data: m["data"].(map[string]interface{})}
    messages = append(messages, message)
  }
  return messages
}

