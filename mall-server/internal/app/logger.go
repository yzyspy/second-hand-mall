package app

import (
	"fmt"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"mall-server/internal/app/config"
	"mall-server/pkg/logger"
	"os"
	"path/filepath"
	"time"
)

// InitLogger 初始化日志模块
func InitLogger() (func(), error) {
	c := config.C.Log
	logger.SetLevel(logger.Level(c.Level))
	logger.SetFormatter(c.Format)

	// 设定日志输出
	var logWriter *rotatelogs.RotateLogs
	if c.Output != "" {
		fmt.Printf("InitLogger1 %s", c.Output)
		switch c.Output {
		case "stderr":
			logger.SetOutput(os.Stderr)
		case "file":
			fmt.Printf("InitLogger2 %s\n", c.OutputFile)
			if name := c.OutputFile; name != "" {
				_ = os.MkdirAll(filepath.Dir(name), 0777)
				//创建日志目录 /data1/weibo
				if err := os.MkdirAll(filepath.Dir(name), 0777); err != nil {
					panic(err)
				}
				var err error
				logWriter, err = rotatelogs.New(
					filepath.Join(name, "logs", "vpn-log_%Y%m%d%H%M%S.log"), //日志路径
					rotatelogs.WithLinkName(filepath.Join(name, "logs", "vpn-log.log")),
					rotatelogs.WithMaxAge(24*time.Hour),      // 最大保留天数：7天
					rotatelogs.WithRotationTime(1*time.Hour), // 日志分割时间：1分钟
				)
				if err != nil {
					panic(fmt.Sprintf("failed to initialize rotatelogs: %v", err))
				}
				if _, err := logWriter.Write([]byte("Init log\n")); err != nil {
					panic(fmt.Sprintf("写入初始化日志失败: %v", err))
				}
				fmt.Printf("InitLogger3 %s \n", logWriter.CurrentFileName())
				if err != nil {
					panic(err)
				}
				logger.SetOutput(logWriter)
				logger.Errorf("日志系统初始化成功，当前文件：%s\n", logWriter.CurrentFileName())
			}
		}
	}

	return func() {
		if logWriter != nil {
			logWriter.Close()
		}
	}, nil
}
