package progress_broker

import (
	"context"
	"encode-box/internal/utils"
	"encoding/json"
)

type ProgressBroker struct {
	// Name of the Dapr Component to use
	componentName string
	// Name of the topic to publish into
	topic string
	// Client to publish event into
	client utils.Publisher
	// Current running context
	ctx *context.Context
}

type EncodeState int8

const (
	InProgress EncodeState = iota
	Done
	Error
)

type EncodeInfos struct {
	JobId string      `json:"jobId"`
	State EncodeState `json:"state"`
	Data  interface{} `json:"data"`
}
type NewBrokerOptions struct {
	Component string
	Topic     string
}

func NewProgressBroker(ctx *context.Context, client utils.Publisher, opt NewBrokerOptions) (*ProgressBroker, error) {
	return &ProgressBroker{
		componentName: opt.Component,
		topic:         opt.Topic,
		client:        client,
		ctx:           ctx,
	}, nil
}

func (eb *ProgressBroker) SendProgress(data EncodeInfos) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = eb.client.PublishEvent(*eb.ctx, eb.componentName, eb.topic, string(b))
	if err != nil {
		return err
	}
	return nil
}
