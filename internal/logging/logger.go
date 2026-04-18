package logging

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string, err error)
	Error(msg string, err error)
	Fatal(msg string, err error)
	Sync() error
}
