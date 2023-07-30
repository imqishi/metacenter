package metacenter

import (
	"context"
	"os"
	"strconv"
	"strings"
)

// MetaCenter 元信息中心接口
type MetaCenter interface {
	// GetTableByName 根据表名获取配置
	GetTableByName(ctx context.Context, name string) *Table
	// GetTableByID 根据表ID获取配置
	GetTableByID(ctx context.Context, id int) *Table
	// GetAllTables 获取所有表配置
	GetAllTables(ctx context.Context) []*Table
	// GenerateGoFiles 生成go文件
	GenerateGoFiles(ctx context.Context) error
}

// DefaultMetaCenter 默认实现
type DefaultMetaCenter struct {
	tableGetter      TableGetter
	tableFieldGetter TableFieldGetter
	fieldGetter      FieldGetter
	enumGetter       EnumGetter
	enumValueGetter  EnumValueGetter
}

// NewDefaultMetaCenter 实例化默认元信息中心
func NewDefaultMetaCenter(ctx context.Context) *DefaultMetaCenter {
	return &DefaultMetaCenter{}
}

// GetTableByName 根据表名获取配置
func (d *DefaultMetaCenter) GetTableByName(ctx context.Context, name string) *Table {
	table := d.tableGetter.GetByName(ctx, name)
	return d.getTableFields(ctx, table)
}

func (d *DefaultMetaCenter) getTableFields(ctx context.Context, table *Table) *Table {
	if table == nil {
		return nil
	}
	table.NameFields = make(map[string]*Field)
	tableFields := d.tableFieldGetter.GetFields(ctx, table.ID)
	for fieldID := range tableFields {
		field := d.fieldGetter.GetByID(ctx, fieldID)
		enum := d.enumGetter.GetByID(ctx, field.EnumID)
		if enum != nil {
			enum.Value2Values = make(map[string]*EnumValue)
			enumValues := d.enumValueGetter.FindByEnumID(ctx, enum.ID)
			for _, enumValue := range enumValues {
				enum.Values = append(enum.Values, enumValue)
				enum.Value2Values[enumValue.Value] = enumValue
			}
			field.Enum = enum
		}
		table.Fields = append(table.Fields, field)
		table.NameFields[field.Name] = field
	}
	return table
}

// GetTableByID 根据表ID获取配置
func (d *DefaultMetaCenter) GetTableByID(ctx context.Context, id int) *Table {
	table := d.tableGetter.GetByID(ctx, id)
	return d.getTableFields(ctx, table)
}

// GetAllTables 获取所有表配置name->*Table
func (d *DefaultMetaCenter) GetAllTables(ctx context.Context) []*Table {
	tables := d.tableGetter.GetAll(ctx)
	ret := make([]*Table, len(tables))
	for i, table := range tables {
		ret[i] = d.getTableFields(ctx, table)
	}
	return ret
}

// GenerateGoFiles 生成go文件
func (d *DefaultMetaCenter) GenerateGoFiles(ctx context.Context) error {
	tables := d.GetAllTables(ctx)
	for _, table := range tables {
		// 构造包名，go语言包名规范为全部小写字母且不包含下划线
		pkgName := strings.ReplaceAll(strings.ToLower(table.Name), "_", "") + strconv.Itoa(table.ID)
		if err := os.MkdirAll("./export_consts/"+pkgName, 0644); err != nil {
			return err
		}
		// todo...
	}
	return nil
}
