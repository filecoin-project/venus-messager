package mysql

import (
	"fmt"
	"time"

	"golang.org/x/xerrors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/ipfs-force-community/venus-messager/models/repo"
)

type MysqlRepo struct {
	*gorm.DB
}

func (d MysqlRepo) MessageRepo() repo.MessageRepo {
	return newMysqlMessageRepo(d)
}

func (d MysqlRepo) WalletRepo() repo.WalletRepo {
	return newMysqlWalletRepo(d)
}

func (d MysqlRepo) AddressRepo() repo.AddressRepo {
	return newMysqlAddressRepo(d)
}

func (d MysqlRepo) AutoMigrate() error {
	err := d.GetDb().AutoMigrate(mysqlMessage{})
	if err != nil {
		return err
	}

	if err := d.GetDb().AutoMigrate(mysqlAddress{}); err != nil {
		return err
	}

	return d.GetDb().AutoMigrate(mysqlWallet{})
}

func (d MysqlRepo) GetDb() *gorm.DB {
	return d.DB
}

func (d MysqlRepo) DbClose() error {
	return d.DbClose()
}

func OpenMysql(cfg *config.MySqlConfig) (repo.Repo, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=%t&loc=%s",
		cfg.User,
		cfg.Pass,
		cfg.Addr,
		cfg.Name,
		true,
		"Local")

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		//Logger: logger.Default.LogMode(logger.Info), // 日志配置
	})

	if err != nil {
		return nil, xerrors.Errorf("[db connection failed] Database name: %s %w", cfg.Name, err)
	}

	db.Set("gorm:table_options", "CHARSET=utf8mb4")

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池 用于设置最大打开的连接数，默认值为0表示不限制.设置最大的连接数，可以避免并发太高导致连接mysql出现too many connections的错误。
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConn)

	// 设置最大连接数 用于设置闲置的连接数.设置闲置的连接数则当开启的一个连接使用完成后可以放在池里等候下一次使用。
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConn)

	// 设置最大连接超时
	sqlDB.SetConnMaxLifetime(time.Minute * cfg.ConnMaxLifeTime)

	// 使用插件
	//db.Use(&TracePlugin{})
	return &MysqlRepo{
		db,
	}, nil
}
