package metacenter

import "context"

// Field 字段定义
type Field struct {
	// ID 唯一ID
	ID int `json:"id"`
	// Name 字段英文名
	Name string `json:"name"`
	// CName 字段中文名
	CName string `json:"cname"`
	// Type 字段类型
	Type int `json:"type"`
	// EnumID 当Type是枚举类型时，指向枚举enum的ID
	EnumID int `json:"enum_id"`
	// ESFieldType ES映射字段类型
	ESFieldType string `json:"es_field_type"`
	// Explain 备注
	Explain string `json:"explain"`
	// IsPK 是否主键
	IsPK bool `json:"is_pk"`
	// AutoIncr 是否自增
	AutoIncr bool `json:"auto_incr"`

	Enum *Enum `json:"-"`
}

// FieldGetter 字段获取接口
type FieldGetter interface {
	// GetByID 根据字段ID获取字段配置
	GetByID(context.Context, int) *Field
	// GetByName 根据字段英文名获取字段配置
	GetByName(context.Context, string) *Field
	// FindByIDs 批量根据字段ID获取id->*Field
	FindByIDs(context.Context, []int) map[int]*Field
	// FindByNames 批量根据字段ID获取name->*Field
	FindByNames(context.Context, []string) map[string]*Field
}

// DefaultFieldGetter 默认字段获取器
type DefaultFieldGetter struct {
}

// NewDefaultFieldGetter 实例化默认字段获取器
func NewDefaultFieldGetter() *DefaultFieldGetter {
	return &DefaultFieldGetter{}
}

// GetByID 根据字段ID获取字段配置
func (d *DefaultFieldGetter) GetByID(context.Context, int) *Field {
	return &Field{}
}

// GetByName 根据字段英文名获取字段配置
func (d *DefaultFieldGetter) GetByName(context.Context, string) *Field {
	return &Field{}
}

// FindByIDs 批量根据字段ID获取id->*Field
func (d *DefaultFieldGetter) FindByIDs(context.Context, []int) map[int]*Field {
	return map[int]*Field{}
}

// FindByNames 批量根据字段ID获取name->*Field
func (d *DefaultFieldGetter) FindByNames(context.Context, []string) map[string]*Field {
	return map[string]*Field{}
}
