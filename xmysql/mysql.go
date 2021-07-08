package xmysql

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/nickxb/pkg/xjson"
	"github.com/nickxb/pkg/xsync"
	"github.com/pkg/errors"
	"sync"
	"time"
)

var (
	mysqlPoolLock = new(sync.RWMutex)
	mysqlPools    = make(map[string]*gorm.DB)
)

type MySqlPoolConfig struct {
	Alias       string        `json:"alias"`
	Address     string        `json:"address"`
	MaxIdleConn int           `json:"max_idle_conn"`
	MaxOpenConn int           `json:"max_open_conn"`
	MaxLifeTime time.Duration `json:"max_life_time"`
	MaxIdleTime time.Duration `json:"max_idle_time"`
}

func InitMySqlPool(configs []*MySqlPoolConfig) error {
	for _, c := range configs {
		if _, ok := mysqlPools[c.Alias]; ok {
			return errors.New("duplicate mysql pool: " + c.Alias)
		}
		p, err := createMySqlPool(c)
		if err != nil {
			return errors.New(fmt.Sprintf("mysql pool %s error %v", xjson.SafeMarshal(c), err))
		}
		xsync.WithLock(mysqlPoolLock, func() {
			mysqlPools[c.Alias] = p
		})
	}
	return nil
}

func createMySqlPool(c *MySqlPoolConfig) (*gorm.DB, error) {
	db, err := gorm.Open("mysql", c.Address)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	db.DB().SetMaxIdleConns(c.MaxIdleConn)
	db.DB().SetMaxOpenConns(c.MaxOpenConn)
	if c.MaxLifeTime != 0 {
		db.DB().SetConnMaxLifetime(c.MaxLifeTime * time.Second)
	}
	if c.MaxIdleTime != 0 {
		db.DB().SetConnMaxIdleTime(c.MaxIdleTime * time.Second)
	}
	db.SingularTable(true)

	if err = db.DB().Ping(); err != nil {
		_ = db.Close()
		return nil, errors.WithStack(err)
	}

	return db, nil
}

func GetMySqlPool(alias string) *gorm.DB {
	mysqlPoolLock.RLock()
	defer mysqlPoolLock.RUnlock()
	return mysqlPools[alias]
}

func WithDbConn(alias string, fn func(db *gorm.DB) error) error {
	db := GetMySqlPool(alias)
	if db == nil {
		return errors.WithMessagef(errors.New("get mysql pool nil"), "alias: %v", alias)
	}

	return fn(db)
}

func Close() {
	xsync.WithLock(mysqlPoolLock, func() {
		for _, db := range mysqlPools {
			if db == nil {
				continue
			}
			_ = db.Close()
		}
	})
}
