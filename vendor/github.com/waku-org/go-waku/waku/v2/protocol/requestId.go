package protocol

import (
	"crypto/rand"
	"sync"

	"github.com/cruxic/go-hmac-drbg/hmacdrbg"
	"github.com/waku-org/go-waku/waku/v2/utils"
	"go.uber.org/zap"
)

var brHmacDrbgPool = sync.Pool{New: func() interface{} {
	seed := make([]byte, 48)
	_, err := rand.Read(seed)
	if err != nil {
		utils.Logger().Fatal("rand.Read err", zap.Error(err))
	}
	return hmacdrbg.NewHmacDrbg(256, seed, nil)
}}

// GenerateRequestID generates a random 32 byte slice that can be used for
// creating requests inf the filter, store and lightpush protocols
func GenerateRequestID() []byte {
	rng := brHmacDrbgPool.Get().(*hmacdrbg.HmacDrbg)
	defer brHmacDrbgPool.Put(rng)

	randData := make([]byte, 32)
	if !rng.Generate(randData) {
		//Reseed is required every 10,000 calls
		seed := make([]byte, 48)
		_, err := rand.Read(seed)
		if err != nil {
			utils.Logger().Fatal("rand.Read err", zap.Error(err))
		}
		err = rng.Reseed(seed)
		if err != nil {
			//only happens if seed < security-level
			utils.Logger().Fatal("rng.Reseed err", zap.Error(err))
		}

		if !rng.Generate(randData) {
			utils.Logger().Error("could not generate random request id")
		}
	}
	return randData
}
