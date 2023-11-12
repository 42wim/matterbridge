package anonmetrics

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/protobuf/ptypes"

	"github.com/status-im/status-go/appmetrics"
	"github.com/status-im/status-go/eth-node/crypto"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/protocol/protobuf"
)

// adaptProtoToModel is an adaptor helper function to convert a protobuf.AnonymousMetric into a appmetrics.AppMetric
func adaptProtoToModel(pbAnonMetric *protobuf.AnonymousMetric) (*appmetrics.AppMetric, error) {
	t, err := ptypes.Timestamp(pbAnonMetric.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &appmetrics.AppMetric{
		MessageID:  pbAnonMetric.Id,
		Event:      appmetrics.AppMetricEventType(pbAnonMetric.Event),
		Value:      pbAnonMetric.Value,
		AppVersion: pbAnonMetric.AppVersion,
		OS:         pbAnonMetric.Os,
		SessionID:  pbAnonMetric.SessionId,
		CreatedAt:  t,
	}, nil
}

// adaptModelToProto is an adaptor helper function to convert a appmetrics.AppMetric into a protobuf.AnonymousMetric
func adaptModelToProto(modelAnonMetric appmetrics.AppMetric, sendID *ecdsa.PublicKey) (*protobuf.AnonymousMetric, error) {
	id := generateProtoID(modelAnonMetric, sendID)
	createdAt, err := ptypes.TimestampProto(modelAnonMetric.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &protobuf.AnonymousMetric{
		Id:         id,
		Event:      string(modelAnonMetric.Event),
		Value:      modelAnonMetric.Value,
		AppVersion: modelAnonMetric.AppVersion,
		Os:         modelAnonMetric.OS,
		SessionId:  modelAnonMetric.SessionID,
		CreatedAt:  createdAt,
	}, nil
}

func adaptModelsToProtoBatch(modelAnonMetrics []appmetrics.AppMetric, sendID *ecdsa.PublicKey) (*protobuf.AnonymousMetricBatch, error) {
	amb := new(protobuf.AnonymousMetricBatch)

	for _, m := range modelAnonMetrics {
		p, err := adaptModelToProto(m, sendID)
		if err != nil {
			return nil, err
		}

		amb.Metrics = append(amb.Metrics, p)
	}

	return amb, nil
}

func adaptProtoBatchToModels(protoBatch *protobuf.AnonymousMetricBatch) ([]*appmetrics.AppMetric, error) {
	if protoBatch == nil {
		return nil, nil
	}

	var ams []*appmetrics.AppMetric

	for _, pm := range protoBatch.Metrics {
		m, err := adaptProtoToModel(pm)
		if err != nil {
			return nil, err
		}

		ams = append(ams, m)
	}

	return ams, nil
}

// NEEDED because we don't send individual metrics, we send only batches
func generateProtoID(modelAnonMetric appmetrics.AppMetric, sendID *ecdsa.PublicKey) string {
	return types.EncodeHex(crypto.Keccak256([]byte(fmt.Sprintf(
		"%s%s",
		types.EncodeHex(crypto.FromECDSAPub(sendID)),
		spew.Sdump(modelAnonMetric)))))
}
