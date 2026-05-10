package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWaitForWorkerStopAlreadyStopped(t *testing.T) {
	t.Parallel()

	errCh := make(chan error)

	if err := waitForWorkerStop(context.Background(), time.Second, true, errCh); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestWaitForWorkerStopReturnsConsumerError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("consumer failed")
	errCh := make(chan error, 1)
	errCh <- expectedErr

	err := waitForWorkerStop(context.Background(), time.Second, false, errCh)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestWaitForWorkerStopIgnoresShutdownTimeout(t *testing.T) {
	t.Parallel()

	errCh := make(chan error)

	if err := waitForWorkerStop(context.Background(), time.Nanosecond, false, errCh); err != nil {
		t.Fatalf("expected nil timeout error, got %v", err)
	}
}
