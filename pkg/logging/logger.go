package logging

import (
	"fmt"
	"maps"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/bobg/go-generics/v4/slices"
	"github.com/sirupsen/logrus"
)

type Fields logrus.Fields
type Logger struct {
	entry           *logrus.Entry
	name            string
	trackingMode    TrackingMode
	trackingMessage string
	start           time.Time
	success         bool
}

type TrackingMode int

const (
	TrackingModeDebug TrackingMode = iota
	TrackingModeTrace

	defaultTrackingMode = TrackingModeDebug

	NameKey        = "logger"
	StartKey       = "start"
	FinishKey      = "finish"
	DurationKey    = "duration"
	SuccessKey     = "success"
	ErrorKey       = "error"
	CommandNameKey = "command"

	// time flags support Nano while logrus defaults to RFC3339 so ensure
	// all times that are logged correspond to what flags support
	TimestampFormat = time.RFC3339Nano
)

var (
	defaultLogger *Logger
	loggerOnce    sync.Once
)

// Clones the current Logger, adds Fields, adds StartKey field with value of time.Now(), and returns the resulting Logger.
// Fields will be merged in first to last order of existing/provided/StartKey.
func (l *Logger) StartTracking(fields ...Fields) *Logger {
	start := time.Now()
	c := l.WithFields(mergeFields(fields, Fields{StartKey: start}))
	c.start = start
	c.logf(l.logrusLevel(), "%v %v", ColorStartText(), c.trackingMessage)
	return c
}

// Clones the current Logger, adds Fields, adds tracking Fields, and returns the resulting Logger.
// Fields will be merged in first to last order of existing/provided/tracking.
//
// Tracking Fields: SuccessKey/FinishKey/DurationKey/ErrorKey
//
// SuccessKey is true if err == nil and WithSuccess/WithSuccessMessage has been called previously.
func (l *Logger) FinishTracking(err error, fields ...Fields) *Logger {
	return l.FinishTrackingWithMessage(err, "", fields...)
}

// Clones the current Logger, appending additional message to existing message, adds Fields, adds tracking Fields, and returns the resulting Logger.
// Fields will be merged in first to last order of existing/provided/tracking.
//
// Tracking Fields: SuccessKey/FinishKey/DurationKey/ErrorKey
//
// SuccessKey is true if err == nil and WithSuccess/WithSuccessMessage has been called previously.
func (l *Logger) FinishTrackingWithMessage(err error, additionalMessage string, fields ...Fields) *Logger {
	// if success wasn't called or if it was but err is not nil, we consider it a failure
	// this handles panic situations without having to call recover everywhere since
	// a panic will cause success not to be called but err will likely still be nil
	// since function that paniced didn't "return" a value
	success := l.success && err == nil
	finish := time.Now()
	duration := finish.Sub(l.start)
	merged := mergeFields(fields, TrackingFields(l.start, finish, duration, success, err))
	c := l.WithTracking("", l.trackingMessage+additionalMessage, merged)
	c.logf(l.logrusLevel(), "%v %v in %v", ColorResultConditionText(success), c.trackingMessage, duration)
	return c
}

// Clones the current Logger, adds Fields, marks succesful, adds SuccessKey field with value of true, and returns the resulting Logger.
// Fields will be merged in first to last order of existing/provided/SuccessKey.
func (l *Logger) WithSuccess(fields ...Fields) *Logger {
	return l.WithSuccessMessage("", fields...)
}

// Clones the current Logger, appending additional message to existing message, adds Fields, marks succesful, adds SuccessKey field with value of true, and returns the resulting Logger.
// Fields will be merged in first to last order of existing/provided/SuccessKey.
func (l *Logger) WithSuccessMessage(additionalMessage string, fields ...Fields) *Logger {
	c := l.WithTracking("", l.trackingMessage+additionalMessage, mergeFields(fields, Fields{SuccessKey: true}))
	c.success = true
	return c
}

