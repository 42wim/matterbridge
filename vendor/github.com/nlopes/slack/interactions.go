package slack

// InteractionType type of interactions
type InteractionType string

// Types of interactions that can be received.
const (
	InteractionTypeDialogCancellation = InteractionType("dialog_cancellation")
	InteractionTypeDialogSubmission   = InteractionType("dialog_submission")
	InteractionTypeDialogSuggestion   = InteractionType("dialog_suggestion")
	InteractionTypeInteractionMessage = InteractionType("interactive_message")
	InteractionTypeMessageAction      = InteractionType("message_action")
)

// InteractionCallback is sent from slack when a user interactions with a button or dialog.
type InteractionCallback struct {
	Type            InteractionType `json:"type"`
	Token           string          `json:"token"`
	CallbackID      string          `json:"callback_id"`
	ResponseURL     string          `json:"response_url"`
	TriggerID       string          `json:"trigger_id"`
	ActionTs        string          `json:"action_ts"`
	Team            Team            `json:"team"`
	Channel         Channel         `json:"channel"`
	User            User            `json:"user"`
	OriginalMessage Message         `json:"original_message"`
	Message         Message         `json:"message"`
	Name            string          `json:"name"`
	Value           string          `json:"value"`
	ActionCallback
	DialogSubmissionCallback
}
