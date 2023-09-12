package mysql

import (
	"context"
	"database/sql"

	"github.com/go-sql-driver/mysql"
	convert "github.com/soulnov23/go-tool/pkg/strconv"
)

type DB struct {
	*sql.DB
}

func New(ctx context.Context, dsn string, opts ...Option) (*DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	// sql.Open无法检测连接是否有效，需要Ping一下
	if err := db.PingContext(ctx); err != nil {
		return nil, err
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
	return &DB{DB: db}, nil
}

func (impl *DB) Query(ctx context.Context, sql string) ([]map[string]string, error) {
	var result []map[string]string
	rows, err := impl.DB.QueryContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	columnLen := len(columns)
	var dest []interface{}
	// 不能用make初始化，要赋值指针
	for i := 0; i < columnLen; i++ {
		var destInterface interface{}
		dest = append(dest, &destInterface)
	}
	for rows.Next() {
		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}
		temp := make(map[string]string)
		for index, column := range columns {
			temp[column] = convert.AnyToString(dest[index])
		}
		result = append(result, temp)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func DuplicateEntry(err error) bool {
	mysqlErr, ok := err.(*mysql.MySQLError)
	if ok && mysqlErr.Number == 1062 /*ER_DUP_ENTRY*/ {
		return true
	}
	return false
}
