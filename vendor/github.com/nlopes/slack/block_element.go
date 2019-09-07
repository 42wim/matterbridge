package slack

// https://api.slack.com/reference/messaging/block-elements

const (
	METImage      MessageElementType = "image"
	METButton     MessageElementType = "button"
	METOverflow   MessageElementType = "overflow"
	METDatepicker MessageElementType = "datepicker"

	MixedElementImage MixedElementType = "mixed_image"
	MixedElementText  MixedElementType = "mixed_text"

	OptTypeStatic        string = "static_select"
	OptTypeExternal      string = "external_select"
	OptTypeUser          string = "users_select"
	OptTypeConversations string = "conversations_select"
	OptTypeChannels      string = "channels_select"
)

type MessageElementType string
type MixedElementType string

// BlockElement defines an interface that all block element types should implement.
type BlockElement interface {
	ElementType() MessageElementType
}

type MixedElement interface {
	MixedElementType() MixedElementType
}

type Accessory struct {
	ImageElement      *ImageBlockElement
	ButtonElement     *ButtonBlockElement
	OverflowElement   *OverflowBlockElement
	DatePickerElement *DatePickerBlockElement
	SelectElement     *SelectBlockElement
}

// NewAccessory returns a new Accessory for a given block element
func NewAccessory(element BlockElement) *Accessory {
	switch element.(type) {
	case *ImageBlockElement:
		return &Accessory{ImageElement: element.(*ImageBlockElement)}
	case *ButtonBlockElement:
		return &Accessory{ButtonElement: element.(*ButtonBlockElement)}
	case *OverflowBlockElement:
		return &Accessory{OverflowElement: element.(*OverflowBlockElement)}
	case *DatePickerBlockElement:
		return &Accessory{DatePickerElement: element.(*DatePickerBlockElement)}
	case *SelectBlockElement:
		return &Accessory{SelectElement: element.(*SelectBlockElement)}
	}

	return nil
}

// BlockElements is a convenience struct defined to allow dynamic unmarshalling of
// the "elements" value in Slack's JSON response, which varies depending on BlockElement type
type BlockElements struct {
	ElementSet []BlockElement `json:"elements,omitempty"`
}

// ImageBlockElement An element to insert an image - this element can be used
// in section and context blocks only. If you want a block with only an image
// in it, you're looking for the image block.
//
// More Information: https://api.slack.com/reference/messaging/block-elements#image
type ImageBlockElement struct {
	Type     MessageElementType `json:"type"`
	ImageURL string             `json:"image_url"`
	AltText  string             `json:"alt_text"`
}

// ElementType returns the type of the Element
func (s ImageBlockElement) ElementType() MessageElementType {
	return s.Type
}

func (s ImageBlockElement) MixedElementType() MixedElementType {
	return MixedElementImage
}

// NewImageBlockElement returns a new instance of an image block element
func NewImageBlockElement(imageURL, altText string) *ImageBlockElement {
	return &ImageBlockElement{
		Type:     METImage,
		ImageURL: imageURL,
		AltText:  altText,
	}
}

type Style string

const (
	StyleDefault Style = "default"
	StylePrimary Style = "primary"
	StyleDanger  Style = "danger"
)

// ButtonBlockElement defines an interactive element that inserts a button. The
// button can be a trigger for anything from opening a simple link to starting
// a complex workflow.
//
// More Information: https://api.slack.com/reference/messaging/block-elements#button
type ButtonBlockElement struct {
	Type     MessageElementType       `json:"type,omitempty"`
	Text     *TextBlockObject         `json:"text"`
	ActionID string                   `json:"action_id,omitempty"`
	URL      string                   `json:"url,omitempty"`
	Value    string                   `json:"value,omitempty"`
	Confirm  *ConfirmationBlockObject `json:"confirm,omitempty"`
	Style    Style                    `json:"style,omitempty"`
}

// ElementType returns the type of the element
func (s ButtonBlockElement) ElementType() MessageElementType {
	return s.Type
}

// add styling to button object
func (s *ButtonBlockElement) WithStyle(style Style) {
	s.Style = style
}

