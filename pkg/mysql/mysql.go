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

	config := &gorm.Config{
		Logger:               gormLogger,
		DryRun:               false, // 调试时可以打开查看生成的SQL
		DisableAutomaticPing: true,  // Open禁用自动ping数据库
		QueryFields:          true,  // select *使用全部字段
	}
	if defaultOpts.DryRun {
		config.DryRun = defaultOpts.DryRun
	}

	// First、Last、Take等方法未找到记录时，GORM会返回gorm.ErrRecordNotFound，其它错误需要使用.(*mysql.MySQLError)转换去判断
	// *mysql.MySQLError.Number参考https://dev.mysql.com/doc/mysql-errors/8.0/en/error-reference-introduction.html
	orm, err := gorm.Open(mysql.New(mysql.Config{Conn: db}), config)
	if err != nil {
		return nil, fmt.Errorf("gorm.Open: %v", err)
	}
	return orm, nil
}
