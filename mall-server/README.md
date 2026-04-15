这是二手交易平台服务端使用go + gorm + gin + sqlite实现


## 项目初始化
```shell
go mod init mall-server

go get gorm.io/driver/sqlite@v1.5.6
go get gorm.io/gorm@v1.30.0


```

```markdown
require (
	github.com/gin-gonic/gin v1.11.0
	github.com/koding/multiconfig v0.0.0-20171124222453-69c27309b2d7
	github.com/sirupsen/logrus v1.9.3
	github.com/urfave/cli/v2 v2.27.5
)
```
go mod tidy