// Clones the current Logger, sets cloned loggers tracking mode to TrackingModeTrace, adds Fields, and returns the resulting Logger.
// Fields will be merged in first to last order of existing/provided.
func (l *Logger) WithTrackingTraceMode(fields ...Fields) *Logger {
	return l.WithTrackingMode(TrackingModeTrace, fields...)
}

// Clones the current Logger, sets cloned loggers tracking mode to TrackingModeDebug, adds Fields, and returns the resulting Logger.
// Fields will be merged in first to last order of existing/provided.
func (l *Logger) WithTrackingDebugMode(fields ...Fields) *Logger {
	return l.WithTrackingMode(TrackingModeDebug, fields...)
}

// Clones the current Logger, sets cloned loggers tracking mode, adds Fields, and returns the resulting Logger.
// Fields will be merged in first to last order of existing/provided.
func (l *Logger) WithTrackingMode(mode TrackingMode, fields ...Fields) *Logger {
	return l.With("", "", mode, fields...)
}

func (l *Logger) WithField(key string, value any) *Logger {
	return l.WithFields(Fields{key: value})
}

// Clones the current Logger, adds Fields, and returns the resulting Logger.
// Fields will be merged in first to last order of existing/provided.
func (l *Logger) WithFields(fields ...Fields) *Logger {
	return l.WithName("", fields...)
}

// Clones the current Logger, adds Fields, adds ErrorKey field, and returns the resulting Logger.
// Fields will be merged in first to last order of existing/provided/ErrorKey.
func (l *Logger) WithError(err error, fields ...Fields) *Logger {
	return l.WithName("", mergeFields(fields, Fields{ErrorKey: err}))
}

// Clones the current Logger, adds a new path segment to cloned loggers name joined by periods, adds Fields, updates NameKey field if name has changed, and returns resulting Logger.
// Fields will be merged in first to last order of existing/provided/NameKey.
func (l *Logger) WithName(name string, fields ...Fields) *Logger {
	return l.WithTracking(name, "", fields...)
}

// Clones the current Logger, adds a new path segment to cloned loggers name joined by periods, replaces cloned loggers tracking message unless empty string, adds Fields, updates NameKey field if name has changed, and returns resulting Logger.
// Fields will be merged in first to last order of existing/provided/NameKey.
func (l *Logger) WithTracking(name string, trackingMessage string, fields ...Fields) *Logger {
	return l.With(name, trackingMessage, l.trackingMode, fields...)
}

// Clones the current Logger, adds a new path segment to cloned loggers name joined by periods, replaces cloned loggers tracking message unless empty string, sets cloned loggers tracking mode to TrackingModeTrace, adds Fields, updates NameKey field if name has changed, and returns resulting Logger.
// Fields will be merged in first to last order of existing/provided/NameKey.
func (l *Logger) WithTraceTracking(name string, trackingMessage string, fields ...Fields) *Logger {
	return l.With(name, trackingMessage, TrackingModeTrace, fields...)
}

// Clones the current Logger, adds a new path segment to cloned loggers name joined by periods, replaces cloned loggers tracking message unless empty string, sets cloned loggers tracking mode to TrackingModeDebug, adds Fields, updates NameKey field if name has changed, and returns resulting Logger.
// Fields will be merged in first to last order of existing/provided/NameKey.
func (l *Logger) WithDebugTracking(name string, trackingMessage string, fields ...Fields) *Logger {
	return l.With(name, trackingMessage, TrackingModeDebug, fields...)
}

// Clones the current Logger, adds a new path segment to cloned loggers name joined by periods, replaces cloned loggers tracking message unless empty string, applies LoggerMode, adds Fields, updates NameKey field if name has changed, and returns resulting Logger.
// Fields will be merged in first to last order of existing/provided/NameKey.
func (l *Logger) With(name string, trackingMessage string, trackingMode TrackingMode, fields ...Fields) *Logger {
	if name == "" && trackingMessage == "" && l.trackingMode == trackingMode && len(fields) == 0 {
		return l
	}

	newFields := make(Fields, 1)
	if name == "" {
		name = l.name
	} else {
		if l.name != "" {
			name = strings.Join([]string{l.name, name}, ".")
		}
		newFields[NameKey] = name
	}

	if trackingMessage == "" {
		trackingMessage = l.trackingMessage
	}

	c := l.clone(mergeFields(fields, newFields))
	c.name = name
	c.trackingMessage = trackingMessage
	c.trackingMode = trackingMode
	return c

}

