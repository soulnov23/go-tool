package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 定义需要解析的命令行参数
	var csvPath string
	var sqlPath string
	var tableName string
	flag.StringVar(&csvPath, "csv", "tmp.csv", "csv file path")
	flag.StringVar(&sqlPath, "sql", "tmp.sql", "sql file path")
	flag.StringVar(&tableName, "table", "tmp", "table name")
	// 开始解析命令行
	flag.Parse()
	// 命令行参数都不匹配，打印help
	if flag.NFlag() == 0 {
		flag.Usage()
		return
	}

	log.SetFlags(0)
	log.SetPrefix("\033[1;32m[sql2json]\033[m ")

	csvFile, err := os.Open(csvPath)
	if err != nil {
		log.Fatalf("❌ [%s]打开CSV文件失败: %s", csvPath, err.Error())
	}
	defer csvFile.Close()
	records, err := csv.NewReader(csvFile).ReadAll()
	if err != nil {
		log.Fatalf("❌ [%s]读取CSV文件失败: %s", csvPath, err.Error())
	}
	if len(records) == 0 {
		log.Fatalf("❌ [%s]CSV文件为空", csvPath)
	}
	var indexs []int
	var columns []string
	for i, column := range records[0] {
		if strings.Contains(strings.ToLower(column), "create_time") ||
			strings.Contains(strings.ToLower(column), "update_time") ||
			strings.Contains(strings.ToLower(column), "createtime") ||
			strings.Contains(strings.ToLower(column), "updatetime") {
			continue
		}
		indexs = append(indexs, i)
		columns = append(columns, column)
	}
	var sqls []string
	for _, record := range records[1:] {
		values := make([]string, len(indexs))
		for i, index := range indexs {
			values[i] = fmt.Sprintf("'%s'", record[index])
		}
		sqls = append(sqls, fmt.Sprintf("insert into %s (%s) values (%s);", tableName, strings.Join(columns, ","), strings.Join(values, ",")))
	}
	sql := strings.Join(sqls, "\n")
	sqlFile, err := os.OpenFile(sqlPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		log.Fatalf("❌ [%s]打开SQL文件失败: %s", sqlPath, err.Error())
	}
	defer sqlFile.Close()
	if _, err := sqlFile.WriteString(sql); err != nil {
		log.Fatalf("❌ [%s]写入SQL文件失败: %s", sqlPath, err.Error())
	}
	log.Printf("✅ 结果已保存到: %s", sqlPath)
}
