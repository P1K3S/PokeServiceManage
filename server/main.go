package main

import (
	"fmt"
	"log"

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
	); err != nil {
		logger.Log.Sugar().Fatalf("自动建表失败: %v", err)
	}

	cleanupDatabase(db)
	seedAdminUser(db)
	backfillExistingData(db)

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
	db.Model(&model.User{}).Where("username = ?", "poke").Count(&count)
	if count == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("shiwan233"), bcrypt.DefaultCost)
		db.Create(&model.User{
			Username: "poke",
			Password: string(hash),
			Role:     model.RoleSuperAdmin,
			Status:   1,
		})
		logger.Log.Info("默认超级管理员账号已创建: poke / shiwan233")
	}
}

func backfillExistingData(db *gorm.DB) {
	var admin model.User
	if err := db.Where("username = ?", "poke").First(&admin).Error; err != nil {
		return
	}

	tables := []string{"machines", "docker_services", "other_services", "egress_methods"}
	for _, table := range tables {
		db.Exec(fmt.Sprintf("UPDATE %s SET user_id = ? WHERE user_id = 0 OR user_id IS NULL", table), admin.ID)
	}
}
