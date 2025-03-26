package utils

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
)

// FileLineHook 新增一个字段用来打印文件路径及行号
type FileLineHook struct {
	LogLevels []logrus.Level // 需要打印的日志级别
	FieldName string         // 字段名称
	Skip      int            // 跳过几层调用栈
	Num       int            // Skip后的查找范围
	Test      bool           // 打印所有调用栈信息，找出合适的 Skip 配置
	filename  string         // 文件名
	line      int            // 行号
}

func (e *FileLineHook) Levels() []logrus.Level {
	return e.LogLevels
}

func (e *FileLineHook) Fire(entry *logrus.Entry) error {
	for i := 0; i < e.Num; i++ {
		_, e.filename, e.line, _ = runtime.Caller(e.Skip + i)
		if !strings.Contains(e.filename, "logrus") {
			break
		}
	}
	entry.Data[e.FieldName] = fmt.Sprintf("%s:%d", e.filename, e.line)
	if e.Test {
		buf := [4096]byte{}
		n := runtime.Stack(buf[:], false)
		fmt.Println(string(buf[:n]))
	}
	return nil
}

// NewFileLineHook 打印文件路径及行号
// levels 指定日志级别
func NewFileLineHook(levels ...logrus.Level) logrus.Hook {
	return &FileLineHook{
		LogLevels: levels,
		FieldName: "source",
		Skip:      8,
		Num:       2,
	}
}

type RotateLogConfig struct {
	Levels        []string `json:"levels"`
	Pattern       string   `json:"pattern"`
	LinkName      string   `json:"link_name"`
	MaxAge        int      `json:"max_age"`
	RotationTime  int      `json:"rotation_time"`
	RotationCount int      `json:"rotation_count"`
	RotationSize  int      `json:"rotation_size"`
}

func NewRotateLogHook(config *RotateLogConfig) logrus.Hook {
	var levels []logrus.Level
	for _, v := range config.Levels {
		level, err := logrus.ParseLevel(v)
		if err != nil {
			panic(err)
		}
		levels = append(levels, level)
	}

	if len(levels) == 0 {
		levels = logrus.AllLevels
	}

	l, err := rotatelogs.New(config.Pattern,
		rotatelogs.WithLinkName(config.LinkName),
		rotatelogs.WithMaxAge(time.Duration(config.MaxAge)*time.Hour),
		rotatelogs.WithRotationTime(time.Duration(config.RotationTime)*time.Hour),
		rotatelogs.WithRotationCount(uint(config.RotationCount)),
		rotatelogs.WithRotationSize(int64(config.RotationSize)),
	)
	if err != nil {
		panic(err)
	}

	return &writer.Hook{
		Writer:    l,
		LogLevels: levels,
	}
}
