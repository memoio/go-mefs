package utils

import (
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// MLogger is global
var MLogger *zap.SugaredLogger

// MLoglevel is zap log level
var MLoglevel zap.AtomicLevel

// StartLogger starts
func StartLoggerWithLevel() {

	MLoglevel = zap.NewAtomicLevel()

	writerSyncer := getLogWriter("mefs")

	encoder := getEncoder()

	core := zapcore.NewCore(encoder, writerSyncer, MLoglevel)

	logger := zap.New(core, zap.AddCaller())

	MLogger = logger.Sugar()

	MLoglevel.SetLevel(zapcore.DebugLevel)

	MLogger.Info("Mefs Logger init success")
}

// StartLogger starts
func StartLogger() {

	debugLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.InfoLevel
	})

	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.InfoLevel
	})

	warnLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.WarnLevel
	})

	// 获取 info、warn日志文件的io.Writer 抽象 getWriter() 在下方实现
	debugWriter := getLogWriter("debug")
	infoWriter := getLogWriter("info")
	warnWriter := getLogWriter("error")

	encoder := getEncoder()

	// 最后创建具体的Logger
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, debugWriter, debugLevel),
		zapcore.NewCore(encoder, infoWriter, infoLevel),
		zapcore.NewCore(encoder, warnWriter, warnLevel),
	)

	logger := zap.New(core, zap.AddCaller())

	MLogger = logger.Sugar()

	MLogger.Info("Mefs Logger init success")
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	return zapcore.NewJSONEncoder(encoderConfig)
}

func getLogWriter(filename string) zapcore.WriteSyncer {
	root, err := mefsPath()
	if err != nil {
		root = "~/.mefs"
	}

	lumberJackLogger := &lumberjack.Logger{
		Filename:   root + "/logs/" + filename + ".log",
		MaxSize:    100, //MB
		MaxBackups: 3,
		MaxAge:     30, //days
		Compress:   false,
	}
	return zapcore.AddSync(lumberJackLogger)
}

func mefsPath() (string, error) {
	mefsPath := "~/.mefs"
	if os.Getenv("MEFS_PATH") != "" { //获取环境变量
		mefsPath = os.Getenv("MEFS_PATH")
	}
	mefsPath, err := homedir.Expand(mefsPath)
	if err != nil {
		return "", err
	}
	return mefsPath, nil
}
