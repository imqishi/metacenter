package metacenter

import "context"

const (
	// DataTypeInt 数字类型
	DataTypeInt = "int"
	// DataTypeString 字符串类型
	DataTypeString = "string"
	// DataTypeFloat 浮点数类型
	DataTypeFloat = "float64"
	// DataTypeDateTime 时间类型
	DataTypeDateTime = "datetime"
	// DataTypeEnum 枚举类型
	DataTypeEnum = "enum"
)

// DataType 数据类型
type DataType struct {
	// ID 唯一ID
	ID int `json:"id"`
	// Name 英文名，如string/int32...
	Name string `json:"name"`
	// CName 中文名，如字符串/int32...
	CName string `json:"cname"`
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

// GetByID 根据id获取数据类型配置
func (d *DefaultDataTypeGetter) GetByID(context.Context, int) *DataType {
	return &DataType{}
}

// GetByName 根据变量类型名称获取类型配置
func (d *DefaultDataTypeGetter) GetByName(context.Context, string) *DataType {
	return &DataType{}
}
