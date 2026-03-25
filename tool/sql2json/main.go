package main

import (
	"database/sql"
	"flag"
	"log"
	"os"

	"github.com/soulnov23/go-tool/pkg/utils"
)

func main() {
	// 定义需要解析的命令行参数
	var dsn string
	var sqlQuery string
	var output string
	flag.StringVar(&dsn, "dsn", "dsn://user:password@tcp(ip:port)/?charset=utf8mb4", "mysql dsn")
	flag.StringVar(&sqlQuery, "sql", "select * from table", "sql query")
	flag.StringVar(&output, "output", "tmp.json", "output file path")
	// 开始解析命令行
	flag.Parse()
	// 命令行参数都不匹配，打印help
	if flag.NFlag() == 0 {
		flag.Usage()
		return
	}

	log.SetFlags(0)
	log.SetPrefix("\033[1;32m[sql2json]\033[m ")

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("❌ [%s]连接失败: %s", dsn, err.Error())
	}
	defer db.Close()
	log.Printf("✅ [%s]连接成功", dsn)
	// sql.Open无法检测连接是否有效，需要Ping一下
	if err := db.Ping(); err != nil {
		log.Fatalf("❌ [%s]Ping失败: %s", dsn, err.Error())
	}
	log.Printf("✅ [%s]Ping成功", dsn)
	rows, err := db.Query(sqlQuery)
	if err != nil {
		log.Fatalf("❌ [%s]查询失败: %s", dsn, err.Error())
	}
	defer rows.Close()
	log.Printf("✅ [%s]查询成功", dsn)
	columns, err := rows.Columns()
	if err != nil {
		log.Fatalf("❌ [%s]获取列名失败: %s", dsn, err.Error())
	}
	log.Printf("✅ [%s]获取列名成功", dsn)
	var result []map[string]any
	for rows.Next() {
		values := make([]any, len(columns))
		for i := range values {
			values[i] = new(any)
		}
		if err := rows.Scan(values...); err != nil {
			log.Fatalf("❌ [%s]扫描行失败: %s", dsn, err.Error())
		}
		record := make(map[string]any, len(columns))
		for i, column := range columns {
			record[column] = values[i]
		}
		result = append(result, record)
	}
	os.WriteFile(output, utils.Bytesify(result), 0o644)
	log.Printf("✅ 结果已保存到 %s", output)
}
