package metacenter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"github.com/pingcap/tidb/parser"
	"github.com/pingcap/tidb/parser/ast"
	_ "github.com/pingcap/tidb/types/parser_driver"
	"github.com/pkg/errors"
)

// GenerateGoFilesParam 生成Go模板文件可指定的参数
type GenerateGoFilesParam struct {
	// Name 输出文件类型，如model/dao/const...
	Name string
	// TplFilePath 模板文件路径
	TplFilePath string
	// OutputDirPath 输出文件夹路径
	OutputDirPath string
	// InjectParams 注入任意额外参数
	InjectParams map[string]string
}

// Fmt 格式化参数
func (p *GenerateGoFilesParam) Fmt() error {
	if p.Name == "" {
		return fmt.Errorf("param Name cannot be empty")
	}
	if p.OutputDirPath == "" {
		p.OutputDirPath = "./default"
	}
	if p.OutputDirPath[len(p.OutputDirPath)-1] == '/' {
		p.OutputDirPath = p.OutputDirPath[:len(p.OutputDirPath)-1]
	}
	return nil
}

// MetaCenter 元信息中心接口
type MetaCenter interface {
	// GetTableByName 根据表名获取配置
	GetTableByName(ctx context.Context, name string) *Table
	// GetTableByID 根据表ID获取配置
	GetTableByID(ctx context.Context, id int) *Table
	// GetAllTables 获取所有表配置
	GetAllTables(ctx context.Context) []*Table
	// GenerateGoFiles 生成go文件
	GenerateGoFiles(ctx context.Context, params []*GenerateGoFilesParam) error
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

// DefaultMetaCenterOption 可选参数
type DefaultMetaCenterOption func(*DefaultMetaCenter)

// WithTableGetter 指定TableGetter
func WithTableGetter(tg TableGetter) DefaultMetaCenterOption {
	return func(dm *DefaultMetaCenter) {
		dm.tableGetter = tg
	}
}

// WithTableFieldGetter 指定TableFieldGetter
func WithTableFieldGetter(tfg TableFieldGetter) DefaultMetaCenterOption {
	return func(dm *DefaultMetaCenter) {
		dm.tableFieldGetter = tfg
	}
}

// WithFieldGetter 指定FieldGetter
func WithFieldGetter(fg FieldGetter) DefaultMetaCenterOption {
	return func(dm *DefaultMetaCenter) {
		dm.fieldGetter = fg
	}
}

// WithEnumGetter 指定EnumGetter
func WithEnumGetter(eg EnumGetter) DefaultMetaCenterOption {
	return func(dm *DefaultMetaCenter) {
		dm.enumGetter = eg
	}
}

// WithEnumValueGetter 指定EnumValueGetter
func WithEnumValueGetter(evg EnumValueGetter) DefaultMetaCenterOption {
	return func(dm *DefaultMetaCenter) {
		dm.enumValueGetter = evg
	}
}

// WithDataTypeGetter 指定DataTypeGetter
func WithDataTypeGetter(dtg DataTypeGetter) DefaultMetaCenterOption {
	return func(dm *DefaultMetaCenter) {
		dm.dataTypeGetter = dtg
	}
}

// NewDefaultMetaCenter 实例化默认元信息中心
func NewDefaultMetaCenter(ctx context.Context, opts ...DefaultMetaCenterOption) *DefaultMetaCenter {
	center := &DefaultMetaCenter{}
	for _, opt := range opts {
		opt(center)
	}
	return center
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
func (d *DefaultMetaCenter) GenerateGoFiles(ctx context.Context, tables []*Table, params []*GenerateGoFilesParam) error {
	for _, param := range params {
		if err := param.Fmt(); err != nil {
			return errors.Wrapf(err, "param check fail")
		}
		tplFileBody, err := os.ReadFile(param.TplFilePath)
		if err != nil {
			return errors.Wrapf(err, "read TplFilePath(%s) fail", param.TplFilePath)
		}
		tpl := template.Must(template.New(param.Name).Parse(string(tplFileBody)))
		for _, table := range tables {
			tplParam := d.getTplParam(ctx, table, param)
			if err := os.MkdirAll(param.OutputDirPath, 0777); err != nil {
				return errors.Wrapf(err, "mkdir(%s) fail", param.OutputDirPath)
			}
			filePath := path.Join(param.OutputDirPath, fmt.Sprintf("%s_%s.go", table.Name, param.Name))
			filePath, err := filepath.Abs(filePath)
			if err != nil {
				return errors.Wrapf(err, "create file abs path(%s) fail", filePath)
			}
			file, err := os.Create(filePath)
			if err != nil {
				return errors.Wrapf(err, "create file(%s) fail", filePath)
			}
			if err := tpl.Execute(file, tplParam); err != nil {
				return errors.Wrapf(err, "tpl(%s) execute fail", param.TplFilePath)
			}
			if err := file.Close(); err != nil {
				return errors.Wrapf(err, "gen file(%s) close fail", param.TplFilePath)
			}
			// 通过go-fmt标准化文件
			goFmtCmd := exec.CommandContext(ctx, "go", "fmt", filePath)
			if err := goFmtCmd.Run(); err != nil {
				return errors.Wrapf(err, "go fmt file(%s) fail", filePath)
			}
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
	IsPK       bool   // true
	AutoIncr   bool   // false
	EnumValues []TplEnumValue
}

// TplParam 生成文件使用到的模板参数
type TplParam struct {
	PkgName      string
	Table        TplTable
	Fields       []TplField
	PKFields     []TplField
	HasEnum      bool
	HasDecimal   bool
	InjectParams map[string]string
}

func (d *DefaultMetaCenter) getTplParam(ctx context.Context, table *Table, genParam *GenerateGoFilesParam) *TplParam {
	param := &TplParam{
		// 包名为输出文件夹同名
		PkgName: path.Base(genParam.OutputDirPath),
		Table: TplTable{
			VarName: strcase.ToCamel(table.Name),
			CName:   table.CName,
			Name:    table.Name,
		},
		InjectParams: genParam.InjectParams,
	}
	for _, field := range table.Fields {
		fieldVarName := strcase.ToCamel(field.Name)
		dataType := d.dataTypeGetter.GetByID(ctx, field.Type)
		tplField := TplField{
			VarName:  fieldVarName,
			Type:     dataType.Name,
			Name:     field.Name,
			CName:    field.CName,
			IsNum:    dataType.IsNum,
			IsPK:     field.IsPK,
			AutoIncr: field.AutoIncr,
			IsEnum:   false,
		}
		if field.Enum != nil {
			param.HasEnum = true
			tplField.IsEnum = true
			tplField.Type = d.dataTypeGetter.GetByID(ctx, field.Enum.DataTypeID).Name
			for _, enumValue := range field.Enum.Values {
				tplField.EnumValues = append(tplField.EnumValues, TplEnumValue{
					VarName: strcase.ToCamel(enumValue.EName),
					CName:   enumValue.Desc,
					Value:   enumValue.Value,
				})
			}
		}
		if dataType.Name == goDecimalType {
			param.HasDecimal = true
		}
		param.Fields = append(param.Fields, tplField)
		if field.IsPK {
			param.PKFields = append(param.PKFields, tplField)
		}
	}
	return param
}

var fmtMySQLDDLRE = regexp.MustCompile(`shardkey=.*`)

// ParseFromMySQLDDL 将MySQL-DDL语句转化为定义的meta结构
func (d *DefaultMetaCenter) ParseFromMySQLDDL(ctx context.Context, ddl string) (*Table, error) {
	ddl = fmtMySQLDDLRE.ReplaceAllString(ddl, "")
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
	// 获取PK信息
	pkFields := make(map[string]bool)
	for _, constraint := range stmt.Constraints {
		if constraint.Tp != ast.ConstraintPrimaryKey {
			continue
		}
		for _, key := range constraint.Keys {
			pkFields[key.Column.Name.O] = true
		}
	}
	ret := d.parseMySQLDDLTable(stmt)
	for _, col := range stmt.Cols {
		field := d.parseMySQLDDLField(ctx, col)
		// 补充pk以及autoincr信息
		if pkFields[field.Name] {
			field.IsPK = true
		}
		for _, colOpt := range col.Options {
			if colOpt.Tp == ast.ColumnOptionAutoIncrement {
				field.AutoIncr = true
			}
		}
		ret.Fields = append(ret.Fields, field)
	}
	return ret, nil
}

var mysqlTypeRE = regexp.MustCompile(`^(\w+)\(\d+\)$`)

func (d *DefaultMetaCenter) parseMySQLDDLField(ctx context.Context, col *ast.ColumnDef) *Field {
	// 解析字段英文名
	field := &Field{
		Name: col.Name.Name.O,
	}
	// 解析字段类型
	tp := col.Tp.CompactStr() // int(11)
	matches := mysqlTypeRE.FindStringSubmatch(tp)
	if len(matches) > 1 {
		tp = matches[1]
	}
	field.Type = d.dataTypeGetter.GetByName(ctx, tp).ID
	// 解析字段注释，尝试解析字段的中文名
	// 以及如果有枚举值解析为枚举类型，否则如果是字符串类型且包含JSON字样解析为JSON
	for _, option := range col.Options {
		if option.Tp == ast.ColumnOptionComment {
			buf := bytes.NewBuffer(nil)
			option.Expr.Format(buf)
			comment := strings.Trim(buf.String(), `"`)
			name, enumKV := d.tryParseEnumFromComment(comment)
			field.CName = name
			if len(enumKV) == 0 {
				if field.Type == d.dataTypeGetter.GetByName(ctx, DataTypeString).ID &&
					d.tryParseJSONFromComment(comment) {
					field.Type = d.dataTypeGetter.GetByName(ctx, DataTypeJSON).ID
				}
				continue
			}
			field.Enum = &Enum{
				CName:      name,
				DataTypeID: field.Type,
			}
			for k, v := range enumKV {
				field.Enum.Values = append(field.Enum.Values, &EnumValue{
					EnumID: field.Enum.ID,
					EName:  strcase.ToCamel(field.Name) + strcase.ToCamel(k),
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

func (*DefaultMetaCenter) tryParseJSONFromComment(comment string) bool {
	lowerComment := strings.ToLower(comment)
	return strings.Contains(lowerComment, "json")
}

var (
	// firstLevelEnumParseRE 提取名称和枚举部分
	firstLevelEnumParseRE = regexp.MustCompile(`^(.*?)\s+(.*)$`)
	// secondLevelEnumDashRE 尝试匹配格式：1-待处理 2-处理中 等 或 A-选项A B-选项B 等
	secondLevelEnumDashRE = regexp.MustCompile(`(\w+)-([^\s]+)`)
	// secondLevelEnumColon1RE 尝试匹配格式：1：待处理 2：处理中 等 或 A：选项A B：选项B 等
	secondLevelEnumColon1RE = regexp.MustCompile(`(\w+)：([^\s]+)`)
	// secondLevelEnumColon2RE 尝试匹配格式：1:待处理 2:处理中 等 或 A:选项A B:选项B 等
	secondLevelEnumColon2RE = regexp.MustCompile(`(\w+):([^\s]+)`)
)

// tryParseEnumFromComment 将以下格式的注释解析为枚举信息，当前只支持数字枚举
// 任务状态 1-待处理 2-处理中 3-成功 4-失败
// 任务状态 1：待处理 2：处理中 3：成功 4：失败
// 任务状态 1:待处理 2:处理中 3:成功 4:失败
func (*DefaultMetaCenter) tryParseEnumFromComment(comment string) (string, map[string]string) {
	comment = strings.TrimSpace(comment)
	// 使用正则表达式提取名称和枚举部分
	matches := firstLevelEnumParseRE.FindStringSubmatch(comment)
	if len(matches) < 3 {
		return comment, map[string]string{}
	}
	name := strings.TrimSpace(matches[1])
	enumPart := matches[2]
	// 创建枚举映射
	enumMap := make(map[string]string)
	var pairs [][]string
	// 尝试不同的分隔符模式
	if secondLevelEnumDashRE.MatchString(enumPart) {
		pairs = secondLevelEnumDashRE.FindAllStringSubmatch(enumPart, -1)
	} else if secondLevelEnumColon1RE.MatchString(enumPart) {
		pairs = secondLevelEnumColon1RE.FindAllStringSubmatch(enumPart, -1)
	} else if secondLevelEnumColon2RE.MatchString(enumPart) {
		pairs = secondLevelEnumColon2RE.FindAllStringSubmatch(enumPart, -1)
	} else {
		return name, map[string]string{}
	}
	// 将匹配的键值对添加到映射中
	for _, pair := range pairs {
		if len(pair) >= 3 {
			key := strings.TrimSpace(pair[1])
			value := strings.TrimSpace(pair[2])
			enumMap[key] = value
		}
	}
	if len(enumMap) == 0 {
		return name, map[string]string{}
	}
	return name, enumMap
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
