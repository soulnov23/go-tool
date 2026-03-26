package main

import (
	"database/sql"
	"flag"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/soulnov23/go-tool/pkg/utils"
)

func main() {
	// 定义需要解析的命令行参数
	var dsn string
	var sqlQuery string
	var output string
	flag.StringVar(&dsn, "dsn", "user:password@tcp(ip:port)/?charset=utf8mb4", "mysql dsn")
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
		log.Fatalf("❌ [%s]参数解析失败: %s", dsn, err.Error())
	}
	defer db.Close()
	// sql.Open 只验证参数格式，不建立真实连接，需要 Ping 确认连通性
	if err := db.Ping(); err != nil {
		log.Fatalf("❌ 连接失败: %s", err.Error())
	}
	log.Printf("✅ 连接成功")
	rows, err := db.Query(sqlQuery)
	if err != nil {
		log.Fatalf("❌ 查询失败: %s", err.Error())
	}
	defer rows.Close()
	log.Printf("✅ 查询成功")
	columns, err := rows.Columns()
	if err != nil {
		log.Fatalf("❌ 获取列名失败: %s", err.Error())
	}
	log.Printf("✅ 获取列名成功")
	result := make([]map[string]any, 0)
	values := make([]any, len(columns))
	ptrs := make([]any, len(columns))
	for i := range ptrs {
		ptrs[i] = &values[i]
	}
	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			log.Fatalf("❌ 扫描行失败: %s", err.Error())
		}
		record := make(map[string]any, len(columns))
		for i, column := range columns {
			value := values[i]
			if bytes, ok := value.([]byte); ok {
				value = string(bytes)
			}
			record[column] = value
		}
		result = append(result, record)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("❌ 行迭代失败: %s", err.Error())
	}
	if err := os.WriteFile(output, utils.Bytesify(result), 0o644); err != nil {
		log.Fatalf("❌ 写入文件失败: %s: %s", output, err.Error())
	}
	log.Printf("✅ 结果已保存到: %s", output)
}
