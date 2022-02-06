package logging

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

type ctxKey string

const loggerKey ctxKey = "logger"

type Logger struct {
	// Entry for every other log level asides access logs
	*logrus.Entry

	accessEntry *logrus.Entry
	closeFuncs  [2]func() error
}

func New(env, logPath string) (*Logger, error) {
	var errLogger, accessLogger *logrus.Logger
	logger := &Logger{
		closeFuncs: [2]func() error{},
	}
	if env == "production" {
		accessPath := filepath.Join(logPath, "access.log")
		af, err := os.OpenFile(accessPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("error creating access files")
		}
		accessLogger = logrus.New()
		accessLogger.Out = af

		errPath := filepath.Join(logPath, "errors.log")
		f, err := os.OpenFile(errPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		errLogger = logrus.New()
		errLogger.Out = f

		logger.closeFuncs[0] = af.Close
		logger.closeFuncs[1] = f.Close
	} else {
		accessLogger = logrus.New()
		errLogger = logrus.New()
	}
	accessLogger.Level = logrus.InfoLevel
	logger.accessEntry = logrus.NewEntry(accessLogger)

	errLogger.Level = logrus.ErrorLevel
	logger.Entry = logrus.NewEntry(errLogger)
	return logger, nil
}

func (l *Logger) Close() error {
	for _, c := range l.closeFuncs {
		if c != nil {
			if err := c(); err != nil {
				return err
			}
		}
	}
	return nil
}

func FromContext(ctx context.Context) *Logger {
	val, ok := ctx.Value(loggerKey).(*Logger)
	if val == nil || !ok {
		return &Logger{
			accessEntry: logrus.NewEntry(logrus.New()),
			Entry:       logrus.NewEntry(logrus.New()),
			closeFuncs:  [2]func() error{},
		}
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

func (l *Logger) LogAccess(method string, statusCode int) {
	l.accessEntry.WithFields(logrus.Fields{
		"method": method,
		"status": statusCode,
	}).Info("new request")
	//	l.Entry.
}
