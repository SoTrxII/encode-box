package progress_broker

import (
	"context"
	mock_client "encode-box/internal/mock/dapr"
	console_parser "encode-box/pkg/encoder/console-parser"
	"fmt"
	"github.com/golang/mock/gomock"
	"testing"
	"time"
)

func TestProgressBroker_SendProgress(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	daprClient := mock_client.NewMockClient(ctrl)
	daprClient.EXPECT().PublishEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	pg, err := NewProgressBroker[*mock_client.MockClient](&ctx, &daprClient, NewBrokerOptions{
		Component: "",
		Topic:     "",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = pg.SendProgress(EncodeInfos{
		JobId: "1",
		State: InProgress,
		Data: console_parser.EncodingProgress{
			Frames:  0,
			Fps:     0,
			Quality: 0,
			Size:    0,
			Time:    time.Time{},
			Bitrate: "",
			Speed:   0,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestProgressBroker_SendError(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	daprClient := mock_client.NewMockClient(ctrl)
	daprClient.EXPECT().PublishEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	pg, err := NewProgressBroker[*mock_client.MockClient](&ctx, &daprClient, NewBrokerOptions{
		Component: "",
		Topic:     "",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = pg.SendProgress(EncodeInfos{
		JobId: "1",
		State: Error,
		Data:  fmt.Errorf("Test"),
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestProgressBroker_SendDone(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	daprClient := mock_client.NewMockClient(ctrl)
	daprClient.EXPECT().PublishEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	pg, err := NewProgressBroker[*mock_client.MockClient](&ctx, &daprClient, NewBrokerOptions{
		Component: "",
		Topic:     "",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = pg.SendProgress(EncodeInfos{
		JobId: "1",
		State: Done,
		Data:  nil,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestProgressBroker_CouldNotSend(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	daprClient := mock_client.NewMockClient(ctrl)
	daprClient.EXPECT().PublishEvent(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("test"))
	pg, err := NewProgressBroker[*mock_client.MockClient](&ctx, &daprClient, NewBrokerOptions{
		Component: "",
		Topic:     "",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = pg.SendProgress(EncodeInfos{
		JobId: "1",
		State: Done,
		Data:  nil,
	})
	if err == nil {
		t.Fatal(err)
	}
}
