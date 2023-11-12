package common

import (
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/rlp"
	"github.com/status-im/status-go/eth-node/types"
)

type Measure struct {
	Timestamp int64
	Size      uint64
}

type StatsTracker struct {
	Uploads   []Measure
	Downloads []Measure

	statsMutex sync.Mutex
}

const measurementPeriod = 15 * time.Second

func measure(input interface{}) (*Measure, error) {
	b, err := rlp.EncodeToBytes(input)
	if err != nil {
		return nil, err
	}
	return &Measure{
		Timestamp: time.Now().UnixNano(),
		Size:      uint64(len(b)),
	}, nil

}

func (s *StatsTracker) AddUpload(input interface{}) {
	go func(input interface{}) {
		m, err := measure(input)
		if err != nil {
			return
		}

		s.statsMutex.Lock()
		defer s.statsMutex.Unlock()
		s.Uploads = append(s.Uploads, *m)
	}(input)
}

func (s *StatsTracker) AddDownload(input interface{}) {
	go func(input interface{}) {
		m, err := measure(input)
		if err != nil {
			return
		}

		s.statsMutex.Lock()
		defer s.statsMutex.Unlock()
		s.Downloads = append(s.Downloads, *m)
	}(input)
}

func (s *StatsTracker) AddUploadBytes(size uint64) {
	go func(size uint64) {
		m := Measure{
			Timestamp: time.Now().UnixNano(),
			Size:      size,
		}

		s.statsMutex.Lock()
		defer s.statsMutex.Unlock()
		s.Uploads = append(s.Uploads, m)
	}(size)
}

func (s *StatsTracker) AddDownloadBytes(size uint64) {
	go func(size uint64) {
		m := Measure{
			Timestamp: time.Now().UnixNano(),
			Size:      size,
		}

		s.statsMutex.Lock()
		defer s.statsMutex.Unlock()
		s.Downloads = append(s.Downloads, m)
	}(size)
}

func calculateAverage(measures []Measure, minTime int64) (validMeasures []Measure, rate uint64) {
	for _, m := range measures {
		if m.Timestamp > minTime {
			// Only use recent measures
			validMeasures = append(validMeasures, m)
			rate += m.Size
		}
	}

	rate /= (uint64(measurementPeriod) / uint64(1*time.Second))
	return
}

func (s *StatsTracker) GetRatePerSecond() (uploadRate uint64, downloadRate uint64) {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()
	minTime := time.Now().Add(-measurementPeriod).UnixNano()
	s.Uploads, uploadRate = calculateAverage(s.Uploads, minTime)
	s.Downloads, downloadRate = calculateAverage(s.Downloads, minTime)
	return
}

func (s *StatsTracker) GetStats() types.StatsSummary {
	uploadRate, downloadRate := s.GetRatePerSecond()
	summary := types.StatsSummary{
		UploadRate:   uploadRate,
		DownloadRate: downloadRate,
	}
	return summary
}
