package logger

import "go.uber.org/zap"

var _ Logger = (*zapLogger)(nil)

type zapLogger struct {
	sugar *zap.SugaredLogger
}

func NewZapLogger() (Logger, error) {
	zl, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}

	return &zapLogger{sugar: zl.Sugar()}, nil
}

func (z *zapLogger) Debug(msg string, args ...interface{}) {
	z.sugar.Debugf(msg, args...)
}

func (z *zapLogger) Info(msg string, args ...interface{}) {
	z.sugar.Infof(msg, args...)
}

func (z *zapLogger) Warn(msg string, args ...interface{}) {
	z.sugar.Warnf(msg, args...)
}

func (z *zapLogger) Error(msg string, args ...interface{}) {
	z.sugar.Errorf(msg, args...)
}

func (z *zapLogger) Sync() error {
	return z.sugar.Sync()
}
