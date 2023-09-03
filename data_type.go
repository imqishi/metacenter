package metacenter

import "context"

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
func (d *DefaultDataTypeGetter) GetByID(ctx context.Context, id int) *DataType {
	switch id {
	case 1:
		return &DataType{
			ID:    1,
			Name:  DataTypeInt,
			CName: "整数",
		}
	case 2:
		return &DataType{
			ID:    2,
			Name:  DataTypeUInt,
			CName: "非负整数",
		}
	case 3:
		return &DataType{
			ID:    3,
			Name:  DataTypeString,
			CName: "字符串",
		}
	case 4:
		return &DataType{
			ID:    4,
			Name:  DataTypeFloat,
			CName: "浮点数",
		}
	case 5:
		return &DataType{
			ID:    5,
			Name:  DataTypeDateTime,
			CName: "日期时间",
		}
	case 6:
		return &DataType{
			ID:    6,
			Name:  DataTypeEnum,
			CName: "枚举",
		}
	case 7:
		return &DataType{
			ID:    7,
			Name:  DataTypeJSON,
			CName: "JSON对象/数组",
		}
	default:
		return &DataType{}
	}
}

// GetByName 根据变量类型名称获取类型配置
func (d *DefaultDataTypeGetter) GetByName(ctx context.Context, name string) *DataType {
	switch name {
	case DataTypeInt:
		return &DataType{
			ID:    1,
			Name:  DataTypeInt,
			CName: "整数",
		}
	case DataTypeUInt:
		return &DataType{
			ID:    2,
			Name:  DataTypeUInt,
			CName: "非负整数",
		}
	case DataTypeString:
		return &DataType{
			ID:    3,
			Name:  DataTypeString,
			CName: "字符串",
		}
	case DataTypeFloat:
		return &DataType{
			ID:    4,
			Name:  DataTypeFloat,
			CName: "浮点数",
		}
	case DataTypeDateTime:
		return &DataType{
			ID:    5,
			Name:  DataTypeDateTime,
			CName: "日期时间",
		}
	case DataTypeEnum:
		return &DataType{
			ID:    6,
			Name:  DataTypeEnum,
			CName: "枚举",
		}
	case DataTypeJSON:
		return &DataType{
			ID:    7,
			Name:  DataTypeJSON,
			CName: "JSON对象/数组",
		}
	default:
		return &DataType{}
	}
}
