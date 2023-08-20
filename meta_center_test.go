package metacenter

import (
	"context"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	_ "github.com/pingcap/tidb/types/parser_driver"
)

func TestDefaultMetaCenter_GenerateGoFiles(t *testing.T) {
	type fields struct {
		tableGetter      TableGetter
		tableFieldGetter TableFieldGetter
		fieldGetter      FieldGetter
		enumGetter       EnumGetter
		enumValueGetter  EnumValueGetter
		dataTypeGetter   DataTypeGetter
	}
	type args struct {
		ctx context.Context
	}
	tableGetter := NewDefaultTableGetter()
	tableFieldGetter := NewDefaultTableFieldGetter()
	fieldGetter := NewDefaultFieldGetter()
	enumGetter := NewDefaultEnumGetter()
	enumValueGetter := NewDefaultEnumValueGetter()
	dataTypeGetter := NewDefaultDataTypeGetter()
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"success",
			fields{
				tableGetter:      tableGetter,
				tableFieldGetter: tableFieldGetter,
				fieldGetter:      fieldGetter,
				enumGetter:       enumGetter,
				enumValueGetter:  enumValueGetter,
				dataTypeGetter:   dataTypeGetter,
			},
			args{
				ctx: context.Background(),
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DefaultMetaCenter{
				tableGetter:      tt.fields.tableGetter,
				tableFieldGetter: tt.fields.tableFieldGetter,
				fieldGetter:      tt.fields.fieldGetter,
				enumGetter:       tt.fields.enumGetter,
				enumValueGetter:  tt.fields.enumValueGetter,
				dataTypeGetter:   tt.fields.dataTypeGetter,
			}
			p0 := gomonkey.ApplyMethodFunc(d, "GetAllTables", func(ctx context.Context) []*Table {
				return []*Table{
					{
						ID:    1,
						Name:  "t_test",
						CName: "测试表",
						Fields: []*Field{
							{
								ID:    1,
								Name:  "id",
								CName: "自增ID",
								Type:  3,
							},
							{
								ID:     2,
								Name:   "task_status",
								CName:  "任务状态",
								EnumID: 1,
								Type:   1,
								Enum: &Enum{
									ID:         1,
									CName:      "通用状态",
									DataTypeID: 3,
									Values: []*EnumValue{
										{
											ID:     1,
											EnumID: 1,
											EName:  "wait",
											Desc:   "待执行",
											Value:  "1",
										},
										{
											ID:     2,
											EnumID: 1,
											EName:  "exec",
											Desc:   "执行中",
											Value:  "2",
										},
										{
											ID:     3,
											EnumID: 1,
											EName:  "finish",
											Desc:   "已完成",
											Value:  "3",
										},
									},
								},
							},
							{
								ID:     3,
								Name:   "phase",
								CName:  "测试任务阶段",
								Type:   1,
								EnumID: 2,
								Enum: &Enum{
									ID:         2,
									CName:      "任务阶段",
									DataTypeID: 2,
									Values: []*EnumValue{
										{
											ID:     4,
											EnumID: 1,
											EName:  "parse_file",
											Desc:   "解析文件",
											Value:  "parse_file",
										},
										{
											ID:     5,
											EnumID: 1,
											EName:  "collect_data",
											Desc:   "收集数据",
											Value:  "collect_data",
										},
										{
											ID:     6,
											EnumID: 1,
											EName:  "send_file",
											Desc:   "发送文件",
											Value:  "send_file",
										},
									},
								},
							},
							{
								ID:    4,
								Name:  "operator",
								CName: "操作人",
								Type:  2,
							},
						},
					},
				}
			})
			defer p0.Reset()
			p1 := gomonkey.ApplyMethodFunc(d.dataTypeGetter, "GetByID", func(ctx context.Context, id int) *DataType {
				if id == 1 {
					return &DataType{Name: DataTypeEnum}
				}
				if id == 2 {
					return &DataType{Name: DataTypeString}
				}
				return &DataType{Name: DataTypeInt}
			})
			defer p1.Reset()
			if err := d.GenerateGoFiles(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("DefaultMetaCenter.GenerateGoFiles() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultMetaCenter_ParseFromMySQLDDL(t *testing.T) {
	type fields struct {
		tableGetter      TableGetter
		tableFieldGetter TableFieldGetter
		fieldGetter      FieldGetter
		enumGetter       EnumGetter
		enumValueGetter  EnumValueGetter
		dataTypeGetter   DataTypeGetter
	}
	type args struct {
		ctx context.Context
		ddl string
	}
	tableGetter := NewDefaultTableGetter()
	tableFieldGetter := NewDefaultTableFieldGetter()
	fieldGetter := NewDefaultFieldGetter()
	enumGetter := NewDefaultEnumGetter()
	enumValueGetter := NewDefaultEnumValueGetter()
	dataTypeGetter := NewDefaultDataTypeGetter()
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Table
		wantErr bool
	}{
		{
			"success",
			fields{
				tableGetter:      tableGetter,
				tableFieldGetter: tableFieldGetter,
				fieldGetter:      fieldGetter,
				enumGetter:       enumGetter,
				enumValueGetter:  enumValueGetter,
				dataTypeGetter:   dataTypeGetter,
			},
			args{
				ctx: context.Background(),
				ddl: "CREATE TABLE `t` (" +
					"`id` int NOT NULL AUTO_INCREMENT COMMENT 'pk-id'," +
					"`s` char(60) DEFAULT NULL COMMENT 'testcomment'," +
					"PRIMARY KEY (`id`)," +
					"UNIQUE KEY (`id`,`s`)" +
					") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT 'tabletestcomment'",
			},
			&Table{
				Name:  "t",
				CName: "tabletestcomment",
				Fields: []*Field{
					{
						Name:  "id",
						CName: "pk-id",
						Type:  1,
					},
					{
						Name:  "s",
						CName: "testcomment",
						Type:  2,
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DefaultMetaCenter{
				tableGetter:      tt.fields.tableGetter,
				tableFieldGetter: tt.fields.tableFieldGetter,
				fieldGetter:      tt.fields.fieldGetter,
				enumGetter:       tt.fields.enumGetter,
				enumValueGetter:  tt.fields.enumValueGetter,
				dataTypeGetter:   tt.fields.dataTypeGetter,
			}
			p0 := gomonkey.ApplyMethodFunc(tt.fields.dataTypeGetter, "GetByName", func(ctx context.Context, name string) *DataType {
				if name == "string" {
					return &DataType{ID: 2}
				}
				if name == "int" {
					return &DataType{ID: 1}
				}
				return &DataType{}
			})
			defer p0.Reset()
			got, err := d.ParseFromMySQLDDL(tt.args.ctx, tt.args.ddl)
			if (err != nil) != tt.wantErr {
				t.Errorf("DefaultMetaCenter.ParseFromMySQLDDL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DefaultMetaCenter.ParseFromMySQLDDL() = %v, want %v", got, tt.want)
			}
		})
	}
}
