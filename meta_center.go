package metacenter

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/pingcap/tidb/parser"
	"github.com/pingcap/tidb/parser/ast"
	"github.com/pingcap/tidb/parser/types"
	_ "github.com/pingcap/tidb/types/parser_driver"
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
	// ParseFromMySQLDDL 将MySQL-DDL语句转化为定义的meta结构
	ParseFromMySQLDDL(ctx context.Context, ddl string) (*Table, error)
}

// DefaultMetaCenter 默认实现
type DefaultMetaCenter struct {
	tableGetter      TableGetter
	tableFieldGetter TableFieldGetter
	fieldGetter      FieldGetter
	enumGetter       EnumGetter
	enumValueGetter  EnumValueGetter
	dataTypeGetter   DataTypeGetter
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
	if err := d.generateConstFile(ctx); err != nil {
		return err
	}
	if err := d.generateModelFile(ctx); err != nil {
		return err
	}
	if err := d.generateDAOFile(ctx); err != nil {
		return err
	}
	return nil
}

// generateConstFile 生成常量文件
func (d *DefaultMetaCenter) generateConstFile(ctx context.Context) error {
	tables := d.GetAllTables(ctx)
	for _, table := range tables {
		// 构造包名，go语言包名规范为全部小写字母且不包含下划线
		pkgName := strings.ReplaceAll(strings.ToLower(table.Name), "_", "") + strconv.Itoa(table.ID)
		dirPath := "./export_files/consts/" + pkgName
		if err := os.MkdirAll(dirPath, 0777); err != nil {
			return err
		}
		fileContent := fmt.Sprintf(
			`// Package %s 表%s的相关常量定义
package %s

`, pkgName, table.CName, pkgName)
		// 构造表级别常量
		fileContent = d.buildTableConsts(table, fileContent)
		// 构造字段级别常量
		// 构造字段英文名
		fileContent = d.buildTableFieldConsts(table, fileContent)
		// 构造枚举字段定义
		fileContent = d.buildTableEnumConsts(ctx, table, fileContent)
		filePath := path.Join(dirPath, "const.go")
		if err := os.WriteFile(filePath, []byte(fileContent), 0777); err != nil {
			return err
		}
		// 通过go-fmt标准化文件
		goFmtCmd := exec.CommandContext(ctx, "go", "fmt", filePath)
		if err := goFmtCmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func (d *DefaultMetaCenter) buildTableEnumConsts(ctx context.Context, table *Table, fileContent string) string {
	fileContent += `
const (`
	for _, field := range table.Fields {
		if field.EnumID == 0 {
			continue
		}
		fieldVarName := strcase.ToCamel(field.Name)
		enum := field.Enum
		fileContent += fmt.Sprintf(`
	// %s-%s枚举定义
`,
			fieldVarName, enum.CName,
		)
		// 获取枚举值类型，是数字还是字符串，默认为字符串
		valueTypeTpl := `"%s"`
		dataType := d.dataTypeGetter.GetByID(ctx, enum.DataTypeID)
		if dataType.Name == DataTypeInt {
			valueTypeTpl = "%d"
		}
		enumValues := enum.Values
		for _, enumValue := range enumValues {
			enumValueVarName := strcase.ToCamel(enumValue.EName)
			fileContent += fmt.Sprintf(`
	// %s%s %s-%s
	%s%s = `+valueTypeTpl,
				fieldVarName, enumValueVarName, enum.CName, enumValue.Desc,
				fieldVarName, enumValueVarName, enumValue.Value,
			)
		}
		fileContent += `
`
	}
	fileContent += `
)
`
	return fileContent
}

func (*DefaultMetaCenter) buildTableFieldConsts(table *Table, fileContent string) string {
	fileContent += `
const (`
	for _, field := range table.Fields {
		fieldVarName := strcase.ToCamel(field.Name)
		fileContent += fmt.Sprintf(`
	// Field%s %s字段名
	Field%s = "%s"`,
			fieldVarName, field.CName,
			fieldVarName, field.Name,
		)
	}
	fileContent += `
)
`
	return fileContent
}

func (*DefaultMetaCenter) buildTableConsts(table *Table, fileContent string) string {
	tableVarName := strcase.ToCamel(table.Name)
	fileContent += fmt.Sprintf(
		`const (
	// Table%sName %s表名
	Table%sName = "%s"
	// Table%sCName %s表中文名
	Table%sCName = "%s"`,
		tableVarName, table.CName, tableVarName, table.Name,
		tableVarName, table.Name, tableVarName, table.CName,
	)
	fileContent += `
)
`
	return fileContent
}

// generateModelFile 生成model文件
func (d *DefaultMetaCenter) generateModelFile(ctx context.Context) error {
	return nil
}

// generateDAOFile 生成dao文件
func (d *DefaultMetaCenter) generateDAOFile(ctx context.Context) error {
	return nil
}

// ParseFromMySQLDDL 将MySQL-DDL语句转化为定义的meta结构
func (d *DefaultMetaCenter) ParseFromMySQLDDL(ctx context.Context, ddl string) (*Table, error) {
	p := parser.New()
	stmts, _, err := p.ParseSQL(ddl)
	if err != nil {
		return nil, fmt.Errorf("parse ddl fail: %w", err)
	}
	if len(stmts) == 0 {
		return nil, fmt.Errorf("parse ddl fail, no stmt found")
	}
	stmt, ok := stmts[0].(*ast.CreateTableStmt)
	if !ok {
		return nil, fmt.Errorf("parse ddl fail, not ast.CreateTableStmt")
	}
	ret := d.parseMySQLDDLTable(stmt)
	for _, col := range stmt.Cols {
		field := d.parseMySQLDDLField(ctx, col)
		ret.Fields = append(ret.Fields, field)
	}
	// for _, constraint := range stmt.Constraints {
	// 	fmt.Println(constraint.Tp) // PK
	// 	for _, key := range constraint.Keys {
	// 		fmt.Println(key.Column)
	// 	}
	// 	fmt.Println("------")
	// }
	return ret, nil
}

func (d *DefaultMetaCenter) parseMySQLDDLField(ctx context.Context, col *ast.ColumnDef) *Field {
	field := &Field{
		Name: col.Name.Name.O,
	}
	switch col.Tp.EvalType() {
	case types.ETInt:
		field.Type = d.dataTypeGetter.GetByName(ctx, DataTypeInt).ID
	case types.ETDecimal:
		field.Type = d.dataTypeGetter.GetByName(ctx, DataTypeFloat).ID
	case types.ETDatetime:
		field.Type = d.dataTypeGetter.GetByName(ctx, DataTypeDateTime).ID
	default:
		field.Type = d.dataTypeGetter.GetByName(ctx, DataTypeString).ID
	}
	for _, option := range col.Options {
		if option.Tp == ast.ColumnOptionComment {
			buf := bytes.NewBuffer(nil)
			option.Expr.Format(buf)
			field.CName = strings.Trim(buf.String(), `"`)
		}
	}
	if field.CName == "" {
		field.CName = field.Name
	}
	return field
}

func (*DefaultMetaCenter) parseMySQLDDLTable(stmt *ast.CreateTableStmt) *Table {
	ret := &Table{
		Name: stmt.Table.Name.O,
	}
	for _, option := range stmt.Options {
		if option.Tp == ast.TableOptionComment {
			ret.CName = option.StrValue
		}
	}
	return ret
}
