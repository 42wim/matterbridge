package kbchat

type Sender struct {
	Uid        string `json:"uid"`
	Username   string `json:"username"`
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
}

type Channel struct {
	Name        string `json:"name"`
	Public      bool   `json:"public"`
	TopicType   string `json:"topic_type"`
	TopicName   string `json:"topic_name"`
	MembersType string `json:"members_type"`
}

type Conversation struct {
	ID      string  `json:"id"`
	Unread  bool    `json:"unread"`
	Channel Channel `json:"channel"`
}

type PaymentHolder struct {
	Payment Payment `json:"notification"`
}

type Payment struct {
	TxID              string `json:"txID"`
	StatusDescription string `json:"statusDescription"`
	FromAccountID     string `json:"fromAccountID"`
	FromUsername      string `json:"fromUsername"`
	ToAccountID       string `json:"toAccountID"`
	ToUsername        string `json:"toUsername"`
	AmountDescription string `json:"amountDescription"`
	WorthAtSendTime   string `json:"worthAtSendTime"`
	ExternalTxURL     string `json:"externalTxURL"`
}

type Result struct {
	Convs []Conversation `json:"conversations"`
}

type Inbox struct {
	Result Result `json:"result"`
}

type ChannelsList struct {
	Result Result `json:"result"`
}

type MsgPaymentDetails struct {
	ResultType int    `json:"resultTyp"` // 0 good. 1 error
	PaymentID  string `json:"sent"`
}

type MsgPayment struct {
	Username    string            `json:"username"`
	PaymentText string            `json:"paymentText"`
	Details     MsgPaymentDetails `json:"result"`
}

type Text struct {
	Body     string       `json:"body"`
	Payments []MsgPayment `json:"payments"`
	ReplyTo  int          `json:"replyTo"`
}

type Content struct {
	Type string `json:"type"`
	Text Text   `json:"text"`
}

type Message struct {
	Content        Content `json:"content"`
	Sender         Sender  `json:"sender"`
	Channel        Channel `json:"channel"`
	ConversationID string  `json:"conversation_id"`
	MsgID          int     `json:"id"`
}

type SendResult struct {
	MsgID int `json:"id"`
}

type SendResponse struct {
	Result SendResult `json:"result"`
}

type TypeHolder struct {
	Type string `json:"type"`
}

type MessageHolder struct {
	Msg    Message `json:"msg"`
	Source string  `json:"source"`
}

type ThreadResult struct {
	Messages []MessageHolder `json:"messages"`
}

type Thread struct {
	Result ThreadResult `json:"result"`
}

type CommandExtendedDescription struct {
	Title       string `json:"title"`
	DesktopBody string `json:"desktop_body"`
	MobileBody  string `json:"mobile_body"`
}

type Command struct {
	Name                string                      `json:"name"`
	Description         string                      `json:"description"`
	Usage               string                      `json:"usage"`
	ExtendedDescription *CommandExtendedDescription `json:"extended_description,omitempty"`
}

type CommandsAdvertisement struct {
	Typ      string `json:"type"`
	Commands []Command
	TeamName string `json:"team_name,omitempty"`
}

type Advertisement struct {
	Alias          string `json:"alias,omitempty"`
	Advertisements []CommandsAdvertisement
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type JoinChannel struct {
	Error  Error             `json:"error"`
	Result JoinChannelResult `json:"result"`
}

type JoinChannelResult struct {
	RateLimit []RateLimit `json:"ratelimits"`
}

type LeaveChannel struct {
	Error  Error              `json:"error"`
	Result LeaveChannelResult `json:"result"`
}

type LeaveChannelResult struct {
	RateLimit []RateLimit `json:"ratelimits"`
}

type RateLimit struct {
	Tank     string `json:"tank"`
	Capacity int    `json:"capacity"`
	Reset    int    `json:"reset"`
	Gas      int    `json:"gas"`
}