func (l *Logger) Info(message string) {
	l.log(logrus.InfoLevel, message)
}

func (l *Logger) Infof(format string, args ...any) {
	l.logf(logrus.InfoLevel, format, args...)
}

func (l *Logger) Debug(message string) {
	l.log(logrus.DebugLevel, message)
}

func (l *Logger) Debugf(format string, args ...any) {
	l.logf(logrus.DebugLevel, format, args...)
}

func (l *Logger) Trace(message string) {
	l.log(logrus.TraceLevel, message)
}

func (l *Logger) Tracef(format string, args ...any) {
	l.logf(logrus.TraceLevel, format, args...)
}

func (l *Logger) Warn(message string) {
	l.log(logrus.WarnLevel, message)
}

func (l *Logger) Warnf(format string, args ...any) {
	l.logf(logrus.WarnLevel, format, args...)
}

// Only use to log an error when continuing execution despite the error.  When returning an error,
// this function should NOT be used.
func (l *Logger) Error(message string) {
	l.log(logrus.ErrorLevel, message)
}

// Only use to log an error when continuing execution despite the error.  When returning an error,
// this function should NOT be used.
func (l *Logger) Errorf(format string, args ...any) {
	l.logf(logrus.ErrorLevel, format, args...)
}

func (l *Logger) DiagEnabled() bool {
	return diagEnabled()
}

// no-op if recovered is nil, otherwise logs a message of the combined formatted prefix,
// formatted recovered and stack trace and exits via os.exit(1)
//
// This should only be called by go routines in defer.  All other code should let panics bubble
// or return errors on panic.  It is simply a last ditch effort to write panics to the log file
// until the way logging works is improved (see https://github.com/skuid/skuid-cli/issues/174)
//
// TODO: When logging is improved to utilize stdout/err this can be eliminated.
func (l *Logger) FatalGoRoutineConditionf(recovered any, prefixFormat string, prefixArgs ...any) {
	if recovered == nil {
		return
	}

	l.FatalGoRoutineCondition(recovered, fmt.Sprintf(prefixFormat, prefixArgs...))
}

// no-op if recovered is nil, otherwise logs a message of the combined formatted prefix,
// formatted recovered and stack trace and exits via os.exit(1)
//
// This should only be called by go routines in defer.  All other code should let panics bubble
// or return errors on panic.  It is simply a last ditch effort to write panics to the log file
// until the way logging works is improved (see https://github.com/skuid/skuid-cli/issues/174)
//
// TODO: When logging is improved to utilize stdout/err this can be eliminated.
func (l *Logger) FatalGoRoutineCondition(recovered any, prefix string) {
	if recovered == nil {
		return
	}

	l.log(logrus.FatalLevel, ColorFailure.Text(FormatPanic(recovered, prefix)))
}

func (l *Logger) logrusLevel() logrus.Level {
	if l.trackingMode == TrackingModeTrace {
		return logrus.TraceLevel
	} else if l.trackingMode == TrackingModeDebug {
		return logrus.DebugLevel
	}

	// should not happen in production
	panic(fmt.Sprintf("unexpected tracking mode [%v]", l.trackingMode))
}

func (l *Logger) logf(level logrus.Level, format string, args ...any) {
	l.log(level, fmt.Sprintf(format, args...))
}

func (l *Logger) log(level logrus.Level, message string) {
	if level == logrus.FatalLevel {
		l.entry.Fatal(message)
	} else {
		l.entry.Log(level, message)
	}
}

func (l *Logger) clone(fields Fields) *Logger {
	c := *l

	if len(fields) == 0 {
		return &c
	}
	c.entry = l.entry.WithFields(logrus.Fields(fields))
	return &c
}

