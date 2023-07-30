package metacenter

import "context"

// EnumValue 枚举值
type EnumValue struct {
	// ID 唯一ID
	ID int `json:"id"`
	// EnumID enum的唯一ID
	EnumID int `json:"enum_id"`
	// EName 英文名，用于生成常量名称
	EName string `json:"ename"`
	// Desc 字段描述
	Desc string `json:"desc"`
	// Value 枚举值，如1/2/3，waiting/start/finish...
	Value string `json:"value"`
	// Status 枚举状态，用于下线部分枚举值
	Status int `json:"status"`
	// Explain 备注
	Explain string `json:"explain"`
}

// EnumValueGetter 枚举值获取接口
type EnumValueGetter interface {
	// FindByEnumID 根据enum的id获取值列表
	FindByEnumID(context.Context, int) []*EnumValue
}

// DefaultEnumValueGetter 默认枚举值获取器
type DefaultEnumValueGetter struct {
}

// NewDefaultEnumValueGetter 实例化默认枚举值获取器
func NewDefaultEnumValueGetter() *DefaultEnumValueGetter {
	return &DefaultEnumValueGetter{}
}

// FindByEnumID 根据enum的id获取值列表
func (d *DefaultEnumValueGetter) FindByEnumID(context.Context, int) []*EnumValue {
	return nil
}
