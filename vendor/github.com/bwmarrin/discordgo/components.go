package discordgo

import (
	"encoding/json"
	"fmt"
)

// ComponentType is type of component.
type ComponentType uint

// MessageComponent types.
const (
	ActionsRowComponent ComponentType = 1
	ButtonComponent     ComponentType = 2
	SelectMenuComponent ComponentType = 3
	TextInputComponent  ComponentType = 4
)

// MessageComponent is a base interface for all message components.
type MessageComponent interface {
	json.Marshaler
	Type() ComponentType
}

type unmarshalableMessageComponent struct {
	MessageComponent
}

// UnmarshalJSON is a helper function to unmarshal MessageComponent object.
func (umc *unmarshalableMessageComponent) UnmarshalJSON(src []byte) error {
	var v struct {
		Type ComponentType `json:"type"`
	}
	err := json.Unmarshal(src, &v)
	if err != nil {
		return err
	}

	switch v.Type {
	case ActionsRowComponent:
		umc.MessageComponent = &ActionsRow{}
	case ButtonComponent:
		umc.MessageComponent = &Button{}
	case SelectMenuComponent:
		umc.MessageComponent = &SelectMenu{}
	case TextInputComponent:
		umc.MessageComponent = &TextInput{}
	default:
		return fmt.Errorf("unknown component type: %d", v.Type)
	}
	return json.Unmarshal(src, umc.MessageComponent)
}

// MessageComponentFromJSON is a helper function for unmarshaling message components
func MessageComponentFromJSON(b []byte) (MessageComponent, error) {
	var u unmarshalableMessageComponent
	err := u.UnmarshalJSON(b)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal into MessageComponent: %w", err)
	}
	return u.MessageComponent, nil
}

// ActionsRow is a container for components within one row.
type ActionsRow struct {
	Components []MessageComponent `json:"components"`
}

// MarshalJSON is a method for marshaling ActionsRow to a JSON object.
func (r ActionsRow) MarshalJSON() ([]byte, error) {
	type actionsRow ActionsRow

	return Marshal(struct {
		actionsRow
		Type ComponentType `json:"type"`
	}{
		actionsRow: actionsRow(r),
		Type:       r.Type(),
	})
}

// UnmarshalJSON is a helper function to unmarshal Actions Row.
func (r *ActionsRow) UnmarshalJSON(data []byte) error {
	var v struct {
		RawComponents []unmarshalableMessageComponent `json:"components"`
	}
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}
	r.Components = make([]MessageComponent, len(v.RawComponents))
	for i, v := range v.RawComponents {
		r.Components[i] = v.MessageComponent
	}

	return err
}

// Type is a method to get the type of a component.
func (r ActionsRow) Type() ComponentType {
	return ActionsRowComponent
}

// ButtonStyle is style of button.
type ButtonStyle uint

// Button styles.
const (
	// PrimaryButton is a button with blurple color.
	PrimaryButton ButtonStyle = 1
	// SecondaryButton is a button with grey color.
	SecondaryButton ButtonStyle = 2
	// SuccessButton is a button with green color.
	SuccessButton ButtonStyle = 3
	// DangerButton is a button with red color.
	DangerButton ButtonStyle = 4
	// LinkButton is a special type of button which navigates to a URL. Has grey color.
	LinkButton ButtonStyle = 5
)

// ComponentEmoji represents button emoji, if it does have one.
type ComponentEmoji struct {
	Name     string `json:"name,omitempty"`
	ID       string `json:"id,omitempty"`
	Animated bool   `json:"animated,omitempty"`
}

// Button represents button component.
type Button struct {
	Label    string         `json:"label"`
	Style    ButtonStyle    `json:"style"`
	Disabled bool           `json:"disabled"`
	Emoji    ComponentEmoji `json:"emoji"`

	// NOTE: Only button with LinkButton style can have link. Also, URL is mutually exclusive with CustomID.
	URL      string `json:"url,omitempty"`
	CustomID string `json:"custom_id,omitempty"`
}

// MarshalJSON is a method for marshaling Button to a JSON object.
func (b Button) MarshalJSON() ([]byte, error) {
	type button Button

	if b.Style == 0 {
		b.Style = PrimaryButton
	}

	return Marshal(struct {
		button
		Type ComponentType `json:"type"`
	}{
		button: button(b),
		Type:   b.Type(),
	})
}

// Type is a method to get the type of a component.
func (Button) Type() ComponentType {
	return ButtonComponent
}

// SelectMenuOption represents an option for a select menu.
type SelectMenuOption struct {
	Label       string         `json:"label,omitempty"`
	Value       string         `json:"value"`
	Description string         `json:"description"`
	Emoji       ComponentEmoji `json:"emoji"`
	// Determines whenever option is selected by default or not.
	Default bool `json:"default"`
}

// SelectMenu represents select menu component.
type SelectMenu struct {
	CustomID string `json:"custom_id,omitempty"`
	// The text which will be shown in the menu if there's no default options or all options was deselected and component was closed.
	Placeholder string `json:"placeholder"`
	// This value determines the minimal amount of selected items in the menu.
	MinValues *int `json:"min_values,omitempty"`
	// This value determines the maximal amount of selected items in the menu.
	// If MaxValues or MinValues are greater than one then the user can select multiple items in the component.
	MaxValues int                `json:"max_values,omitempty"`
	Options   []SelectMenuOption `json:"options"`
	Disabled  bool               `json:"disabled"`
}

// Type is a method to get the type of a component.
func (SelectMenu) Type() ComponentType {
	return SelectMenuComponent
}

// MarshalJSON is a method for marshaling SelectMenu to a JSON object.
func (m SelectMenu) MarshalJSON() ([]byte, error) {
	type selectMenu SelectMenu

	return Marshal(struct {
		selectMenu
		Type ComponentType `json:"type"`
	}{
		selectMenu: selectMenu(m),
		Type:       m.Type(),
	})
}

// TextInput represents text input component.
type TextInput struct {
	CustomID    string         `json:"custom_id"`
	Label       string         `json:"label"`
	Style       TextInputStyle `json:"style"`
	Placeholder string         `json:"placeholder,omitempty"`
	Value       string         `json:"value,omitempty"`
	Required    bool           `json:"required"`
	MinLength   int            `json:"min_length,omitempty"`
	MaxLength   int            `json:"max_length,omitempty"`
}

// Type is a method to get the type of a component.
func (TextInput) Type() ComponentType {
	return TextInputComponent
}

// MarshalJSON is a method for marshaling TextInput to a JSON object.
func (m TextInput) MarshalJSON() ([]byte, error) {
	type inputText TextInput

	return Marshal(struct {
		inputText
		Type ComponentType `json:"type"`
	}{
		inputText: inputText(m),
		Type:      m.Type(),
	})
}

// TextInputStyle is style of text in TextInput component.
type TextInputStyle uint

// Text styles
const (
	TextInputShort     TextInputStyle = 1
	TextInputParagraph TextInputStyle = 2
)
