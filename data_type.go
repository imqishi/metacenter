package metacenter

import (
	"context"
	"strings"
)

const (
	// DataTypeInt 整数类型
	DataTypeInt = "int"
	// DataTypeUInt 非负整数类型
	DataTypeUInt = "uint"
	// DataTypeString 字符串类型
	DataTypeString = "string"
	// DataTypeFloat 浮点数类型
	DataTypeFloat = "float64"
	// DataTypeDateTime 时间类型
	DataTypeDateTime = "datetime"
	// DataTypeEnum 枚举类型
	DataTypeEnum = "enum"
	// DataTypeJSON json对象类型
	DataTypeJSON = "json"
)

// DataType 数据类型
type DataType struct {
	// ID 唯一ID
	ID int `json:"id"`
	// Name 英文名，如string/int32...
	Name string `json:"name"`
	// CName 中文名，如字符串/int32...
	CName string `json:"cname"`
	// IsNum 是否为数字类型
	IsNum bool `json:"is_num"`
}

// DataTypeGetter 数据类型获取器
type DataTypeGetter interface {
	// GetByID 根据id获取数据类型配置
	GetByID(context.Context, int) *DataType
	// GetByName 根据变量类型获取数据类型配置
	GetByName(context.Context, string) *DataType
}

// DefaultDataTypeGetter 默认数据类型获取器
type DefaultDataTypeGetter struct {
}

// NewDefaultDataTypeGetter 实例化默认数据类型获取器
func NewDefaultDataTypeGetter() *DefaultDataTypeGetter {
	return &DefaultDataTypeGetter{}
}

var defaultDataType = []*DataType{
	{},
	{
		ID:    1,
		Name:  DataTypeInt,
		CName: "整数",
		IsNum: true,
	},
	{
		ID:    2,
		Name:  DataTypeUInt,
		CName: "非负整数",
		IsNum: true,
	},
	{
		ID:    3,
		Name:  DataTypeString,
		CName: "字符串",
	},
	{
		ID:    4,
		Name:  DataTypeFloat,
		CName: "浮点数",
		IsNum: true,
	},
	{
		ID:    5,
		Name:  DataTypeDateTime,
		CName: "日期时间",
	},
	{
		ID:    6,
		Name:  DataTypeEnum,
		CName: "枚举",
	},
	{
		ID:    7,
		Name:  DataTypeJSON,
		CName: "JSON对象/数组",
	},
}

var defaultDataTypeMap map[string]*DataType

func init() {
	defaultDataTypeMap = make(map[string]*DataType)
	for _, dataType := range defaultDataType {
		defaultDataTypeMap[dataType.Name] = dataType
	}
	golangDataTypeMap = make(map[string]*DataType)
	for _, dataType := range golangDataType {
		golangDataTypeMap[dataType.Name] = dataType
	}
}

// GetByID 根据id获取数据类型配置
func (d *DefaultDataTypeGetter) GetByID(ctx context.Context, id int) *DataType {
	if id >= len(defaultDataType) {
		return defaultDataType[0]
	}
	return defaultDataType[id]
}

// GetByName 根据变量类型名称获取类型配置
func (d *DefaultDataTypeGetter) GetByName(ctx context.Context, name string) *DataType {
	return defaultDataTypeMap[name]
}

// GolangDataTypeGetter Golang数据类型获取器
type GolangDataTypeGetter struct {
}

// NewGolangDataTypeGetter 实例化Golang数据类型获取器
func NewGolangDataTypeGetter() *GolangDataTypeGetter {
	return &GolangDataTypeGetter{}
}

// GetByID 根据id获取数据类型配置
func (d *GolangDataTypeGetter) GetByID(ctx context.Context, id int) *DataType {
	if id >= len(golangDataType) {
		return golangDataType[0]
	}
	return golangDataType[id]
}

var golangDataType = []*DataType{
	{},
	{
		ID:    1,
		Name:  "int",
		CName: "整数",
		IsNum: true,
	},
	{
		ID:    2,
		Name:  "int64",
		CName: "大整数",
		IsNum: true,
	},
	{
		ID:    3,
		Name:  "uint",
		CName: "无符号整数",
		IsNum: true,
	},
	{
		ID:    4,
		Name:  "uint64",
		CName: "无符号整数",
		IsNum: true,
	},
	{
		ID:    5,
		Name:  "float64",
		CName: "长浮点数",
		IsNum: true,
	},
	{
		ID:    6,
		Name:  "string",
		CName: "字符串",
	},
	{
		ID:    7,
		Name:  "utils.DateTime",
		CName: "时间",
	},
	{
		ID:    8,
		Name:  "decimal.Decimal",
		CName: "小数",
	},
}

var golangDataTypeMap map[string]*DataType

// GetByName 根据变量类型名称获取类型配置
func (d *GolangDataTypeGetter) GetByName(ctx context.Context, name string) *DataType {
	if strings.Contains(name, "int") {
		if strings.Contains(name, "big") {
			if strings.Contains(name, "unsigned") {
				return golangDataType[4]
			}
			return golangDataType[2]
		}
		if strings.Contains(name, "unsigned") {
			return golangDataType[3]
		}
		return golangDataType[1]
	}
	if strings.Contains(name, "float") || strings.Contains(name, "double") {
		return golangDataType[5]
	}
	if strings.Contains(name, "time") {
		return golangDataType[7]
	}
	if strings.Contains(name, "decimal") {
		return golangDataType[8]
	}
	return golangDataType[6]
}
