package wakusync

import (
	"github.com/status-im/status-go/protocol/protobuf"
)

type FetchingBackupedDataDetails struct {
	DataNumber  uint32 `json:"dataNumber,omitempty"`
	TotalNumber uint32 `json:"totalNumber,omitempty"`
}

func (sfwr *WakuBackedUpDataResponse) AddFetchingBackedUpDataDetails(section string, details *protobuf.FetchingBackedUpDataDetails) {
	if details == nil {
		return
	}
	if sfwr.FetchingDataProgress == nil {
		sfwr.FetchingDataProgress = make(map[string]*protobuf.FetchingBackedUpDataDetails)
	}

	sfwr.FetchingDataProgress[section] = details
}

func (sfwr *WakuBackedUpDataResponse) FetchingBackedUpDataDetails() map[string]FetchingBackupedDataDetails {
	if len(sfwr.FetchingDataProgress) == 0 {
		return nil
	}

	result := make(map[string]FetchingBackupedDataDetails)
	for section, details := range sfwr.FetchingDataProgress {
		result[section] = FetchingBackupedDataDetails{
			DataNumber:  details.DataNumber,
			TotalNumber: details.TotalNumber,
		}
	}
	return result
}
