package logs

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	Log *zap.Logger
	//cfg zap.Config
	atomicLevel zap.AtomicLevel
)

func init() {
	hook := lumberjack.Logger{
		Filename:   "./logs/eshop.log", // 日志文件路径
		MaxSize:    128,                // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: 30,                 // 日志文件最多保存多少个备份
		MaxAge:     7,                  // 文件最多保存多少天
		Compress:   true,               // 是否压缩
	}

	/* 	cfg = zap.Config{
		Encoding:    "json",
		Level:       zap.NewAtomicLevelAt(zapcore.DebugLevel),
		OutputPaths: []string{"stdout", "./eshop.log"},
		//ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:      "time",
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			LevelKey:     "level",
			EncodeLevel:  zapcore.CapitalColorLevelEncoder,
			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
			MessageKey:   "message",
		},
	} */
	cfg := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "func",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder, // 有颜色编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,       // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,   //
		EncodeCaller:   zapcore.FullCallerEncoder,        // 全路径编码器
		EncodeName:     zapcore.FullNameEncoder,
	}
	cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	atomicLevel = zap.NewAtomicLevel()
	atomicLevel.SetLevel(zap.DebugLevel)

	core := zapcore.NewCore(
		//zapcore.NewJSONEncoder(cfg), // 编码器配置
		zapcore.NewConsoleEncoder(cfg),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&hook)), // 打印到控制台和文件
		atomicLevel, // 日志级别
	)

	// 开启开发模式，堆栈跟踪
	caller := zap.AddCaller()

	// 开启文件及行号
	development := zap.Development()
	// 设置初始化字段
	//filed := zap.Fields(zap.String("serviceName", "serviceName"))
	// 构造日志
	Log = zap.New(core, caller, development)

	//Log, _ = cfg.Build()

	//fmt.Println("Logger system initialized!")

}

//zapcore.Levels ("debug", "info", "warn", "error", "dpanic", "panic", and "fatal").
func SetLevel(level string) {
	atomicLevel.UnmarshalText([]byte(level))
	atomicLevel.SetLevel((atomicLevel.Level()))
}
