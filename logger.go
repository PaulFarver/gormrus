package gormrus

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type Logger struct {
	entry interface {
		logrus.FieldLogger
		WithContext(context.Context) *logrus.Entry
	}
	IgnoreRecordNotFoundError bool
	SlowThreshold             time.Duration
}

func (l *Logger) LogMode(gormlogger.LogLevel) gormlogger.Interface {
	return l
}

func (l *Logger) Info(ctx context.Context, str string, v ...interface{}) {
	l.entry.WithContext(ctx).Infof(str, v...)
}

func (l *Logger) Warn(ctx context.Context, str string, v ...interface{}) {
	l.entry.WithContext(ctx).Warnf(str, v...)
}

func (l *Logger) Error(ctx context.Context, str string, v ...interface{}) {
	l.entry.WithContext(ctx).Errorf(str, v...)
}

func (l *Logger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	entry := l.entry.WithContext(ctx).WithField("elapsed", elapsed)
	switch {
	case err != nil && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		entry.WithError(err).WithFields(logrus.Fields{
			"sql":  sql,
			"rows": rows,
		}).Error("query failed")
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0:
		sql, rows := fc()
		entry.WithFields(logrus.Fields{
			"sql":       sql,
			"rows":      rows,
			"threshold": l.SlowThreshold,
		}).Warn("query exceeded slow threshold")
	default:
		sql, rows := fc()
		entry.WithFields(logrus.Fields{
			"sql":  sql,
			"rows": rows,
		}).Trace("query succeeded")
	}
}
