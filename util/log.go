package util

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/33cn/chain33/common/log/log15"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	timeFormat   = "2006-01-02T15:04:05-0700"
	floatFormat  = 'f'
	errorKey     = "LOG15_ERROR"
	maxOutputLen = 1000
	trimLen      = maxOutputLen / 2
	OmitMessage  = "  (...... logger omit %d character ......)  "
)

var stringBufPool = sync.Pool{
	New: func() interface{} { return new(bytes.Buffer) },
}

// SetupLog Setup Log
func SetupLog(name string, logLevel string) {
	log15.Root().SetHandler(log15.MultiHandler(*ErrLogger(name), *NormalLogger(name, logLevel)))
}

// ErrLogger log error
func ErrLogger(name string) *log15.Handler {
	return addLogger("logs/"+name+".error", "error", 365)
}

// NormalLogger all log
func NormalLogger(name string, logLevel string) *log15.Handler {
	return addLogger("logs/"+name, logLevel, 30)
}

func addLogger(name string, logLevel string, keepDays int) *log15.Handler {
	rotateLogger := &lumberjack.Logger{
		Filename:   name + ".log",
		MaxBackups: int(100),
		MaxAge:     keepDays,
		LocalTime:  true,
		Compress:   true,
	}

	fileh := log15.LvlFilterHandler(
		getLevel(logLevel),
		log15.StreamHandler(rotateLogger, LogFmtFormat()),
	)
	return &fileh
}

func getLevel(lvlString string) log15.Lvl {
	lvl, err := log15.LvlFromString(lvlString)
	if err != nil {
		// 日志级别配置不正确时默认为error级别
		return log15.LvlError
	}
	return lvl
}

// LogFmtFormat copy from log15.LogfmtFormat(), add max output length process: CompressContent
func LogFmtFormat() log15.Format {
	return log15.FormatFunc(func(r *log15.Record) []byte {
		common := []interface{}{r.KeyNames.Time, r.Time, r.KeyNames.Lvl, r.Lvl, r.KeyNames.Msg, r.Msg}
		buf := &bytes.Buffer{}
		logFmt(buf, append(common, r.Ctx...), 0)
		return buf.Bytes()
	})
}

func logFmt(buf *bytes.Buffer, ctx []interface{}, color int) {
	for i := 0; i < len(ctx); i += 2 {
		if i != 0 {
			buf.WriteByte(' ')
		}

		k, ok := ctx[i].(string)
		v := formatLogFmtValue(ctx[i+1])
		if !ok {
			k, v = errorKey, formatLogFmtValue(k)
		}
		v = CompressContent(v)
		// XXX: we should probably check that all of your key bytes aren't invalid
		if color > 0 {
			_, _ = fmt.Fprintf(buf, "\x1b[%dm%s\x1b[0m=%s", color, k, v)
		} else {
			buf.WriteString(k)
			buf.WriteByte('=')
			buf.WriteString(v)
		}
	}

	buf.WriteByte('\n')
}

// CompressContent process max output length
func CompressContent(s string) string {
	if len(s) > maxOutputLen {
		return s[:trimLen] + fmt.Sprintf(OmitMessage, len(s)-maxOutputLen) + s[len(s)-trimLen:]
	}
	return s
}

// formatValue formats a value for serialization
func formatLogFmtValue(value interface{}) string {
	if value == nil {
		return "nil"
	}

	if t, ok := value.(time.Time); ok {
		// Performance optimization: No need for escaping since the provided
		// timeFormat doesn't have any escape characters, and escaping is
		// expensive.
		return t.Format(timeFormat)
	}
	value = formatShared(value)
	switch v := value.(type) {
	case bool:
		return strconv.FormatBool(v)
	case float32:
		return strconv.FormatFloat(float64(v), floatFormat, 3, 64)
	case float64:
		return strconv.FormatFloat(v, floatFormat, 3, 64)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", value)
	case string:
		return escapeString(v)
	default:
		return escapeString(fmt.Sprintf("%+v", value))
	}
}

func formatShared(value interface{}) (result interface{}) {
	defer func() {
		if err := recover(); err != nil {
			if v := reflect.ValueOf(value); v.Kind() == reflect.Ptr && v.IsNil() {
				result = "nil"
			} else {
				panic(err)
			}
		}
	}()

	switch v := value.(type) {
	case time.Time:
		return v.Format(timeFormat)

	case error:
		return v.Error()

	case fmt.Stringer:
		return v.String()

	default:
		return v
	}
}

func escapeString(s string) string {
	needsQuotes := false
	needsEscape := false
	for _, r := range s {
		if r <= ' ' || r == '=' || r == '"' {
			needsQuotes = true
		}
		if r == '\\' || r == '"' || r == '\n' || r == '\r' || r == '\t' {
			needsEscape = true
		}
	}
	if !needsEscape && !needsQuotes {
		return s
	}
	e := stringBufPool.Get().(*bytes.Buffer)
	e.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\\', '"':
			e.WriteByte('\\')
			e.WriteByte(byte(r))
		case '\n':
			e.WriteString("\\n")
		case '\r':
			e.WriteString("\\r")
		case '\t':
			e.WriteString("\\t")
		default:
			e.WriteRune(r)
		}
	}
	e.WriteByte('"')
	var ret string
	if needsQuotes {
		ret = e.String()
	} else {
		ret = string(e.Bytes()[1 : e.Len()-1])
	}
	e.Reset()
	stringBufPool.Put(e)
	return ret
}
