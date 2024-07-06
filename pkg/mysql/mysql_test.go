package mysql

import (
	"context"
	"testing"
	"time"

	"github.com/soulnov23/go-tool/pkg/log"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TestTable struct {
	ID             int64     `gorm:"column:id" json:"id"`
	IntField       int64     `gorm:"column:int_field" json:"int_field"`
	BigintField    int64     `gorm:"column:bigint_field" json:"bigint_field"`
	FloatField     float64   `gorm:"column:float_field" json:"float_field"`
	DoubleField    float64   `gorm:"column:double_field" json:"double_field"`
	DecimalField   float64   `gorm:"column:decimal_field" json:"decimal_field"`
	TimeField      string    `gorm:"column:time_field" json:"time_field"`
	DatetimeField  time.Time `gorm:"column:datetime_field" json:"datetime_field"`
	TimestampField time.Time `gorm:"column:timestamp_field" json:"timestamp_field"`
	CharField      string    `gorm:"column:char_field" json:"char_field"`
	VarcharField   string    `gorm:"column:varchar_field" json:"varchar_field"`
	BlobField      []byte    `gorm:"column:blob_field" json:"blob_field"`
	TextField      string    `gorm:"column:text_field" json:"text_field"`
	EnumField      string    `gorm:"column:enum_field" json:"enum_field"`
	SetField       string    `gorm:"column:set_field" json:"set_field"`
}

var dbClient *gorm.DB

func getDBClient() error {
	var err error
	dbClient, err = New(context.Background(),
		"user:password@tcp(ip:port)/?timeout=1s&charset=utf8mb4&parseTime=true&loc=Local",
		log.GetDefaultLogger())
	return err
}

func TestCreateEmpty(t *testing.T) {
	if err := getDBClient(); err != nil {
		log.ErrorFields("New gorm db client failed", zap.Error(err))
		return
	}

	// Create时即使struct字段都为默认值，insert时也会把字段带上
	data := &TestTable{}
	if err := dbClient.Table("test_database.test_table").Create(data).Error; err != nil {
		log.ErrorFields("gorm.Create failed", zap.Error(err))
		return
	}
}

func TestCreate(t *testing.T) {
	if err := getDBClient(); err != nil {
		log.ErrorFields("New gorm db client failed", zap.Error(err))
		return
	}

	data := &TestTable{
		IntField:       123456789,
		BigintField:    123456789123456789,
		FloatField:     123456789.123456789,
		DoubleField:    123456789.123456789,
		DecimalField:   123456789.123456789,
		DatetimeField:  time.Now(),
		TimestampField: time.Now(),
		CharField:      "char",
		VarcharField:   "varchar",
		BlobField:      []byte("0x0123456789ABCDEF"),
		TextField:      "text",
		EnumField:      "enum1",
		SetField:       "set1",
	}
	if err := dbClient.Table("test_database.test_table").Create(data).Error; err != nil {
		log.ErrorFields("gorm.Create failed", zap.Error(err))
		return
	}
}

func TestWhere(t *testing.T) {
	if err := getDBClient(); err != nil {
		log.ErrorFields("New gorm db client failed", zap.Error(err))
		return
	}

	var results []*TestTable
	// Where时struct中字段是默认值时，都不会带在where条件中
	if err := dbClient.Table("test_database.test_table").Where(&TestTable{ID: 0, CharField: ""}).Find(&results).Error; err != nil {
		log.ErrorFields("gorm.Find failed", zap.Error(err))
		return
	}
	// Where时即使map字段都为默认值，都会带在where条件中
	if err := dbClient.Table("test_database.test_table").Where(map[string]any{"id": 0, "char_field": ""}).Find(&results).Error; err != nil {
		log.ErrorFields("gorm.Find failed", zap.Error(err))
		return
	}
}

func TestUpdates(t *testing.T) {
	if err := getDBClient(); err != nil {
		log.ErrorFields("New gorm db client failed", zap.Error(err))
		return
	}

	// Updates时struct中字段是默认值时，都不会带在set值中
	if err := dbClient.Table("test_database.test_table").Where("char_field = ?", "char").Updates(&TestTable{ID: 0, CharField: ""}).Error; err != nil {
		log.ErrorFields("gorm.Find failed", zap.Error(err))
		return
	}
	// Updates时可以使用Select强制字段默认值也要带在set值中
	if err := dbClient.Table("test_database.test_table").Where("char_field = ?", "char").Select("char_field").Updates(&TestTable{ID: 0, CharField: ""}).Error; err != nil {
		log.ErrorFields("gorm.Find failed", zap.Error(err))
		return
	}
}

func TestSave(t *testing.T) {
	if err := getDBClient(); err != nil {
		log.ErrorFields("New gorm db client failed", zap.Error(err))
		return
	}

	// Save是一个组合函数，如果保存值不包含主键，它将执行Create，否则它将执行Update(包含所有字段)
	if err := dbClient.Table("test_database.test_table").Save(&TestTable{}).Error; err != nil {
		log.ErrorFields("gorm.Find failed", zap.Error(err))
		return
	}
}
