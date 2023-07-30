package metacenter

import "context"

// Enum 枚举定义
type Enum struct {
	// ID 唯一ID
	ID int `json:"id"`
	// CName 枚举中文名
	CName string `json:"cname"`
	// DataTypeID 枚举值类型
	DataTypeID int `json:"data_type_id"`
	// Explain 备注
	Explain string `json:"explain"`

	Values       []*EnumValue          `json:"-"`
	Value2Values map[string]*EnumValue `json:"-"`
}

// EnumGetter 枚举获取接口
type EnumGetter interface {
	// GetByID 根据枚举ID获取枚举配置
	GetByID(context.Context, int) *Enum
	// FindByIDs 批量根据枚举ID获取id->*Enum
	FindByIDs(context.Context, []int) map[int]*Enum
}

// DefaultEnumGetter 默认枚举获取器
type DefaultEnumGetter struct {
}

// NewDefaultEnumGetter 实例化默认枚举获取器
func NewDefaultEnumGetter() *DefaultEnumGetter {
	return &DefaultEnumGetter{}
}

// GetByID 根据枚举ID获取枚举配置
func (d *DefaultEnumGetter) GetByID(context.Context, int) *Enum {
	return &Enum{}
}

// FindByIDs 批量根据枚举ID获取id->*Enum
func (d *DefaultEnumGetter) FindByIDs(context.Context, []int) map[int]*Enum {
	return map[int]*Enum{}
}
