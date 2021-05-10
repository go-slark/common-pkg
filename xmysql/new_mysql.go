package xmysql

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/smallfish-root/common-pkg/xjson"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"sync"
)

var (
	newMysqlPools = make(map[string]*gorm.DB)
	once          sync.Once
)

func InitMySql(configs []*MySqlPoolConfig) error {
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
