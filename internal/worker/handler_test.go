package worker

import (
	"context"
	"encoding/json"
	"testing"

	async "github.com/kitti12911/lib-async"
)

func TestHandleLogsJobAndReturnsNil(t *testing.T) {
	t.Parallel()

	handler := NewHandler()
	msg := async.Envelope[Job]{
		UUID: "message-1",
		Payload: Job{
			ID:      "job-1",
			Type:    "debug.print",
			Payload: json.RawMessage(`{"message":"hello"}`),
		},
	}

	if err := handler.Handle(context.Background(), msg); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
