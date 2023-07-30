package metacenter

import "context"

// TableField 表和字段的关联
type TableField struct {
	// ID 唯一ID
	ID int `json:"id"`
	// TableID 表ID
	TableID int `json:"table_id"`
	// FieldID 字段ID
	FieldID int `json:"field_id"`
	// RefTableID 当该字段不是TableID的物理字段，而是从其他表关联来的字段时，表的ID
	RefTableID int `json:"ref_table_id"`
	// IsUnique 是否唯一
	IsUnique int `json:"is_unique"`
	// IsPrimaryKey 是否主键
	IsPrimaryKey int `json:"is_primary_key"`
	// IsEncrypt 是否加密
	IsEncrypt int `json:"is_encrypt"`
}

// TableFieldGetter 表和字段关联获取接口
type TableFieldGetter interface {
	// GetFields 根据表ID获取field_id->*TableField
	GetFields(ctx context.Context, tableID int) map[int]*TableField
	// GetTableField 根据表ID和字段ID获取*TableField
	GetTableField(ctx context.Context, tableID, fieldID int) *TableField
}

// DefaultTableFieldGetter 默认表和字段关联获取器
type DefaultTableFieldGetter struct {
}

// NewDefaultTableFieldGetter 实例化默认表和字段关联获取器
func NewDefaultTableFieldGetter() *DefaultTableFieldGetter {
	return &DefaultTableFieldGetter{}
}

// GetFields 根据表ID获取field_id->*TableField
func (d *DefaultTableFieldGetter) GetFields(ctx context.Context, tableID int) map[int]*TableField {
	return map[int]*TableField{}
}

// GetTableField 根据表ID和字段ID获取*TableField
func (d *DefaultTableFieldGetter) GetTableField(ctx context.Context, tableID, fieldID int) *TableField {
	return &TableField{}
}
