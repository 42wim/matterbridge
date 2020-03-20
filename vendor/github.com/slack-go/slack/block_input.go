package slack

// InputBlock defines data that is used to collect information from users -
// it can hold a plain-text input element, a select menu element,
// a multi-select menu element, or a datepicker.
//
// More Information: https://api.slack.com/reference/messaging/blocks#input
type InputBlock struct {
	Type     MessageBlockType `json:"type"`
	BlockID  string           `json:"block_id,omitempty"`
	Label    *TextBlockObject `json:"label"`
	Element  BlockElement     `json:"element"`
	Hint     *TextBlockObject `json:"hint,omitempty"`
	Optional bool             `json:"optional,omitempty"`
}

// BlockType returns the type of the block
func (s InputBlock) BlockType() MessageBlockType {
	return s.Type
}

// NewInputBlock returns a new instance of an Input Block
func NewInputBlock(blockID string, label *TextBlockObject, element BlockElement) *InputBlock {
	return &InputBlock{
		Type:    MBTInput,
		BlockID: blockID,
		Label:   label,
		Element: element,
	}
}
