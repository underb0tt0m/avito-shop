package mocks

import "avito-shop/internal/logging"

type logger struct {
	SyncFunc func() error
}

func NewLogger(SyncFunc func() error) logging.Logger {
	return logger{SyncFunc: SyncFunc}
}

func (l logger) Debug(msg string) {}

func (l logger) Info(msg string) {}

func (l logger) Warn(msg string, err error) {}

func (l logger) Error(msg string, err error) {}

func (l logger) Fatal(msg string, err error) {}

func (l logger) Sync() error {
	if l.SyncFunc != nil {
		return l.SyncFunc()
	}
	return nil
}
