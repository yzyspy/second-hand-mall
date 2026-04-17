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

增加启动参数, 参见 newWebCmd 方法
```shell
./mall-server web -config configs/config.yaml
```

## 数据库设计

启动redis,并开启redis的持久话功能
```shell
brew services restart redis
redis-cli -h 127.0.0.1 -p 6379
```



```sql
-- 系统用户表
CREATE TABLE IF NOT EXISTS "sys_users" (
    "id" INTEGER PRIMARY KEY AUTOINCREMENT,      -- 主键ID
    "created_at" DATETIME,                       -- 创建时间
    "updated_at" DATETIME,                       -- 更新时间
    "deleted_at" DATETIME,                       -- 删除时间（软删除标记）
    "username" TEXT NOT NULL DEFAULT '',         -- 用户名
    "password" TEXT NOT NULL DEFAULT '',         -- 密码
    "phone" TEXT NOT NULL DEFAULT '',            -- 手机号
    "wx_userid" TEXT NOT NULL DEFAULT '',        -- 微信用户ID
    "wx_openid" TEXT NOT NULL DEFAULT '',        -- 微信开放ID
    "avatar" TEXT NOT NULL DEFAULT '',           -- 头像地址
    "sex" TEXT NOT NULL DEFAULT '',              -- 性别
    "email" TEXT NOT NULL DEFAULT '',            -- 邮箱
    "remarks" TEXT NOT NULL DEFAULT '',          -- 备注信息
    "role_id" INTEGER NOT NULL DEFAULT 0         -- 角色ID
);

-- 为软删除字段创建索引，提升查询效率
CREATE INDEX idx_sys_users_deleted_at ON sys_users(deleted_at);
```

