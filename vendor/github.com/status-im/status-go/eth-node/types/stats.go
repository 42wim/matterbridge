package types

type StatsSummary struct {
	UploadRate   uint64 `json:"uploadRate"`
	DownloadRate uint64 `json:"downloadRate"`
}
