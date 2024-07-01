package socketmode

import (
	"context"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type SocketmodeHandler struct {
	Client *Client

	//lvl 1 - the most generic type of event
	EventMap map[EventType][]SocketmodeHandlerFunc
	//lvl 2 - Manage event by inner type
	InteractionEventMap map[slack.InteractionType][]SocketmodeHandlerFunc
	EventApiMap         map[slackevents.EventsAPIType][]SocketmodeHandlerFunc
	//lvl 3 - the most userfriendly way of managing event
	InteractionBlockActionEventMap map[string]SocketmodeHandlerFunc
	SlashCommandMap                map[string]SocketmodeHandlerFunc

	Default SocketmodeHandlerFunc
}

// Handler have access to the event and socketmode client
type SocketmodeHandlerFunc func(*Event, *Client)

// Middleware accept SocketmodeHandlerFunc, and return SocketmodeHandlerFunc
type SocketmodeMiddlewareFunc func(SocketmodeHandlerFunc) SocketmodeHandlerFunc

// Initialization constructor for SocketmodeHandler
func NewSocketmodeHandler(client *Client) *SocketmodeHandler {
	eventMap := make(map[EventType][]SocketmodeHandlerFunc)
	interactionEventMap := make(map[slack.InteractionType][]SocketmodeHandlerFunc)
	eventApiMap := make(map[slackevents.EventsAPIType][]SocketmodeHandlerFunc)

	interactionBlockActionEventMap := make(map[string]SocketmodeHandlerFunc)
	slackCommandMap := make(map[string]SocketmodeHandlerFunc)

	return &SocketmodeHandler{
		Client:                         client,
		EventMap:                       eventMap,
		EventApiMap:                    eventApiMap,
		InteractionEventMap:            interactionEventMap,
		InteractionBlockActionEventMap: interactionBlockActionEventMap,
		SlashCommandMap:                slackCommandMap,
		Default: func(e *Event, c *Client) {
			c.log.Printf("Unexpected event type received: %v\n", e.Type)
		},
	}
}

// Register a middleware or handler for an Event from socketmode
// This most general entrypoint
func (r *SocketmodeHandler) Handle(et EventType, f SocketmodeHandlerFunc) {
	r.EventMap[et] = append(r.EventMap[et], f)
}

// Register a middleware or handler for an Interaction
// There is several types of interactions, decated functions lets you better handle them
// See
// * HandleInteractionBlockAction
// * (Not Implemented) HandleShortcut
// * (Not Implemented) HandleView
func (r *SocketmodeHandler) HandleInteraction(et slack.InteractionType, f SocketmodeHandlerFunc) {
	r.InteractionEventMap[et] = append(r.InteractionEventMap[et], f)
}

// Register a middleware or handler for a Block Action referenced by its ActionID
func (r *SocketmodeHandler) HandleInteractionBlockAction(actionID string, f SocketmodeHandlerFunc) {
	if actionID == "" {
		panic("invalid command cannot be empty")
	}
	if f == nil {
		panic("invalid handler cannot be nil")
	}
	if _, exist := r.InteractionBlockActionEventMap[actionID]; exist {
		panic("multiple registrations for actionID" + actionID)
	}
	r.InteractionBlockActionEventMap[actionID] = f
}

// Register a middleware or handler for an Event (from slackevents)
func (r *SocketmodeHandler) HandleEvents(et slackevents.EventsAPIType, f SocketmodeHandlerFunc) {
	r.EventApiMap[et] = append(r.EventApiMap[et], f)
}

// Register a middleware or handler for a Slash Command
func (r *SocketmodeHandler) HandleSlashCommand(command string, f SocketmodeHandlerFunc) {
	if command == "" {
		panic("invalid command cannot be empty")
	}
	if f == nil {
		panic("invalid handler cannot be nil")
	}
	if _, exist := r.SlashCommandMap[command]; exist {
		panic("multiple registrations for command" + command)
	}
	r.SlashCommandMap[command] = f
}

// Register a middleware or handler to use as a last resort
func (r *SocketmodeHandler) HandleDefault(f SocketmodeHandlerFunc) {
	r.Default = f
}

// RunSlackEventLoop receives the event via the socket
func (r *SocketmodeHandler) RunEventLoop() error {

	go r.runEventLoop(context.Background())

	return r.Client.Run()
}

func (r *SocketmodeHandler) RunEventLoopContext(ctx context.Context) error {
	go r.runEventLoop(ctx)

	return r.Client.RunContext(ctx)
}

// Call the dispatcher for each incomming event
func (r *SocketmodeHandler) runEventLoop(ctx context.Context) {
	for {
		select {
		case evt, ok := <-r.Client.Events:
			if !ok {
				return
			}

			r.dispatcher(evt)

		case <-ctx.Done():
			return
		}
	}
}

// Dispatch events to the specialized dispatcher
func (r *SocketmodeHandler) dispatcher(evt Event) {
	var ishandled bool

	// Some eventType can be further decomposed
	switch evt.Type {
	case EventTypeInteractive:
		ishandled = r.interactionDispatcher(&evt)
	case EventTypeEventsAPI:
		ishandled = r.eventAPIDispatcher(&evt)
	case EventTypeSlashCommand:
		ishandled = r.slashCommandDispatcher(&evt)
	default:
		ishandled = r.socketmodeDispatcher(&evt)
	}

	if !ishandled {
		go r.Default(&evt, r.Client)
	}
}

// Dispatch socketmode events to the registered middleware
func (r *SocketmodeHandler) socketmodeDispatcher(evt *Event) bool {
	if handlers, ok := r.EventMap[evt.Type]; ok {
		// If we registered an event
		for _, f := range handlers {
			go f(evt, r.Client)
		}

		return true
	}

	return false
}

// Dispatch interactions to the registered middleware
func (r *SocketmodeHandler) interactionDispatcher(evt *Event) bool {
	var ishandled bool = false

	interaction, ok := evt.Data.(slack.InteractionCallback)
	if !ok {
		r.Client.log.Printf("Ignored %+v\n", evt)
		return false
	}

	// Level 1 - socketmode EventType
	ishandled = r.socketmodeDispatcher(evt)

	// Level 2 - interaction EventType
	if handlers, ok := r.InteractionEventMap[interaction.Type]; ok {
		// If we registered an event
		for _, f := range handlers {
			go f(evt, r.Client)
		}

		ishandled = true
	}

	// Level 3 - interaction with actionID
	blockActions := interaction.ActionCallback.BlockActions
	// outmoded approach won`t be implemented
	// attachments_actions := interaction.ActionCallback.AttachmentActions

	for _, action := range blockActions {
		if handler, ok := r.InteractionBlockActionEventMap[action.ActionID]; ok {

			go handler(evt, r.Client)

			ishandled = true
		}
	}
	return ishandled
}

// Dispatch eventAPI events to the registered middleware
func (r *SocketmodeHandler) eventAPIDispatcher(evt *Event) bool {
	var ishandled bool = false
	eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
	if !ok {
		r.Client.log.Printf("Ignored %+v\n", evt)
		return false
	}

	innerEventType := slackevents.EventsAPIType(eventsAPIEvent.InnerEvent.Type)

	// Level 1 - socketmode EventType
	ishandled = r.socketmodeDispatcher(evt)

	// Level 2 - EventAPI EventType
	if handlers, ok := r.EventApiMap[innerEventType]; ok {
		// If we registered an event
		for _, f := range handlers {
			go f(evt, r.Client)
		}

		ishandled = true
	}

	return ishandled
}

// Dispatch SlashCommands events to the registered middleware
func (r *SocketmodeHandler) slashCommandDispatcher(evt *Event) bool {
	var ishandled bool = false
	slashCommandEvent, ok := evt.Data.(slack.SlashCommand)
	if !ok {
		r.Client.log.Printf("Ignored %+v\n", evt)
		return false
	}

	// Level 1 - socketmode EventType
	ishandled = r.socketmodeDispatcher(evt)

	// Level 2 - SlackCommand by name
	if handler, ok := r.SlashCommandMap[slashCommandEvent.Command]; ok {

		go handler(evt, r.Client)

		ishandled = true
	}

	return ishandled

}
