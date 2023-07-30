package metacenter

import "context"

// Table 表基础信息
type Table struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	CName   string `json:"cname"`
	DBDsn   string `json:"db_dsn"`
	ESDsn   string `json:"es_dsn"`
	ESIndex string `json:"es_index"`
	ESSync  int    `json:"es_sync"`

	Fields     []*Field          `json:"-"`
	NameFields map[string]*Field `json:"-"`
}

// TableGetter 表配置获取接口
type TableGetter interface {
	// GetAll 获取所有表配置
	GetAll(context.Context) []*Table
	// GetByID 根据表ID获取配置
	GetByID(context.Context, int) *Table
	// GetByName 根据表名获取配置
	GetByName(context.Context, string) *Table
}

// DefaultTableGetter 默认表配置获取器
type DefaultTableGetter struct {
}

// NewDefaultTableGetter 实例化默认表配置获取器
func NewDefaultTableGetter() *DefaultTableGetter {
	return &DefaultTableGetter{}
}

// GetAll 获取所有表配置
func (d *DefaultTableGetter) GetAll(ctx context.Context) []*Table {
	return nil
}

// GetByID 根据表ID获取配置
func (d *DefaultTableGetter) GetByID(ctx context.Context, id int) *Table {
	return &Table{}
}

// GetByName 根据表名获取配置
func (d *DefaultTableGetter) GetByName(ctx context.Context, name string) *Table {
	return &Table{}
}