// NewButtonBlockElement returns an instance of a new button element to be used within a block
func NewButtonBlockElement(actionID, value string, text *TextBlockObject) *ButtonBlockElement {
	return &ButtonBlockElement{
		Type:     METButton,
		ActionID: actionID,
		Text:     text,
		Value:    value,
	}
}

// SelectBlockElement defines the simplest form of select menu, with a static list
// of options passed in when defining the element.
//
// More Information: https://api.slack.com/reference/messaging/block-elements#select
type SelectBlockElement struct {
	Type                string                    `json:"type,omitempty"`
	Placeholder         *TextBlockObject          `json:"placeholder,omitempty"`
	ActionID            string                    `json:"action_id,omitempty"`
	Options             []*OptionBlockObject      `json:"options,omitempty"`
	OptionGroups        []*OptionGroupBlockObject `json:"option_groups,omitempty"`
	InitialOption       *OptionBlockObject        `json:"initial_option,omitempty"`
	InitialUser         string                    `json:"initial_user,omitempty"`
	InitialConversation string                    `json:"initial_conversation,omitempty"`
	InitialChannel      string                    `json:"initial_channel,omitempty"`
	MinQueryLength      int                       `json:"min_query_length,omitempty"`
	Confirm             *ConfirmationBlockObject  `json:"confirm,omitempty"`
}

// ElementType returns the type of the Element
func (s SelectBlockElement) ElementType() MessageElementType {
	return MessageElementType(s.Type)
}

// NewOptionsSelectBlockElement returns a new instance of SelectBlockElement for use with
// the Options object only.
func NewOptionsSelectBlockElement(optType string, placeholder *TextBlockObject, actionID string, options ...*OptionBlockObject) *SelectBlockElement {
	return &SelectBlockElement{
		Type:        optType,
		Placeholder: placeholder,
		ActionID:    actionID,
		Options:     options,
	}
}

// NewOptionsGroupSelectBlockElement returns a new instance of SelectBlockElement for use with
// the Options object only.
func NewOptionsGroupSelectBlockElement(
	optType string,
	placeholder *TextBlockObject,
	actionID string,
	optGroups ...*OptionGroupBlockObject,
) *SelectBlockElement {
	return &SelectBlockElement{
		Type:         optType,
		Placeholder:  placeholder,
		ActionID:     actionID,
		OptionGroups: optGroups,
	}
}

// OverflowBlockElement defines the fields needed to use an overflow element.
// And Overflow Element is like a cross between a button and a select menu -
// when a user clicks on this overflow button, they will be presented with a
// list of options to choose from.
//
// More Information: https://api.slack.com/reference/messaging/block-elements#overflow
type OverflowBlockElement struct {
	Type     MessageElementType       `json:"type"`
	ActionID string                   `json:"action_id,omitempty"`
	Options  []*OptionBlockObject     `json:"options"`
	Confirm  *ConfirmationBlockObject `json:"confirm,omitempty"`
}

// ElementType returns the type of the Element
func (s OverflowBlockElement) ElementType() MessageElementType {
	return s.Type
}

// NewOverflowBlockElement returns an instance of a new Overflow Block Element
func NewOverflowBlockElement(actionID string, options ...*OptionBlockObject) *OverflowBlockElement {
	return &OverflowBlockElement{
		Type:     METOverflow,
		ActionID: actionID,
		Options:  options,
	}
}

// DatePickerBlockElement defines an element which lets users easily select a
// date from a calendar style UI. Date picker elements can be used inside of
// section and actions blocks.
//
// More Information: https://api.slack.com/reference/messaging/block-elements#datepicker
type DatePickerBlockElement struct {
	Type        MessageElementType       `json:"type"`
	ActionID    string                   `json:"action_id"`
	Placeholder *TextBlockObject         `json:"placeholder,omitempty"`
	InitialDate string                   `json:"initial_date,omitempty"`
	Confirm     *ConfirmationBlockObject `json:"confirm,omitempty"`
}

// ElementType returns the type of the Element
func (s DatePickerBlockElement) ElementType() MessageElementType {
	return s.Type
}

// NewDatePickerBlockElement returns an instance of a date picker element
func NewDatePickerBlockElement(actionID string) *DatePickerBlockElement {
	return &DatePickerBlockElement{
		Type:     METDatepicker,
		ActionID: actionID,
	}
}