func FormatPanic(recovered any, prefixFormat string, prefixArgs ...any) string {
	if recovered == nil {
		return ""
	}

	if prefixFormat != "" {
		prefixFormat += " "
	}

	msgFmt := "%vfailed unexpectedly:\n%s\n\n%s"
	msgArgs := []any{fmt.Sprintf(prefixFormat, prefixArgs...), recovered, debug.Stack()}
	return fmt.Sprintf(msgFmt, msgArgs...)
}

func TrackingFields(start time.Time, finish time.Time, duration time.Duration, success bool, err error, fields ...Fields) Fields {
	return mergeFields(fields, Fields{
		StartKey:    start,
		FinishKey:   finish,
		DurationKey: duration,
		SuccessKey:  success,
		ErrorKey:    err,
	})
}

func FormatTime(t *time.Time) string {
	if t == nil {
		// use Sprint to ensure "nil" output is consistent
		return fmt.Sprint(t)
	}
	return t.Format(TimestampFormat)
}

func FormatDuration(d *time.Duration) string {
	return fmt.Sprint(d.Nanoseconds())
}

// Clones the default Logger, adds a new path segment to cloned loggers name joined by periods, adds Fields, updates NameKey field if name has changed, and returns resulting Logger.
// Fields will be merged in first to last order of existing/provided/NameKey.
func WithName(name string, fields ...Fields) *Logger {
	return WithTracking(name, "", fields...)
}

// Clones the default Logger, adds a new path segment to cloned loggers name joined by periods, replaces cloned loggers tracking message unless empty string, adds Fields, updates NameKey field if name has changed, and returns resulting Logger.
// Fields will be merged in first to last order of existing/provided/NameKey.
func WithTracking(name string, trackingMessage string, fields ...Fields) *Logger {
	return With(name, trackingMessage, defaultTrackingMode, fields...)
}

// Clones the default Logger, adds a new path segment to cloned loggers name joined by periods, replaces cloned loggers tracking message unless empty string, sets cloned loggers tracking mode to TrackingModeTrace, adds Fields, updates NameKey field if name has changed, and returns resulting Logger.
// Fields will be merged in first to last order of existing/provided/NameKey.
func WithTraceTracking(name string, trackingMessage string, fields ...Fields) *Logger {
	return With(name, trackingMessage, TrackingModeTrace, fields...)
}

// Clones the default Logger, adds a new path segment to cloned loggers name joined by periods, replaces cloned loggers tracking message unless empty string, sets cloned loggers tracking mode to TrackingModeDebug, adds Fields, updates NameKey field if name has changed, and returns resulting Logger.
// Fields will be merged in first to last order of existing/provided/NameKey.
func WithDebugTracking(name string, trackingMessage string, fields ...Fields) *Logger {
	return With(name, trackingMessage, TrackingModeDebug, fields...)
}

// Clones the default Logger, adds a new path segment to cloned loggers name joined by periods, replaces cloned loggers tracking message unless empty string, applies LoggerMode, adds Fields, updates NameKey field if name has changed, and returns resulting Logger.
// Fields will be merged in first to last order of existing/provided/NameKey.
func With(name string, trackingMessage string, trackingMode TrackingMode, fields ...Fields) *Logger {
	loggerOnce.Do(func() {
		entry := newLogrusLogger().WithFields(logrus.Fields(mergeFields(fields, Fields{
			NameKey: name,
		})))
		defaultLogger = &Logger{entry: entry, trackingMode: TrackingModeDebug}

	})
	return defaultLogger.With(name, trackingMessage, trackingMode, fields...)
}

func mergeFields(fields []Fields, additional ...Fields) Fields {
	// default capacity following https://github.com/sirupsen/logrus/blob/master/entry.go#L77
	f := make(Fields, (len(fields)+len(fields))*6)

	slices.Each(fields, func(tf Fields) {
		maps.Copy(f, tf)
	})

	slices.Each(additional, func(tf Fields) {
		maps.Copy(f, tf)
	})

	return f
}
