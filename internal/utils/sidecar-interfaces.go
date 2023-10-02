// This package is mainly interfaces used elsewhere in the codebase.
// All of these are used for easier mocking and testing
package utils

import (
	"context"
	dapr "github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/service/common"
)

type PublishEventOption = dapr.PublishEventOption
type Publisher interface {
	PublishEvent(ctx context.Context, pubsubName string, topicName string, data interface{}, opts ...PublishEventOption) error
}
type Subscriber interface {
	AddTopicEventHandler(sub *common.Subscription, fn common.TopicEventHandler) error
}

type DataContent = dapr.DataContent
type Invoker interface {
	InvokeMethodWithContent(ctx context.Context, appID, method, verb string, content *DataContent) ([]byte, error)
}

type StateOption = dapr.StateOption
type StateItem = dapr.StateItem
type StateSaver interface {
	GetState(ctx context.Context, storeName string, key string, meta map[string]string) (item *StateItem, err error)
	SaveState(ctx context.Context, storeName string, key string, data []byte, meta map[string]string, so ...StateOption) error
	DeleteState(ctx context.Context, storeName string, key string, meta map[string]string) error
}

type InvokeBindingRequest = dapr.InvokeBindingRequest
type BindingEvent = dapr.BindingEvent

// Proxy to query the backend storage
type Binder interface {
	// Invoke
	InvokeBinding(ctx context.Context, in *InvokeBindingRequest) (out *BindingEvent, err error)
}
