package settings

type SyncSource int

const (
	FromInterface SyncSource = iota + 1
	FromStruct
)

type ProfilePicturesVisibilityType int

const (
	ProfilePicturesVisibilityContactsOnly ProfilePicturesVisibilityType = iota + 1
	ProfilePicturesVisibilityEveryone
	ProfilePicturesVisibilityNone
)

type ProfilePicturesShowToType int

const (
	ProfilePicturesShowToContactsOnly ProfilePicturesShowToType = iota + 1
	ProfilePicturesShowToEveryone
	ProfilePicturesShowToNone
)

type URLUnfurlingModeType int

const (
	URLUnfurlingAlwaysAsk URLUnfurlingModeType = iota + 1
	URLUnfurlingEnableAll
	URLUnfurlingDisableAll
)
