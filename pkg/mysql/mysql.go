package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/soulnov23/go-tool/pkg/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func New(ctx context.Context, dsn string, logger log.Logger, opts ...Option) (*gorm.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %v", err)
	}
	// sql.Open无法检测连接是否有效，需要Ping一下
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("db.PingContext: %v", err)
	}
	defaultOpts := &Options{
		MaxIdleConns:    0, // 不保留空闲连接
		MaxOpenConns:    0, // 不限制打开连接
		ConnMaxLifetime: 0, // 不限制连接可以重用时间
		ConnMaxIdleTime: 0, // 不限制连接可以空闲时间
	}
	for _, opt := range opts {
		opt(defaultOpts)
	}
	db.SetMaxIdleConns(defaultOpts.MaxIdleConns)
	db.SetMaxOpenConns(defaultOpts.MaxOpenConns)
	db.SetConnMaxLifetime(defaultOpts.ConnMaxLifetime)
	db.SetConnMaxIdleTime(defaultOpts.ConnMaxIdleTime)

	gormLogger := new(logger, opts...)

	orm, err := gorm.Open(mysql.New(mysql.Config{Conn: db}), &gorm.Config{Logger: gormLogger, QueryFields: true})
	if err != nil {
		return nil, fmt.Errorf("gorm.Open: %v", err)
	}
	return orm, nil
}
