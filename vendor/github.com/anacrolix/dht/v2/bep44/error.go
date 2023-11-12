package bep44

import "github.com/anacrolix/dht/v2/krpc"

var (
	ErrValueFieldTooBig = krpc.Error{
		Code: krpc.ErrorCodeMessageValueFieldTooBig,
		Msg:  "message (v field) too big",
	}
	ErrInvalidSignature = krpc.Error{
		Code: krpc.ErrorCodeInvalidSignature,
		Msg:  "invalid signature",
	}

	ErrSaltFieldTooBig = krpc.Error{
		Code: krpc.ErrorCodeSaltFieldTooBig,
		Msg:  "salt (salt field) too big",
	}
	ErrCasHashMismatched = krpc.Error{
		Code: krpc.ErrorCodeCasHashMismatched,
		Msg:  "the CAS hash mismatched, re-read value and try again",
	}
	ErrSequenceNumberLessThanCurrent = krpc.Error{
		Code: krpc.ErrorCodeSequenceNumberLessThanCurrent,
		Msg:  "sequence number less than current",
	}
)
