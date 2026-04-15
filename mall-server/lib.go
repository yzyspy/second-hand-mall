package main

/**
定义这个文件的目的是在执行 go mod tidy 时，保留在go.mod文件中的这些依赖
*/
import (
	_ "github.com/gin-gonic/gin" //web框架
	_ "github.com/koding/multiconfig"
	_ "github.com/sirupsen/logrus" //日志库
	_ "github.com/urfave/cli/v2"
	_ "gorm.io/driver/sqlite" // 需要保留的依赖
	_ "gorm.io/gorm"          // 需要保留的依赖
)
