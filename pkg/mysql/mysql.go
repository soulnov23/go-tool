package mysql

import (
	"context"
	"database/sql"

	"github.com/go-sql-driver/mysql"
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

func (impl *DB) QueryResult(ctx context.Context, sql string) ([]map[string]any, error) {
	var result []map[string]any
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
	var dest []any
	// 不能用make初始化，要赋值指针
	for i := 0; i < columnLen; i++ {
		var destInterface any
		dest = append(dest, &destInterface)
	}
	for rows.Next() {
		if err := rows.Scan(dest...); err != nil {
			return nil, err
		}
		temp := make(map[string]any)
		for index, column := range columns {
			temp[column] = dest[index]
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

func Escape(value string) string {
	var dest []rune
	var escape rune
	for _, character := range value {
		escape = 0
		switch character {
		case 0: /* Must be escaped for 'mysql' */
			escape = '0'
		case '\n': /* Must be escaped for logs */
			escape = 'n'
		case '\r':
			escape = 'r'
		case '\\':
			escape = '\\'
		case '\'':
			escape = '\''
		case '"': /* Better safe than sorry */
			escape = '"'
		case '\032': /* This gives problems on Win32 */
			escape = 'Z'
		}
		if escape != 0 {
			dest = append(dest, '\\', escape)
		} else {
			dest = append(dest, character)
		}
	}
	return string(dest)
}
