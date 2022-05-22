package xmysql

import (
	"fmt"
	"github.com/smallfish-root/common-pkg/xjson"
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"sync"
)

var (
	newMysqlPools = make(map[string]*gorm.DB)
	once          sync.Once
)

func InitMySql(configs []*MySqlPoolConfig) {
	once.Do(func() {
		for _, c := range configs {
			if _, ok := mysqlPools[c.Alias]; ok {
				panic(errors.New("duplicate mysql pool: " + c.Alias))
			}
			p, err := createNewMySqlPool(c)
			if err != nil {
				panic(errors.New(fmt.Sprintf("mysql pool %s error %v", xjson.SafeMarshal(c), err)))
			}
			newMysqlPools[c.Alias] = p
		}
	})
}

func createNewMySqlPool(c *MySqlPoolConfig) (*gorm.DB, error) {
	cfg := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
		Logger:         logger.Default.LogMode(logger.LogLevel(c.LogMode)),
	}
	db, err := gorm.Open(mysql.Open(c.Address), cfg)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	sqlDB.SetMaxIdleConns(c.MaxIdleConn)
	sqlDB.SetMaxOpenConns(c.MaxOpenConn)
	if c.MaxLifeTime != 0 {
		sqlDB.SetConnMaxLifetime(c.MaxLifeTime)
	}
	if c.MaxIdleTime != 0 {
		sqlDB.SetConnMaxIdleTime(c.MaxIdleTime)
	}

	if err = sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, errors.WithStack(err)
	}

	if db == nil {
		return nil, errors.New("db is nil")
	}
	return db, nil
}

func GetMySqlConn(alias string) *gorm.DB {
	return newMysqlPools[alias]
}

func CloseMysql() {
	for _, db := range newMysqlPools {
		if db == nil {
			continue
		}
		sqlDB, err := db.DB()
		if err != nil {
			continue
		}
		_ = sqlDB.Close()
	}
}
