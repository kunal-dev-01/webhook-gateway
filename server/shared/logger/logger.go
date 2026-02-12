package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

/* ───────────────── Context Keys ───────────────── */

type ContextKey string

const (
	RequestIDKey ContextKey = "request_id"
	UserIDKey    ContextKey = "user_id"
	CompanyIDKey ContextKey = "company_id"
	ServiceKey   ContextKey = "service"
)

/* ───────────────── Config ───────────────── */

type Config struct {
	Environment      string
	ServiceName      string
	Level            zerolog.Level
	EnableConsole    bool
	EnableFile       bool
	FilePath         string
	MaxSize          int
	MaxBackups       int
	MaxAge           int
	Compress         bool
	EnableCaller     bool
	EnableStackTrace bool
	LogQueryParams   bool
	TimeFormat       string
}

func (c *Config) setDefaults() {
	if c.ServiceName == "" {
		c.ServiceName = "unknown-service"
	}
	if c.TimeFormat == "" {
		c.TimeFormat = time.RFC3339
	}
	if c.Environment == "" {
		c.Environment = "development"
	}
	if c.MaxSize == 0 {
		c.MaxSize = 100
	}
	if c.MaxBackups == 0 {
		c.MaxBackups = 5
	}
	if c.MaxAge == 0 {
		c.MaxAge = 30
	}
}

/* ───────────────── Logger ───────────────── */

type Logger struct {
	logger zerolog.Logger
	config Config
	mu     sync.RWMutex
}

var (
	globalLogger *Logger
	initOnce     sync.Once
)

/* ───────────────── Initialization ───────────────── */

func New(cfg Config) *Logger {
	cfg.setDefaults()
	zerolog.TimeFieldFormat = cfg.TimeFormat

	writers := buildWriters(cfg)
	output := combineWriters(writers)

	zlog := zerolog.New(output).
		Level(cfg.Level).
		With().
		Timestamp().
		Str("service", cfg.ServiceName).
		Logger()

	if cfg.EnableCaller {
		zlog = zlog.With().Caller().Logger()
	}

	return &Logger{
		logger: zlog,
		config: cfg,
	}
}

func buildWriters(cfg Config) []io.Writer {
	var writers []io.Writer

	if cfg.EnableConsole {
		if cfg.Environment == "development" {
			writers = append(writers, zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: time.RFC3339,
			})
		} else {
			writers = append(writers, os.Stdout)
		}
	}

	if cfg.EnableFile && cfg.FilePath != "" {
		dir := filepath.Dir(cfg.FilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "failed to create log directory: %v, using stdout\n", err)
			return []io.Writer{os.Stdout}
		}

		fileWriter := &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
			LocalTime:  true,
		}
		writers = append(writers, fileWriter)
	}

	return writers
}

func combineWriters(writers []io.Writer) io.Writer {
	switch len(writers) {
	case 0:
		return os.Stdout
	case 1:
		return writers[0]
	default:
		return zerolog.MultiLevelWriter(writers...)
	}
}

func Init(cfg Config) {
	initOnce.Do(func() {
		globalLogger = New(cfg)
		zerolog.DefaultContextLogger = &globalLogger.logger
	})
}

func Get() *Logger {
	if globalLogger == nil {
		return New(Config{
			Environment:   "development",
			EnableConsole: true,
			Level:         zerolog.InfoLevel,
		})
	}
	return globalLogger
}

func (l *Logger) Close() error {
	return nil
}

/* ───────────────── Context Extraction ───────────────── */

type contextExtractor struct {
	logger zerolog.Logger
}

func (ce *contextExtractor) extractFromContext(ctx context.Context) zerolog.Logger {
	logger := ce.logger

	if reqID := ctx.Value(RequestIDKey); reqID != nil {
		if id, ok := reqID.(string); ok && id != "" {
			logger = logger.With().Str("request_id", id).Logger()
		}
	}
	if userID := ctx.Value(UserIDKey); userID != nil {
		if id, ok := userID.(uuid.UUID); ok && id != uuid.Nil {
			logger = logger.With().Str("user_id", id.String()).Logger()
		}
	}
	if companyID := ctx.Value(CompanyIDKey); companyID != nil {
		if id, ok := companyID.(uuid.UUID); ok && id != uuid.Nil {
			logger = logger.With().Str("company_id", id.String()).Logger()
		}
	}

	return logger
}

func (ce *contextExtractor) extractFromFiber(c *fiber.Ctx) zerolog.Logger {
	logger := ce.logger

	extractLocal := func(key string, setter func(string)) {
		if val := c.Locals(key); val != nil {
			switch v := val.(type) {
			case string:
				if v != "" {
					setter(v)
				}
			case uuid.UUID:
				if v != uuid.Nil {
					setter(v.String())
				}
			}
		}
	}

	extractLocal("requestid", func(s string) { logger = logger.With().Str("request_id", s).Logger() })
	extractLocal("user_id", func(s string) { logger = logger.With().Str("user_id", s).Logger() })
	extractLocal("company_id", func(s string) { logger = logger.With().Str("company_id", s).Logger() })

	logger = logger.With().
		Str("method", c.Method()).
		Str("path", c.Path()).
		Str("ip", c.IP()).
		Logger()

	return logger
}

/* ───────────────── Context-Aware Logging ───────────────── */

