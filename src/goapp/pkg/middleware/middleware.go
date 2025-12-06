//go:build !windows

package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type ZapConf struct {
	Level       string
	Path        string
	Format      string
	Prefix      string
	EncodeLevel string
}

type LogrotateConf struct {
	MaxSize    int
	MaxBackups int
	MaxAges    int
	Compress   bool
}

var (
	level            zapcore.Level
	zapConf          ZapConf
	logrotateConf    LogrotateConf
	lumberJackLogger *lumberjack.Logger
)

func Zap(zConf ZapConf, lConf LogrotateConf) *zap.SugaredLogger {
	zapConf = zConf
	logrotateConf = lConf

	level = getLevel(zapConf.Level)

	return zap.New(getEncoderCore(), zap.Development(), zap.AddCaller()).Sugar()
}

func StopWithRotate() {
	// logger.Info("StopWithRotate ...")
	err := lumberJackLogger.Rotate()
	if err != nil {
		// logger.Error("StopWithRotate error", zap.Error(err))
	}
}

func getEncoderCore() (core zapcore.Core) {
	lumberJackLogger = &lumberjack.Logger{
		Filename:   zapConf.Path,
		MaxSize:    logrotateConf.MaxSize,
		MaxBackups: logrotateConf.MaxBackups,
		MaxAge:     logrotateConf.MaxAges,
		Compress:   logrotateConf.Compress,
	}

	return zapcore.NewTee(
		zapcore.NewCore(
			getEncoder(),
			zapcore.AddSync(lumberJackLogger),
			level,
		),
	)
}

func getEncoder() zapcore.Encoder {
	if zapConf.Format == "json" {
		return zapcore.NewJSONEncoder(getEncoderConfig())
	}
	return zapcore.NewConsoleEncoder(getEncoderConfig())
}

func getEncoderConfig() (config zapcore.EncoderConfig) {
	config = zapcore.EncoderConfig{
		MessageKey:    "msg",
		LevelKey:      "level",
		TimeKey:       "time",
		NameKey:       "logger",
		CallerKey:     "caller",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		//EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeTime: CustomTimeEncoder,
		//EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	switch zapConf.EncodeLevel {
	case "LowercaseLevelEncoder":
		config.EncodeLevel = zapcore.LowercaseLevelEncoder // 小写编码器(默认)
	case "LowercaseColorLevelEncoder":
		config.EncodeLevel = zapcore.LowercaseColorLevelEncoder // 小写编码器带颜色
	case "CapitalLevelEncoder":
		config.EncodeLevel = zapcore.CapitalLevelEncoder // 大写编码器
	case "CapitalColorLevelEncoder":
		config.EncodeLevel = zapcore.CapitalColorLevelEncoder // 大写编码器带颜色
	default:
		config.EncodeLevel = zapcore.LowercaseLevelEncoder
	}
	return config
}

func getLevel(level string) (lv zapcore.Level) {
	level = strings.ToLower(level)

	switch level {
	case "debug":
		lv = zap.DebugLevel
	case "info":
		lv = zap.InfoLevel
	case "warn":
		lv = zap.WarnLevel
	case "error":
		lv = zap.ErrorLevel
	case "dpanic":
		lv = zap.DPanicLevel
	case "panic":
		lv = zap.PanicLevel
	case "fatal":
		lv = zap.FatalLevel
	default:
		lv = zap.InfoLevel
	}
	return
}

// 自定义日志输出时间格式
func CustomTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(zapConf.Prefix + "2006-01-02 15:04:05"))
}

// 自定义日志级别显示
func CustomEncodeLevel(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + level.CapitalString() + "]")
}

func ViperParseConf(path ...string) (v *viper.Viper) {
	var config string
	if len(path) == 0 {
		config = "config.yaml"
	} else {
		config = path[0]
	}
	fmt.Printf("use config:%v\n", config)

	v = viper.New()
	//viper.AddConfigPath(".")
	v.SetConfigFile(config)
	err := v.ReadInConfig() // Find and read the config file
	if err != nil {         // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	v.WatchConfig()

	return v
}
