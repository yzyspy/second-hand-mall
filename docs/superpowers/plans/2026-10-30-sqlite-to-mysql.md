# SQLite to MySQL Implementation Plan

**Goal:** 支持配置切换 MySQL，并提供停服后执行的一次性保留 ID 的数据迁移。
**Architecture:** 保留 SQLite 测试支持；集中连接和建表；独立迁移命令使用 SQLite 一致性备份，校验后在空 MySQL 库内事务导入。
**Tech Stack:** Go、GORM、go-sql-driver/mysql、SQLite、MySQL 8。

## Constraints
- 不修改原 SQLite 数据，不自动切换未配置的运行环境。
- 不修改用户其他功能；密码通过 MYSQL_PASSWORD 覆盖配置。
- 保留六张表的 ID、密码哈希、软删除及关联；MySQL 时间精度为微秒。
- 目标表非空立即拒绝；导入后逐行校验，在提交前失败则回滚。

## Tasks
- [x] 连接：config.go 增加 MySQL；models/database.go 构建连接；gormx.go 实现驱动和未知类型错误；单测验证 DSN、环境变量和自动迁移开关。
- [x] 结构：独立 MigrateSchema，修正商品 TEXT 默认值；连接池读取配置，保留原表名。
- [x] 迁移：internal/dbmigration 检查六表、字段长度、金额和引用；cmd/migrate-db 提供 --check、--apply、--backup，实际导入前生成一致性备份；测试空库保护、软删除、零值和事务回滚。
- [x] 验证：go test ./...；使用独立 MySQL 实例执行集成测试；原 SQLite 保留；用户提供本机连接后已停服备份、实际迁移并切换服务。
- [x] 文档：配置示例、停服/备份/导入/切换步骤、验收命令及切回 SQLite 的限制。

## Completed verification
- go test ./... 通过；真实 MySQL 9.4 集成演练通过。
- second-mall 六表 21 行导入并逐行校验通过。
- 原 SQLite Go 服务已停止，新服务使用 MySQL 配置运行在 8080。
- 健康检查、7 个在售商品列表与原库一致、对应商品详情接口通过。
- 备份与本地密码配置均由 Git 忽略，未提交代码。
