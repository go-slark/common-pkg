package xmysql

import (
	"fmt"
	"github.com/nickxb/pkg/xjson"
	"github.com/nickxb/pkg/xsync"
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"sync"
)

var (
	newMysqlPools = make(map[string]*gorm.DB)
	mysqlLock     = &sync.RWMutex{}
)

func InitMySql(configs []*MySqlPoolConfig) error {
	for _, c := range configs {
		if _, ok := mysqlPools[c.Alias]; ok {
			return errors.New("duplicate mysql pool: " + c.Alias)
		}
		p, err := createNewMySqlPool(c)
		if err != nil {
			return errors.New(fmt.Sprintf("mysql pool %s error %v", xjson.SafeMarshal(c), err))
		}
		xsync.WithLock(mysqlLock, func() {
			newMysqlPools[c.Alias] = p
		})
	}
	return nil
}

func createNewMySqlPool(c *MySqlPoolConfig) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(c.Address), &gorm.Config{NamingStrategy: schema.NamingStrategy{SingularTable: true}})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	sqlDB.SetMaxIdleConns(c.MaxIdleConn)
	sqlDB.SetMaxOpenConns(c.MaxOpenConn)

	if err = sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, errors.WithStack(err)
	}

	return db, nil
}

func GetMySqlConn(alias string) *gorm.DB {
	mysqlLock.RLock()
	defer mysqlLock.RUnlock()
	return newMysqlPools[alias]
}

func CloseMysql() {
	xsync.WithLock(mysqlLock, func() {
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
	})
}
