package discord

import (
	"errors"
	"fmt"
	"sync"

	"github.com/status-im/status-go/images"
	"github.com/status-im/status-go/protocol/protobuf"
)

type ErrorCodeType uint

const (
	NoErrorType ErrorCodeType = iota + 1
	WarningType
	ErrorType
)

const MaxTaskErrorItemsCount = 3
const MaxImportFileSizeBytes = 52428800

var (
	ErrNoChannelData    = errors.New("No channels to import messages from")
	ErrNoMessageData    = errors.New("No messages to import")
	ErrMarshalMessage   = errors.New("Couldn't marshal discord message")
	ErrImportFileTooBig = fmt.Errorf("File is too big (max. %d MB)", MaxImportFileSizeBytes/1024/1024)
)

type MessageType string

const (
	MessageTypeDefault       MessageType = "Default"
	MessageTypeReply         MessageType = "Reply"
	MessageTypeChannelPinned MessageType = "ChannelPinnedMessage"
)

type ImportTask uint

const (
	CommunityCreationTask ImportTask = iota + 1
	ChannelsCreationTask
	ImportMessagesTask
	DownloadAssetsTask
	InitCommunityTask
)

func (t ImportTask) String() string {
	switch t {
	case CommunityCreationTask:
		return "import.communityCreation"
	case ChannelsCreationTask:
		return "import.channelsCreation"
	case ImportMessagesTask:
		return "import.importMessages"
	case DownloadAssetsTask:
		return "import.downloadAssets"
	case InitCommunityTask:
		return "import.initializeCommunity"
	}
	return "unknown"
}

type ImportTaskState uint

const (
	TaskStateInitialized ImportTaskState = iota
	TaskStateSaving
)

func (t ImportTaskState) String() string {
	switch t {
	case TaskStateInitialized:
		return "import.taskState.initialized"
	case TaskStateSaving:
		return "import.taskState.saving"
	}
	return "import.taskState.unknown"
}

type Channel struct {
	ID           string `json:"id"`
	CategoryName string `json:"category"`
	CategoryID   string `json:"categoryId"`
	Name         string `json:"name"`
	Description  string `json:"topic"`
	FilePath     string `json:"filePath"`
}

type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ExportedData struct {
	Channel      Channel                    `json:"channel"`
	Messages     []*protobuf.DiscordMessage `json:"messages"`
	MessageCount int                        `json:"messageCount"`
}

type ExtractedData struct {
	Categories             map[string]*Category
	ExportedData           []*ExportedData
	OldestMessageTimestamp int
	MessageCount           int
}

type ImportError struct {
	// This code is used to distinguish between errors
	// that are considered "criticial" and those that are not.
	//
	// Critical errors are the ones that prevent the imported community
	// from functioning properly. For example, if the creation of the community
	// or its categories and channels fails, this is a critical error.
	//
	// Non-critical errors are the ones that would not prevent the imported
	// community from functioning. For example, if the channel data to be imported
	// has no messages, or is not parsable.
	Code     ErrorCodeType `json:"code"`
	Message  string        `json:"message"`
	TaskInfo string        `json:"taskInfo"`
}

func (d ImportError) Error() string {
	return fmt.Sprintf("%d: %s", d.Code, d.Message)
}

func Error(message string) *ImportError {
	return &ImportError{
		Message: message,
		Code:    ErrorType,
	}
}

func Warning(message string) *ImportError {
	return &ImportError{
		Message: message,
		Code:    WarningType,
	}
}

type ImportTaskProgress struct {
	Type          string         `json:"type"`
	Progress      float32        `json:"progress"`
	Errors        []*ImportError `json:"errors"`
	Stopped       bool           `json:"stopped"`
	ErrorsCount   uint           `json:"errorsCount"`
	WarningsCount uint           `json:"warningsCount"`
	State         string         `json:"state"`
}

type ImportTasks map[ImportTask]*ImportTaskProgress

