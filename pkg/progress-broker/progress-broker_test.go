package progress_broker

import (
	"context"
	console_parser "encode-box/pkg/encoder/console-parser"
	test_utils "encode-box/test-utils"
	"fmt"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestProgressBroker_SendProgress(t *testing.T) {
	ctx := context.Background()
	pub := test_utils.MockPublisher{}
	pub.EXPECT().PublishEvent(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	pg, err := NewProgressBroker(&ctx, &pub, NewBrokerOptions{
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
			Time:    0,
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
	pub := test_utils.MockPublisher{}
	pub.EXPECT().PublishEvent(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	pg, err := NewProgressBroker(&ctx, &pub, NewBrokerOptions{
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
	pub := test_utils.MockPublisher{}
	pub.EXPECT().PublishEvent(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	pg, err := NewProgressBroker(&ctx, &pub, NewBrokerOptions{
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
	pub := test_utils.MockPublisher{}
	pub.EXPECT().PublishEvent(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("test"))
	pg, err := NewProgressBroker(&ctx, &pub, NewBrokerOptions{
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
