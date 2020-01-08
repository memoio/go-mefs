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
func StartLogger() {

	MLoglevel = zap.NewAtomicLevel()

	writerSyncer := getLogWriter()

	encoder := getEncoder()

	core := zapcore.NewCore(encoder, writerSyncer, MLoglevel)

	logger := zap.New(core, zap.AddCaller())

	MLogger = logger.Sugar()

	MLoglevel.SetLevel(zapcore.DebugLevel)

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
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	return zapcore.NewJSONEncoder(encoderConfig)
}

func getLogWriter() zapcore.WriteSyncer {
	root, err := mefsPath()
	if err != nil {
		root = "~/.mefs"
	}

	lumberJackLogger := &lumberjack.Logger{
		Filename:   root + "/mefs.log",
		MaxSize:    1024, //MB
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
