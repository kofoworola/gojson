package logging

import (
	"context"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

type ctxKey string

const loggerKey ctxKey = "logger"

type Logger struct {
	*logrus.Entry

	Logger    *logrus.Logger
	closeFunc func() error
}

func New(env, logPath string) (*Logger, error) {
	l := logrus.New()
	logger := &Logger{
		Logger: l,
	}
	if env == "production" {
		path := filepath.Join(logPath, "gojson.log")
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		logger.closeFunc = f.Close
		l.Out = f
		l.Level = logrus.ErrorLevel
	} else {
		l.Level = logrus.DebugLevel
	}
	logger.Entry = logrus.NewEntry(l)
	return logger, nil
}

func (l *Logger) Close() error {
	if l.closeFunc != nil {
		return l.closeFunc()
	} else {
		return nil
	}
}

func FromContext(ctx context.Context) *Logger {
	val, ok := ctx.Value(loggerKey).(*Logger)
	if val == nil || !ok {
		return &Logger{Logger: logrus.New()}
	}
	return val
}

func ToContext(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

func (l *Logger) WithField(key string, value interface{}) *Logger {
	logger := *l
	logger.Entry = logger.Entry.WithField(key, value)
	return &logger
}
