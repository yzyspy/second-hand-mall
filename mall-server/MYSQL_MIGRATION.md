# 本地 SQLite 切换 MySQL

本工具用于尚未上线、数据量小的项目。停写后一次性迁移，不做双写或持续同步。

## 配置和代码

- `examples/mysql.yaml` 是无密码示例；实际配置放在 Git 忽略的 `configs/` 中。
- 在原配置保留日志、微信和对象存储配置，增加 `mysql` 段，将 `gorm.db_type` 改为 `mysql`。
- `mysql.password` 支持直接读取本地配置。若设置了 `MYSQL_PASSWORD`，它会覆盖文件密码。
- `gorm.table_prefix` 必须为空，现有业务 SQL 使用固定表名。
- `gorm.enable_auto_migrate` 现在真正控制启动时建表。迁移命令独立建表，日常可设为 false。
- MySQL 使用 InnoDB、utf8mb4、区分大小写的 utf8mb4_bin；时间为 DATETIME(6)。

以下命令均在 `mall-server/` 下执行。

## 迁移

1. 编译并检查源数据和目标库（只读）：

```sh
go build -o /tmp/mall-db-migrate ./cmd/migrate-db
/tmp/mall-db-migrate --check --source ./second-hand-mall.db
/tmp/mall-db-migrate --check-target --config configs/mysql.local.yaml
```

2. 停止 IDE、终端等所有写入源 SQLite 或目标 MySQL 的服务。退出微信开发者工具中的自动请求，或保持后端停止。目标库必须专用、无业务数据。

3. 创建备份目录并迁移。备份路径必须是新文件：

```sh
mkdir -p backups
/tmp/mall-db-migrate --apply --confirm-stopped \
  --source ./second-hand-mall.db \
  --backup ./backups/before-mysql.db \
  --config configs/mysql.local.yaml
```

目标数据库不存在时，可额外加 `--create-database`，需要 CREATE DATABASE 权限。默认不会自动创建数据库。

命令通过 SQLite `VACUUM INTO` 创建一致性快照（0600 权限），然后从快照导入。不会修改原数据库。六张表按关联顺序导入，保留 ID、密码哈希、零值、时间和软删除行。所有记录在同一个 MySQL 事务中插入并逐行比较，成功后提交；非空目标库、非 InnoDB 表、超长文本、无效时间、超范围金额和缺失引用会中止。

时间按同一时刻归一化为 UTC 微秒；SQLite 的纳秒尾数会截断到 MySQL 的微秒精度。原始纳秒数据仍保留在 SQLite 原文件和备份中。DDL 建表不属于导入事务；失败后可能留下空表，修正原因后可重新导入。已成功导入的库再次运行会被拒绝。

4. 验证成功后，保留原配置副本，把 MySQL 配置作为服务启动配置：

```sh
go run . web -config configs/mysql.local.yaml
```

或者把经过验证的本地 MySQL 配置复制到 `configs/config.yaml`，沿用 IDE 的原启动参数。此目录已被 Git 忽略，不要强制提交含密码的配置或数据库备份。

## 验收与回滚

- 检查登录、商品列表/详情、发布、收藏、聊天和后台登录。
- 导入工具已比较全部字段；再测试新增用户/商品，确认自增 ID 正常。
- 确认所有旧服务已退出，避免继续写 SQLite。
- 若未产生需要保留的 MySQL 新数据，停止新服务，恢复 SQLite 配置即可；原 SQLite 文件保持不变。
- 若 MySQL 已新增或修改业务数据，切回旧 SQLite 会丢失这些变更，需要先同步。不要直接切回。
- 图片 URL 原样迁移；历史 HTTP 图片链接问题需独立处理。

## 测试

```sh
go test ./...
MYSQL_TEST_CONFIG="$PWD/configs/mysql.local.yaml" go test ./internal/dbmigration -run TestMySQLMigrationIntegration -v
```

集成测试会创建并删除名为 `mall_migration_test_<随机时间>` 的独立数据库，需要建库/删库权限，不会写入配置中的业务库。覆盖微秒时间、emoji、软删除、自增、聊天聚合、非空库保护、MyISAM 拒绝及实际 MySQL 回滚。