func (l *Logger) WithContext(ctx context.Context) *zerolog.Logger {
	ce := &contextExtractor{logger: l.logger}
	result := ce.extractFromContext(ctx)
	return &result
}

func (l *Logger) WithFiberContext(c *fiber.Ctx) *zerolog.Logger {
	ce := &contextExtractor{logger: l.logger}
	result := ce.extractFromFiber(c)
	return &result
}

/* ───────────────── Structured Logging ───────────────── */

type EventBuilder func() *zerolog.Event

func (l *Logger) log(eventFn EventBuilder, msg string, err error, fields ...any) {
	event := eventFn()
	if err != nil {
		event = event.Err(err)
	}
	if len(fields) > 0 {
		l.addFields(event, fields...)
	}
	event.Msg(msg)
}

func (l *Logger) Info(msg string, fields ...any)  { l.log(l.logger.Info, msg, nil, fields...) }
func (l *Logger) Debug(msg string, fields ...any) { l.log(l.logger.Debug, msg, nil, fields...) }
func (l *Logger) Warn(msg string, fields ...any)  { l.log(l.logger.Warn, msg, nil, fields...) }

func (l *Logger) Error(msg string, err error, fields ...any) {
	l.log(l.logger.Error, msg, err, fields...)
}

func (l *Logger) Fatal(msg string, err error, fields ...any) {
	l.log(l.logger.Fatal, msg, err, fields...)
}

/* ───────────────── Specialized Logging ───────────────── */

func (l *Logger) HTTP(method, path string, status int, duration time.Duration, fields ...any) {
	level := l.logger.Info
	if status >= 500 {
		level = l.logger.Error
	} else if status >= 400 {
		level = l.logger.Warn
	}

	event := level().
		Str("type", "http").
		Str("method", method).
		Str("path", path).
		Int("status", status).
		Dur("duration_ms", duration)

	l.addFields(event, fields...)
	event.Msg("HTTP request completed")
}

func (l *Logger) Database(operation string, duration time.Duration, err error, fields ...any) {
	level := l.logger.Debug
	msg := "Database operation completed"
	if err != nil {
		level = l.logger.Error
		msg = "Database operation failed"
	}

	event := level().Err(err).
		Str("type", "database").
		Str("operation", operation).
		Dur("duration_ms", duration)

	l.addFields(event, fields...)
	event.Msg(msg)
}

func (l *Logger) Auth(event string, userID uuid.UUID, success bool, fields ...any) {
	level := l.logger.Info
	if !success {
		level = l.logger.Warn
	}

	logEvent := level().
		Str("type", "auth").
		Str("event", event).
		Str("user_id", userID.String()).
		Bool("success", success)

	l.addFields(logEvent, fields...)
	logEvent.Msg("Authentication event")
}

func (l *Logger) Security(event, severity string, fields ...any) {
	var level func() *zerolog.Event

	switch severity {
	case "critical":
		level = l.logger.Error
	case "high":
		level = l.logger.Warn
	default:
		level = l.logger.Info
	}

	logEvent := level().
		Str("type", "security").
		Str("event", event).
		Str("severity", severity)

	l.addFields(logEvent, fields...)
	logEvent.Msg("Security event")
}

func (l *Logger) Audit(action string, userID uuid.UUID, resource string, fields ...any) {
	event := l.logger.Info().
		Str("type", "audit").
		Str("action", action).
		Str("user_id", userID.String()).
		Str("resource", resource)

	l.addFields(event, fields...)
	event.Msg("Audit event")
}

/* ───────────────── Helper ───────────────── */

func (l *Logger) addFields(event *zerolog.Event, fields ...any) {
	if len(fields) == 0 {
		return
	}

	if len(fields)%2 != 0 {
		fields = append(fields, "MISSING_VALUE")
	}

	for i := 0; i < len(fields); i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			event.Interface(fmt.Sprintf("field_%d", i), fields[i])
			event.Interface(fmt.Sprintf("value_%d", i+1), fields[i+1])
			continue
		}

		switch v := fields[i+1].(type) {
		case string:
			event.Str(key, v)
		case int:
			event.Int(key, v)
		case int32:
			event.Int32(key, v)
		case int64:
			event.Int64(key, v)
		case uint:
			event.Uint(key, v)
		case uint32:
			event.Uint32(key, v)
		case uint64:
			event.Uint64(key, v)
		case float32:
			event.Float32(key, v)
		case float64:
			event.Float64(key, v)
		case bool:
			event.Bool(key, v)
		case time.Duration:
			event.Dur(key, v)
		case time.Time:
			event.Time(key, v)
		case error:
			event.AnErr(key, v)
		case uuid.UUID:
			event.Str(key, v.String())
		default:
			event.Interface(key, v)
		}
	}
}

func (l *Logger) Logger() *zerolog.Logger {
	return &l.logger
}

/* ───────────────── Global Functions ───────────────── */

func Info(msg string, fields ...any)                 { Get().Info(msg, fields...) }
func Debug(msg string, fields ...any)               { Get().Debug(msg, fields...) }
func Warn(msg string, fields ...any)                { Get().Warn(msg, fields...) }
func Error(msg string, err error, fields ...any)    { Get().Error(msg, err, fields...) }
func WithContext(ctx context.Context) *zerolog.Logger { return Get().WithContext(ctx) }
func WithFiberContext(c *fiber.Ctx) *zerolog.Logger   { return Get().WithFiberContext(c) }