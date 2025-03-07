package logging

import (
	"context"
	"errors"
	"flag"
	"os"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// JSONFormat represents JSON logging mode.
	JSONFormat = "json"
	// TextFormat represents text logging mode.
	// Default logging mode is TextFormat.
	TextFormat = "text"
)

// Initially, globalLog comes from controller-runtime/log with logger created earlier by controller-runtime.
// When logging.Setup is called, globalLog is switched to the real logger.
// Call depth of all loggers created before logging.Setup will not work, including package level loggers as they are created before main.
// All loggers created after logging.Setup won't be subject to the call depth limitation and will work if the underlying sink supports it.
var globalLog = log.Log

func Init(flags *flag.FlagSet) {
	// clear flags initialized in static dependencies
	if flag.CommandLine.Lookup("log_dir") != nil {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}
	klog.InitFlags(flags)
}

// Setup configures the logger with the supplied log format.
// It returns an error if the JSON logger could not be initialized or passed logFormat is not recognized.
func Setup(logFormat string) error {
	switch logFormat {
	case TextFormat:
		// in text mode we use FormatSerialize format
		globalLog = klogr.New()
	case JSONFormat:
		zapLog, err := zap.NewProduction()
		if err != nil {
			return err
		}
		klog.SetLogger(zapr.NewLogger(zapLog))
		// in json mode we use FormatKlog format
		globalLog = klog.NewKlogr()
	default:
		return errors.New("log format not recognized, pass `text` for text mode or `json` to enable JSON logging")
	}
	log.SetLogger(globalLog)
	return nil
}

// GlobalLogger returns a logr.Logger as configured in main.
func GlobalLogger() logr.Logger {
	return globalLog
}

// WithName returns a new logr.Logger instance with the specified name element added to the Logger's name.
func WithName(name string) logr.Logger {
	return GlobalLogger().WithName(name)
}

// WithValues returns a new logr.Logger instance with additional key/value pairs.
func WithValues(keysAndValues ...interface{}) logr.Logger {
	return GlobalLogger().WithValues(keysAndValues...)
}

// V returns a new logr.Logger instance for a specific verbosity level.
func V(level int) logr.Logger {
	return GlobalLogger().V(level)
}

// Info logs a non-error message with the given key/value pairs.
func Info(msg string, keysAndValues ...interface{}) {
	GlobalLogger().WithCallDepth(1).Info(msg, keysAndValues...)
}

// Error logs an error, with the given message and key/value pairs.
func Error(err error, msg string, keysAndValues ...interface{}) {
	GlobalLogger().WithCallDepth(1).Error(err, msg, keysAndValues...)
}

// FromContext returns a logger with predefined values from a context.Context.
func FromContext(ctx context.Context, keysAndValues ...interface{}) (logr.Logger, error) {
	logger, err := logr.FromContext(ctx)
	if err != nil {
		return logger, err
	}
	return logger.WithValues(keysAndValues...), nil
}

// IntoContext takes a context and sets the logger as one of its values.
// Use FromContext function to retrieve the logger.
func IntoContext(ctx context.Context, log logr.Logger) context.Context {
	return logr.NewContext(ctx, log)
}

// IntoBackground calls IntoContext with the logger and a Background context.
func IntoBackground(log logr.Logger) context.Context {
	return IntoContext(context.Background(), log)
}

// IntoTODO calls IntoContext with the logger and a TODO context.
func IntoTODO(log logr.Logger) context.Context {
	return IntoContext(context.TODO(), log)
}

// Background calls IntoContext with the global logger and a Background context.
func Background() context.Context {
	return IntoContext(context.Background(), GlobalLogger())
}

// TODO calls IntoContext with the global logger and a TODO context.
func TODO() context.Context {
	return IntoContext(context.TODO(), GlobalLogger())
}
