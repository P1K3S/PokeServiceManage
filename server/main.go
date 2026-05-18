package main

import (
	"fmt"
	"log"
	"os"

	"service-manage/config"
	"service-manage/logger"
	"service-manage/model"
	"service-manage/router"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	if err := config.InitConfig(); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	cfg := config.AppConfig

	os.Setenv("GIN_MODE", cfg.Server.Mode)

	logger.InitLog(cfg.Log.Level, cfg.Log.Filename, cfg.Log.MaxSize, cfg.Log.MaxBackups, cfg.Log.MaxAge)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Dbname,
		cfg.Database.Charset,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Log.Sugar().Fatalf("连接数据库失败: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Log.Sugar().Fatalf("获取数据库实例失败: %v", err)
	}
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)

	if err := db.AutoMigrate(
		&model.User{},
		&model.Machine{},
		&model.DockerService{},
		&model.OtherService{},
		&model.EgressMethod{},
		&model.Notice{},
		&model.OperationLog{},
	); err != nil {
		logger.Log.Sugar().Fatalf("自动建表失败: %v", err)
	}

	cleanupDatabase(db)
	seedAdminUser(db)
	backfillExistingData(db)
	seedNotice(db)

	logger.Log.Info("数据库表初始化完成")

	r := router.SetupRouter(db)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Log.Sugar().Infof("服务启动于 %s", addr)
	if err := r.Run(addr); err != nil {
		logger.Log.Sugar().Fatalf("服务启动失败: %v", err)
	}
}

func cleanupDatabase(db *gorm.DB) {
	dropForeignKeyIfExists(db, "egress_methods", "fk_egress_methods_docker_service")
	dropForeignKeyIfExists(db, "egress_methods", "fk_egress_methods_egress_service")
	dropForeignKeyIfExists(db, "docker_services", "fk_docker_services_machine")
	dropForeignKeyIfExists(db, "other_services", "fk_other_services_machine")

	if db.Migrator().HasColumn(&model.EgressMethod{}, "method_type") {
		db.Migrator().DropColumn(&model.EgressMethod{}, "method_type")
	}
}

func dropForeignKeyIfExists(db *gorm.DB, table, fkName string) {
	var count int64
	db.Raw(
		"SELECT COUNT(*) FROM information_schema.TABLE_CONSTRAINTS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND CONSTRAINT_NAME = ? AND CONSTRAINT_TYPE = 'FOREIGN KEY'",
		table, fkName,
	).Scan(&count)
	if count > 0 {
		db.Exec(fmt.Sprintf("ALTER TABLE %s DROP FOREIGN KEY %s", table, fkName))
	}
}

func seedAdminUser(db *gorm.DB) {
	var count int64
	adminUser := config.AppConfig.Auth.AdminUsername
	adminPass := config.AppConfig.Auth.AdminPassword
	db.Model(&model.User{}).Where("username = ?", adminUser).Count(&count)
	if count == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
		db.Create(&model.User{
			Username: adminUser,
			Password: string(hash),
			Role:     model.RoleSuperAdmin,
			Status:   1,
		})
		logger.Log.Sugar().Infof("默认超级管理员账号已创建: %s", adminUser)
	}
}

func backfillExistingData(db *gorm.DB) {
	var admin model.User
	if err := db.Where("username = ?", config.AppConfig.Auth.AdminUsername).First(&admin).Error; err != nil {
		return
	}

	tables := []string{"machines", "docker_services", "other_services", "egress_methods"}
	for _, table := range tables {
		db.Exec(fmt.Sprintf("UPDATE %s SET user_id = ? WHERE user_id = 0 OR user_id IS NULL", table), admin.ID)
	}
}

func seedNotice(db *gorm.DB) {
	var count int64
	db.Model(&model.Notice{}).Where("status = 1").Count(&count)
	if count > 0 {
		return
	}
	db.Create(&model.Notice{
		Title:   "系统更新通知",
		Content: "## 【2026-05-18 更新】\n\n1. **通知公告升级**：支持 Markdown 语法渲染，支持新增、编辑、删除多条通知\n2. **SSH 终端优化**：切换侧边栏不再断开连接，终端状态完整保留\n3. **仪表盘统计修复**：\"运行中Docker服务\"现在只统计 Docker 服务，不再混入其他服务\n4. **导出安全加固**：SSH 密码导出时自动脱敏为 `******`，非管理员只能导出自己创建的数据\n5. **浏览器图标更新**：标签页图标更换为全新仪表盘图标",
		Status:  1,
	})
	db.Create(&model.Notice{
		Title:   "历史更新记录",
		Content: "## 【2026-05-17 更新】\n\n1. 所有配置抽象到 `config.yaml`，项目开箱即用\n2. FRP 配置全自动发现：选了出站服务后自动 inspect 容器、读取配置、解析端口和 token\n3. SSH 终端上线，浏览器直连主机\n4. 健康检查改为公网探测 + 并发超时\n5. 仪表盘通知公告替代最近操作，支持在线编辑\n6. Docker 服务连通检测改为并发执行\n7. 代理名称统一更名为隧道名称",
		Status:  1,
	})
}
