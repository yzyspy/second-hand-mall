package config

import (
	"fmt"
	"github.com/koding/multiconfig"
	"strings"
	"sync"
)

// import (
//
//	"fmt"
//	"gitlab.bee.to/v-project/vpn-server/internal/app/module/kafka"
//	"gitlab.bee.to/v-project/vpn-server/pkg/redisx"
//	"gitlab.bee.to/v-project/vpn-server/pkg/util/json"
//	"os"
//	"strings"
//	"sync"
//
//	"github.com/jealone/sli2zap"
//	"github.com/koding/multiconfig"
//
// )

var (
	//logConfig = Log{
	//	Level:         2,                                    // 信息级别
	//	Format:        "json",                               // JSON 格式
	//	Output:        "file",                               // 输出到标准输出
	//	OutputFile:    "/data1/weibo/",                      // 输出到文件 app.log
	//	EnableHook:    true,                                 // 启用钩子
	//	HookLevels:    []string{"error", "warning", "info"}, // 钩子触发的日志级别
	//	HookMaxThread: 5,                                    // 钩子处理的最大线程数
	//	HookMaxBuffer: 1000,                                 // 钩子处理的最大缓冲区大小
	//}

	// C 全局配置(需要先执行MustLoad，否则拿不到配置)
	C    = new(Config)
	once sync.Once
)

// MustLoad 加载配置
func MustLoad(fpaths ...string) {
	once.Do(func() {
		loaders := []multiconfig.Loader{
			&multiconfig.TagLoader{},
			&multiconfig.EnvironmentLoader{},
		}

		for _, fpath := range fpaths {
			if strings.HasSuffix(fpath, "toml") {
				loaders = append(loaders, &multiconfig.TOMLLoader{Path: fpath})
			}
			if strings.HasSuffix(fpath, "json") {
				loaders = append(loaders, &multiconfig.JSONLoader{Path: fpath})
			}
			if strings.HasSuffix(fpath, "yaml") {
				loaders = append(loaders, &multiconfig.YAMLLoader{Path: fpath})
			}
		}

		m := multiconfig.DefaultLoader{
			Loader:    multiconfig.MultiLoader(loaders...),
			Validator: multiconfig.MultiValidator(&multiconfig.RequiredValidator{}),
		}
		fmt.Printf("must load config before: %v\n", C.Log)
		m.MustLoad(C) // 这个代码会给C的 Log Log `yaml:"log"` 赋值吗？？？？ 是的，会从config.yaml中读取log配置。
		fmt.Printf("must load config after:  %v\n", C.Log)
	})
}

func init() {
	//	C.Log = logConfig
}

// // Config 配置参数
type Config struct {
	Log        Log        `yaml:"log"`
	Gorm       Gorm       `yaml:"gorm"`
	RunMode    string     `yaml:"run_mode"`
	VpnMySQL   MySQL      `yaml:"vpn_mysql"`
	Redis      Redis      `yaml:"redis"`
	SQLite     SQLite     `yaml:"sqlite"`
	Consul     Consul     `yaml:"consul"`
	GrpcClient GrpcClient `yaml:"grpc_client"`
}

// // IsDebugMode 是否是debug模式
func (c *Config) IsDebugMode() bool {
	return c.RunMode == "debug"
}

// // Log 日志配置参数
type Log struct {
	Level         int      `yaml:"level"`
	Format        string   `yaml:"format"`
	Output        string   `yaml:"output"`
	OutputFile    string   `yaml:"output_file"`
	EnableHook    bool     `yaml:"enable_hook"`
	HookLevels    []string `yaml:"hook_levels"`
	HookMaxThread int      `yaml:"hook_max_thread"`
	HookMaxBuffer int      `yaml:"hook_max_buffer"`
}

// Gorm gorm配置参数
type Gorm struct {
	Debug             bool   `yaml:"debug"`
	DBType            string `yaml:"db_type"`
	MaxLifetime       int    `yaml:"max_lifetime"`
	MaxOpenConns      int    `yaml:"max_open_conns"`
	MaxIdleConns      int    `yaml:"max_idle_conns"`
	TablePrefix       string `yaml:"table_prefix"`
	EnableAutoMigrate bool   `yaml:"enable_auto_migrate"`
	EnabledMonitor    bool   `yaml:"enabled_monitor"`
}

// MySQL mysql配置参数
type MySQL struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	User       string `yaml:"user"`
	Password   string `yaml:"password"`
	DBName     string `yaml:"db_name"`
	Parameters string `yaml:"parameters"`
}

type SQLite struct {
	FilePath   string `yaml:"file_path"`  // 数据库文件路径，如 "data.db" 或 ":memory:"
	Parameters string `yaml:"parameters"` // 可选参数，如 "cache=shared&mode=rwc"
}

type Redis struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type Consul struct {
	Schema string `yaml:"schema"`
	Host   string `yaml:"host"`
}

type GrpcClient struct {
	Jaco_rc_decision_grpc_name    string `yaml:"jaco-rc-decision-grpc-name"`
	Jaco_rc_decision_grpc_address string `yaml:"jaco-rc-decision-grpc-address"`
}
