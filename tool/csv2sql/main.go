package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/soulnov23/go-tool/pkg/utils"
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
	log.SetPrefix("\033[1;32m[csv2sql]\033[m ")

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
	size := len(records[0])
	var indexs []int
	var columns []string
	for i, column := range records[0] {
		tmp := strings.ToLower(column)
		if strings.Contains(tmp, "create_time") ||
			strings.Contains(tmp, "update_time") ||
			strings.Contains(tmp, "createtime") ||
			strings.Contains(tmp, "updatetime") {
			continue
		}
		indexs = append(indexs, i)
		columns = append(columns, column)
	}
	var sqls []string
	for line, record := range records[1:] {
		if len(record) != size {
			log.Fatalf("❌ [%s]第%d行列数不匹配", csvPath, line+2)
		}
		var values []string
		for _, index := range indexs {
			values = append(values, fmt.Sprintf("'%s'", utils.MySQLRealEscapeString(record[index])))
		}
		sqls = append(sqls, fmt.Sprintf("insert into %s (%s) values (%s);", tableName, strings.Join(columns, ","), strings.Join(values, ",")))
	}
	sql := strings.Join(sqls, "\n")
	if err := os.WriteFile(sqlPath, utils.StringToBytes(sql), 0o644); err != nil {
		log.Fatalf("❌ [%s]写入SQL文件失败: %s", sqlPath, err.Error())
	}
	log.Printf("✅ 结果已保存到: %s", sqlPath)
}
