package metacenter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"text/template"

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
	// ToESTemplate 将Table转换为es模板
	ToESTemplate(ctx context.Context, table *Table) (string, error)
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
	tplFileBody, err := os.ReadFile("./tpl_files/const.tpl")
	if err != nil {
		return err
	}
	tpl := template.Must(template.New("const").Parse(string(tplFileBody)))
	tables := d.GetAllTables(ctx)
	for _, table := range tables {
		param := d.getTplParam(ctx, table)
		dirPath := "./export_files/" + param.PkgName
		if err := os.MkdirAll(dirPath, 0777); err != nil {
			return err
		}
		filePath := path.Join(dirPath, "const.go")
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}
		if err := tpl.Execute(file, param); err != nil {
			return err
		}
		if err := file.Close(); err != nil {
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

// TplTable 生成文件使用到的模板表参数
type TplTable struct {
	VarName string // TTestName
	CName   string // 测试表
	Name    string // t_test
}

// TplEnumValue 生成文件使用到的模板枚举参数
type TplEnumValue struct {
	VarName string // Running
	CName   string // 运行中
	Value   string // 1
}

// TplField 生成文件使用到的模板字段参数
type TplField struct {
	VarName    string // Status
	Type       string // int
	Name       string // status
	CName      string // 状态
	IsEnum     bool   // true
	IsNum      bool   // true
	EnumValues []TplEnumValue
}

// TplParam 生成文件使用到的模板参数
type TplParam struct {
	PkgName string
	Table   TplTable
	Fields  []TplField
	HasEnum bool
}

func (d *DefaultMetaCenter) getTplParam(ctx context.Context, table *Table) *TplParam {
	// 构造包名，go语言包名规范为全部小写字母且不包含下划线，避免有重名表，增加id后缀
	pkgName := strings.ReplaceAll(strings.ToLower(table.Name), "_", "") + strconv.Itoa(table.ID)
	param := &TplParam{
		PkgName: pkgName,
		Table: TplTable{
			VarName: strcase.ToCamel(table.Name),
			CName:   table.CName,
			Name:    table.Name,
		},
	}
	for _, field := range table.Fields {
		fieldVarName := strcase.ToCamel(field.Name)
		dataType := d.dataTypeGetter.GetByID(ctx, field.Type)
		tplField := TplField{
			VarName: fieldVarName,
			Type:    dataType.Name,
			Name:    field.Name,
			CName:   field.CName,
			IsEnum:  false,
		}
		if dataType.Name == DataTypeEnum {
			param.HasEnum = true
			tplField.Type = d.dataTypeGetter.GetByID(ctx, field.Enum.DataTypeID).Name
			tplField.IsEnum = true
			tplField.IsNum = tplField.Type == DataTypeInt || tplField.Type == DataTypeUInt
			for _, enumValue := range field.Enum.Values {
				tplField.EnumValues = append(tplField.EnumValues, TplEnumValue{
					VarName: strcase.ToCamel(enumValue.EName),
					CName:   enumValue.Desc,
					Value:   enumValue.Value,
				})
			}
		}
		param.Fields = append(param.Fields, tplField)
	}
	return param
}

// generateModelFile 生成model文件
func (d *DefaultMetaCenter) generateModelFile(ctx context.Context) error {
	tplFileBody, err := os.ReadFile("./tpl_files/model.tpl")
	if err != nil {
		return err
	}
	tpl := template.Must(template.New("model").Parse(string(tplFileBody)))
	tables := d.GetAllTables(ctx)
	for _, table := range tables {
		param := d.getTplParam(ctx, table)
		dirPath := "./export_files/" + param.PkgName
		if err := os.MkdirAll(dirPath, 0777); err != nil {
			return err
		}
		filePath := path.Join(dirPath, "model.go")
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}
		if err := tpl.Execute(file, param); err != nil {
			return err
		}
		if err := file.Close(); err != nil {
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
	// 解析字段英文名
	field := &Field{
		Name: col.Name.Name.O,
	}
	// 解析字段类型
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
	// 解析字段注释，尝试解析字段的中文名，以及如果有枚举值解析为枚举类型
	for _, option := range col.Options {
		if option.Tp == ast.ColumnOptionComment {
			buf := bytes.NewBuffer(nil)
			option.Expr.Format(buf)
			comment := strings.Trim(buf.String(), `"`)
			name, enumKV := d.tryParseEnumFromComment(comment)
			field.CName = name
			if len(enumKV) == 0 {
				continue
			}
			field.Type = d.dataTypeGetter.GetByName(ctx, DataTypeEnum).ID
			field.Enum = &Enum{
				CName:      name,
				DataTypeID: d.dataTypeGetter.GetByName(ctx, DataTypeInt).ID,
			}
			for k, v := range enumKV {
				field.Enum.Values = append(field.Enum.Values, &EnumValue{
					EnumID: field.Enum.ID,
					EName:  strcase.ToCamel(field.Name) + k,
					Desc:   v,
					Value:  k,
				})
			}
		}
	}
	if field.CName == "" {
		field.CName = field.Name
	}
	return field
}

var (
	enumNameRE = regexp.MustCompile(`(\S+)(\s+(\d+)(:|：|-)(\S+))`)
	enumKVRE   = regexp.MustCompile(`(\d+)(:|：|-)(\S+)`)
)

// tryParseEnumFromComment 将以下格式的注释解析为枚举信息，当前只支持数字枚举
// 任务状态 1-待处理 2-处理中 3-成功 4-失败
// 任务状态 1：待处理 2：处理中 3：成功 4：失败
// 任务状态 1:待处理 2:处理中 3:成功 4:失败
func (*DefaultMetaCenter) tryParseEnumFromComment(comment string) (string, map[string]string) {
	res := enumNameRE.FindAllStringSubmatch(comment, 1)
	if len(res) == 0 {
		return comment, nil
	}
	name := res[0][1]
	res = enumKVRE.FindAllStringSubmatch(comment, -1)
	if len(res) == 0 {
		return name, nil
	}
	kv := make(map[string]string)
	for _, subRes := range res {
		kv[subRes[1]] = subRes[3]
	}
	return name, kv
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

// ESTemplate es模板配置
type ESTemplate struct {
	IndexPatterns []string `json:"index_patterns"`
	Template      struct {
		Settings struct {
			MaxResultWindow  int `json:"max_result_window"`
			NumberOfShards   int `json:"number_of_shards"`
			NumberOfReplicas int `json:"number_of_replicas"`
		} `json:"settings"`
		Mappings struct {
			Source struct {
				Enabled bool `json:"enabled"`
			} `json:"_source"`
			Properties map[string]interface{} `json:"properties"`
		} `json:"mappings"`
	} `json:"template"`
	Priority int `json:"priority"`
	Version  int `json:"version"`
}

// ToESTemplate 将Table转换为es模板
func (d *DefaultMetaCenter) ToESTemplate(ctx context.Context, table *Table) (string, error) {
	tpl := ESTemplate{}
	indexConfig := table.ESConfig.Index
	tpl.IndexPatterns = []string{indexConfig.NameOrPrefix}
	if indexConfig.MultiIndex {
		tpl.IndexPatterns = []string{indexConfig.NameOrPrefix + "*"}
	}
	if indexConfig.MaxResultWindow != 0 {
		tpl.Template.Settings.MaxResultWindow = indexConfig.MaxResultWindow
	}
	tpl.Template.Settings.NumberOfShards = 3
	if indexConfig.NumberOfShards != 0 {
		tpl.Template.Settings.NumberOfShards = indexConfig.NumberOfShards
	}
	if indexConfig.NumberOfReplicas != 0 {
		tpl.Template.Settings.NumberOfReplicas = indexConfig.NumberOfReplicas
	}
	tpl.Template.Mappings.Source.Enabled = true
	for _, field := range table.Fields {
		var fieldMapping map[string]interface{}
		typeName := d.dataTypeGetter.GetByID(ctx, field.Type).Name
		switch typeName {
		case DataTypeInt:
			fieldMapping = map[string]interface{}{"type": "long"}
		case DataTypeUInt:
			fieldMapping = map[string]interface{}{"type": "unsigned_long"}
		case DataTypeFloat:
			fieldMapping = map[string]interface{}{"type": "double"}
		case DataTypeDateTime:
			fieldMapping = map[string]interface{}{
				"type": "date", "format": "yyyy-MM-dd HH:mm:ss", "ignore_malformed": true}
		case DataTypeEnum:
			fieldMapping = map[string]interface{}{"type": "keyword"}
			enumTypeName := d.dataTypeGetter.GetByID(ctx, field.Enum.DataTypeID).Name
			if enumTypeName == DataTypeInt || enumTypeName == DataTypeUInt {
				fieldMapping = map[string]interface{}{"type": "long"}
			}
		case DataTypeJSON:
			fieldMapping = map[string]interface{}{"type": "nested"}
		default:
			fieldMapping = map[string]interface{}{"type": "keyword"}
			if field.ESFieldType == "text" {
				fieldMapping = map[string]interface{}{
					"type":            "text",
					"search_analyzer": "ik_smart",
					"analyzer":        "ik_max_word",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type": "keyword",
						},
					},
				}
			}
		}
		tpl.Template.Mappings.Properties[field.Name] = fieldMapping
	}
	body, _ := json.Marshal(tpl)
	return string(body), nil
}
