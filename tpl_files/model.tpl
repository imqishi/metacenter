// Package {{.PkgName}} {{.PkgName}}的模型定义
package {{.PkgName}}

// {{.Table.VarName}} {{.Table.CName}}
type {{.Table.VarName}} struct {
    {{- range .Fields}}
    // {{.VarName}} {{.CName}}
    {{.VarName}} {{.Type}} `json:"{{.Name}}" xorm:"'{{.Name}}'" gorm:"column:{{.Name}}"`
    {{- end}}
}

// TableName 实现orm接口
func (*{{.Table.VarName}}) TableName() string {
    return "{{.Table.Name}}"
}