type ImportProgress struct {
	CommunityID     string                          `json:"communityId,omitempty"`
	CommunityName   string                          `json:"communityName"`
	ChannelID       string                          `json:"channelId"`
	ChannelName     string                          `json:"channelName"`
	CommunityImages map[string]images.IdentityImage `json:"communityImages"`
	Tasks           []*ImportTaskProgress           `json:"tasks"`
	Progress        float32                         `json:"progress"`
	ErrorsCount     uint                            `json:"errorsCount"`
	WarningsCount   uint                            `json:"warningsCount"`
	Stopped         bool                            `json:"stopped"`
	TotalChunkCount int                             `json:"totalChunksCount,omitempty"`
	CurrentChunk    int                             `json:"currentChunk,omitempty"`
	m               sync.Mutex
}

func (progress *ImportProgress) Init(totalChunkCount int, tasks []ImportTask) {
	progress.Progress = 0
	progress.Tasks = make([]*ImportTaskProgress, 0)
	for _, task := range tasks {
		progress.Tasks = append(progress.Tasks, &ImportTaskProgress{
			Type:          task.String(),
			Progress:      0,
			Errors:        []*ImportError{},
			Stopped:       false,
			ErrorsCount:   0,
			WarningsCount: 0,
			State:         TaskStateInitialized.String(),
		})
	}
	progress.ErrorsCount = 0
	progress.WarningsCount = 0
	progress.Stopped = false
	progress.TotalChunkCount = totalChunkCount
	progress.CurrentChunk = 0
}

func (progress *ImportProgress) Stop() {
	progress.Stopped = true
}

func (progress *ImportProgress) AddTaskError(task ImportTask, err *ImportError) {
	progress.m.Lock()
	defer progress.m.Unlock()

	for i, t := range progress.Tasks {
		if t.Type == task.String() {
			errorsAndWarningsCount := progress.Tasks[i].ErrorsCount + progress.Tasks[i].WarningsCount
			if (errorsAndWarningsCount < MaxTaskErrorItemsCount) || err.Code > WarningType {
				errors := progress.Tasks[i].Errors
				progress.Tasks[i].Errors = append(errors, err)
			}
			if err.Code > WarningType {
				progress.Tasks[i].ErrorsCount++
			}
			if err.Code > NoErrorType {
				progress.Tasks[i].WarningsCount++
			}
		}
	}
	if err.Code > WarningType {
		progress.ErrorsCount++
		return
	}
	if err.Code > NoErrorType {
		progress.WarningsCount++
	}
}

func (progress *ImportProgress) StopTask(task ImportTask) {
	progress.m.Lock()
	defer progress.m.Unlock()
	for i, t := range progress.Tasks {
		if t.Type == task.String() {
			progress.Tasks[i].Stopped = true
		}
	}
	progress.Stop()
}

func (progress *ImportProgress) UpdateTaskProgress(task ImportTask, value float32) {
	progress.m.Lock()
	defer progress.m.Unlock()
	for i, t := range progress.Tasks {
		if t.Type == task.String() {
			progress.Tasks[i].Progress = value
		}
	}
	sum := float32(0)
	for _, t := range progress.Tasks {
		sum = sum + t.Progress
	}
	// Update total progress now that sub progress has changed
	progress.Progress = sum / float32(len(progress.Tasks))
}

func (progress *ImportProgress) UpdateTaskState(task ImportTask, state ImportTaskState) {
	progress.m.Lock()
	defer progress.m.Unlock()
	for i, t := range progress.Tasks {
		if t.Type == task.String() {
			progress.Tasks[i].State = state.String()
		}
	}
}

type AssetCounter struct {
	m sync.RWMutex
	v uint64
}

func (c *AssetCounter) Value() uint64 {
	c.m.RLock()
	defer c.m.RUnlock()
	return c.v
}

func (c *AssetCounter) Increase() {
	c.m.Lock()
	c.v++
	c.m.Unlock()
}
