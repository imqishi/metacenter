// Package {{.PkgName}} {{.PkgName}}的相关常量定义
package {{.PkgName}}

// 表名称定义
const (
    // TableName{{.Table.VarName}} {{.Table.CName}}表英文名
    TableName{{.Table.VarName}} = "{{.Table.Name}}"
    // TableCName{{.Table.VarName}} {{.Table.Name}}表中文名
    TableCName{{.Table.VarName}} = "{{.Table.CName}}"
)

// 表字段定义
const (
    {{- range .Fields}}
    // Field{{.VarName}} 字段-{{.CName}}
    Field{{.VarName}} = "{{.Name}}"
    {{- end}}
)

{{if .HasEnum}}
// 表枚举字段定义
    {{- range $index, $field := .Fields}}
    {{if .IsEnum}}
// {{$field.VarName}}-{{$field.CName}}枚举定义
const (
        {{- range .EnumValues}}
    // {{$field.VarName}}{{.VarName}} {{$field.CName}}-{{.CName}}
            {{- if $field.IsNum}}
    {{$field.VarName}}{{.VarName}} = {{.Value}}
            {{- else}}
    {{$field.VarName}}{{.VarName}} = "{{.Value}}"
            {{- end}}
        {{- end}}
)
    {{end}}
    {{- end}}
{{end}}
