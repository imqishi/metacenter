package metacenter

import "context"

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
	GetByID(context.Context, string) *DataType
}

// DefaultDataTypeGetter 默认数据类型获取器
type DefaultDataTypeGetter struct {
}

// NewDefaultDataTypeGetter 实例化默认数据类型获取器
func NewDefaultDataTypeGetter() *DefaultDataTypeGetter {
	return &DefaultDataTypeGetter{}
}

// GetByID 根据id获取数据类型配置
func (d *DefaultDataTypeGetter) GetByID(context.Context, string) *DataType {
	return &DataType{}
}
