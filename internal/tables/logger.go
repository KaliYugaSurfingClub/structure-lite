package tables

type logger interface {
	Info(msg string, args ...any)
}

type NoLogger struct{}

func (NoLogger) Info(string, ...any) {}
