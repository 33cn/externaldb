package parser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"strings"

	sliceUtil "github.com/33cn/go-kit/slice"

	"github.com/33cn/externaldb/tools/swag2md/pkg/builder"
	"github.com/33cn/externaldb/tools/swag2md/types"
)

const (
	methodGet         = "GET"
	schemaReferPrefix = "#/definitions/"
	// propertyNameData  = "data"
	paramRequired    = "必填"
	paramNotRequired = "非必填"
	reqFormat        = "| %s | %s | %s | %s | %s | %s |\n"
	respFormat       = "| %s | %s | %s | %s |\n"
	emsp             = "&emsp;"
)

// Parser Swagger解析器详情
type Parser struct {
	swagger   *types.Swagger
	keys      []string
	pathGroup map[string][]*types.PathInfo
	num       int
}

// NewParser 新建Swagger解析器
func NewParser(filePath string) (*Parser, error) {
	if filePath == "" {
		return nil, errors.New("illegal parser configure")
	}

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	swagger := &types.Swagger{}
	err = json.Unmarshal(content, swagger)
	if err != nil {
		return nil, err
	}

	pathInfos := ParsePathInfos(swagger.Paths)
	keys, pathGroup := ParsePathGroup(pathInfos)

	return &Parser{
		swagger:   swagger,
		keys:      keys,
		pathGroup: pathGroup,
		num:       len(pathInfos),
	}, nil
}

// BuildTitle 构建标题
func (p *Parser) BuildTitle(title string) string {
	return fmt.Sprintf("# %s\n", title)
}

// BuildOverview 构建概览
func (p *Parser) BuildOverview() string {
	b := &strings.Builder{}
	b.WriteString(fmt.Sprintf("\n## 接口概览（总计%d个）\n", p.num))

	for _, key := range p.keys {
		b.WriteString(fmt.Sprintf("\n### %s\n", key))
		b.WriteString("\n| **路径** | **功能** | **请求方式** |\n")
		b.WriteString("|---------|---------|-------------|\n")

		if pathInfos, ok := p.pathGroup[key]; ok {
			for _, pi := range pathInfos {
				b.WriteString(fmt.Sprintf("| [%s](#%s) | %s | %s |\n",
					pi.Path, pi.Summary, pi.Summary, pi.Method))
			}
		}
	}

	return b.String()
}

// BuildDetail 构建详情
func (p *Parser) BuildDetail() string {
	b := &strings.Builder{}
	b.WriteString("\n## 接口详情\n")

	for _, key := range p.keys {
		i := strings.Index(key, " ")
		if i == -1 {
			i = len(key)
		}

		b.WriteString(fmt.Sprintf("\n### %s\n", key[:i]))

		if pathInfos, ok := p.pathGroup[key]; ok {
			for _, pi := range pathInfos {
				b.WriteString(fmt.Sprintf("\n### %s\n", pi.Summary))
				b.WriteString(fmt.Sprintf("\n[返回概览](#%s)\n", strings.ReplaceAll(key, " ", "-")))
				if pi.Method == methodGet {
					b.WriteString(fmt.Sprintf("\n%s %s\n", pi.Method, pi.Path))
				} else {
					b.WriteString(fmt.Sprintf("\n%s %s  \nContent-Type: %s\n", pi.Method, pi.Path, "application/json"))
				}
				// 构建请求参数
				p.buildParameters(b, pi)
				// 构建响应参数
				p.buildResponses(b, pi)
			}
		}
	}

	return b.String()
}

func (p *Parser) BuildDefine() string {
	b := &strings.Builder{}
	b.WriteString("\n## 类型定义\n\n")
	keys := make([]string, 0, len(p.swagger.Definitions))
	for k := range p.swagger.Definitions {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b.WriteString("### " + k)
		b.WriteString("\n\n| **参数** | **描述** | **类型** | **说明** |\n")
		b.WriteString("|----------|----------|----------|----------|\n")
		v := p.swagger.Definitions[k]
		ps := types.NewPropertySorter(v.Properties)
		for _, pt := range ps {
			b.WriteString(fmt.Sprintf(respFormat, pt.Name,
				convertDescription(pt.Schema.Description, true),
				convertType(pt.Schema),
				convertDescription(pt.Schema.Description, false),
			))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// buildParameters 构建请求参数
func (p *Parser) buildParameters(b io.StringWriter, pi *types.PathInfo) {
	if len(pi.Parameters) > 0 {
		eb := builder.NewExampleBuilder(pi.Path, false)

		b.WriteString("\n请求参数：\n")
		b.WriteString("\n| **来源** | **参数** | **描述** | **类型** | **约束** | **说明** |\n")
		b.WriteString("|----------|----------|----------|----------|----------|----------|\n")

		for _, param := range pi.Parameters {
			if param.Schema == nil {
				limit := paramNotRequired
				if param.Required {
					limit = paramRequired
				}

				b.WriteString(fmt.Sprintf(reqFormat, param.In, param.Name,
					convertDescription(param.Description, true),
					param.Type, limit,
					convertDescription(param.Description, false),
				))

				switch param.In {
				case types.ParameterInQuery:
					eb.AddQuery(param.Name, param.Type)
				case types.ParameterInFormData:
					eb.AddForm(param.Name, param.Type)
				case types.ParameterInHeader:
					eb.AddHeader(param.Name, param.Type)
				}
			} else {
				buildSchemaAllOf(param.Schema)
				ref, isArray := getRefAndIsArray(param.Schema)
				eb.SetIsArray(isArray)
				dk := strings.TrimPrefix(ref, schemaReferPrefix)

				if d, ok := p.swagger.Definitions[dk]; ok && d.Type == types.SchemaTypeObject {
					ps := types.NewPropertySorter(types.CombineProperties(d.Properties, param.Schema.Properties))
					rm := sliceUtil.CountString(d.Required)
					for _, property := range ps {
						limit := paramNotRequired
						if _, ok := rm[property.Name]; ok {
							limit = paramRequired
						}

						b.WriteString(fmt.Sprintf(reqFormat, param.In, property.Name,
							convertDescription(property.Schema.Description, true),
							convertType(property.Schema), limit,
							convertDescription(property.Schema.Description, false),
						))

						pp := p.buildProperty(b, property, param.In, ref, 1)
						if pp == "" {
							eb.AddJSON(property.Name, property.Schema.Example, property.Schema.Type)
						} else {
							eb.AddJSONString(property.Name, pp, property.Schema.Type)
						}
					}
				}
			}
		}

		es := eb.String()
		if es != "" {
			b.WriteString("\n请求示例：\n")
			b.WriteString(es)
		}
	}
}

func buildSchemaAllOf(schema *types.Schema) {
	for _, s := range schema.AllOf {
		if s.AllOf != nil {
			for _, v := range s.AllOf {
				buildSchemaAllOf(v)
			}
		}
		if s.Ref != "" {
			schema.Ref = s.Ref
		}
		if s.Type != "" {
			schema.Type = s.Type
		}
		if s.Properties != nil {
			if schema.Properties == nil {
				schema.Properties = s.Properties
			} else {
				for k, v := range s.Properties {
					schema.Properties[k] = v
				}
			}
		}
		if s.Description != "" {
			schema.Description = s.Description
		}
		if s.Items != nil {
			schema.Items = s.Items
		}
	}
}

// buildResponses 构建响应参数
func (p *Parser) buildResponses(b io.StringWriter, pi *types.PathInfo) {
	if len(pi.Responses) > 0 {
		eb := builder.NewExampleBuilder(pi.Path, true)

		b.WriteString("\n响应参数：\n")
		b.WriteString("\n| **参数** | **描述** | **类型** | **说明** |\n")
		b.WriteString("|----------|----------|----------|----------|\n")

		if resp, ok := pi.Responses[200]; ok {
			var ref string
			if resp.Schema != nil {
				buildSchemaAllOf(resp.Schema)
			}
			if resp.Schema != nil && resp.Schema.Ref != "" {
				ref, _ = getRefAndIsArray(resp.Schema)
				eb.SetNeedWrap(false)
			}
			if ref != "" {
				dk := strings.TrimPrefix(ref, schemaReferPrefix)
				if d, ok := p.swagger.Definitions[dk]; ok && d.Type == types.SchemaTypeObject {
					var pro map[string]*types.Schema
					if resp.Schema != nil {
						pro = resp.Schema.Properties
					}
					ps := types.NewPropertySorter(types.CombineProperties(d.Properties, pro))
					for _, property := range ps {
						b.WriteString(fmt.Sprintf(respFormat, property.Name,
							convertDescription(property.Schema.Description, true),
							convertType(property.Schema),
							convertDescription(property.Schema.Description, false),
						))

						pp := p.buildProperty(b, property, "", ref, 1)
						if pp == "" {
							eb.AddJSON(property.Name, property.Schema.Example, property.Schema.Type)
						} else {
							eb.AddJSONString(property.Name, pp, property.Schema.Type)
						}
					}
				}
			}
		}

		es := eb.String()
		if es != "" {
			b.WriteString("\n响应示例：\n")
			b.WriteString(es)
		}
	}
}

// buildProperty 构建属性参数
func (p *Parser) buildProperty(b io.StringWriter, pty *types.Property, paramIn, parentRef string, depth int) string {
	if pty.Schema.Type == types.SchemaTypeArray && pty.Schema.Items.Type != "" {
		return builder.GetArrayString(pty.Name, pty.Schema.Example, pty.Schema.Items.Type)
	}

	ref, isArray := getRefAndIsArray(pty.Schema)
	if parentRef == ref {
		return ""
	}

	if ref != "" {
		eb := builder.NewExampleBuilder("", false)
		dk := strings.TrimPrefix(ref, schemaReferPrefix)

		if d, ok := p.swagger.Definitions[dk]; ok && d.Type == types.SchemaTypeObject {
			ps := types.NewPropertySorter(d.Properties)
			rm := sliceUtil.CountString(d.Required)

			for _, property := range ps {
				if depth <= 1 {
					prefix := strings.Repeat(emsp, depth) + " "
					if paramIn != "" {
						limit := paramNotRequired
						if _, ok := rm[property.Name]; ok {
							limit = paramRequired
						}
						b.WriteString(fmt.Sprintf(reqFormat, paramIn, prefix+property.Name,
							convertDescription(property.Schema.Description, true),
							convertType(property.Schema), limit,
							convertDescription(property.Schema.Description, false),
						))
					} else {
						b.WriteString(fmt.Sprintf(respFormat, prefix+property.Name,
							convertDescription(property.Schema.Description, true),
							convertType(property.Schema),
							convertDescription(property.Schema.Description, false),
						))
					}
				}
				pp := p.buildProperty(b, property, paramIn, ref, depth+1)
				if pp == "" {
					eb.AddJSON(property.Name, property.Schema.Example, property.Schema.Type)
				} else {
					eb.AddJSONString(property.Name, pp, property.Schema.Type)
				}
			}
			// 处理map
			//if d.AdditionalProperties != nil {
			//	eb := builder.NewExampleBuilder("", false)
			//	property := &types.Property{Name: "map_key01", Schema: d.AdditionalProperties}
			//	fmt.Println("property",property)
			//	pp := p.buildProperty(b, property, paramIn, ref, depth+1)
			//	if pp == "" {
			//		eb.AddJSON(property.Name, property.Schema.Example, property.Schema.Type)
			//	} else {
			//		eb.AddJSONString(property.Name, pp, property.Schema.Type)
			//	}
			//	jsonString := eb.GetJSON()
			//	return fmt.Sprintf("{\"map_key01\":%s}", jsonString)
			//}
		}

		jsonString := eb.GetJSON()
		if isArray {
			return fmt.Sprintf("[%s]", jsonString)
		}
		return jsonString
	}

	return ""
}

// ParsePathInfos 解析路径信息列表
func ParsePathInfos(paths map[string]map[string]*types.Operation) []*types.PathInfo {
	var pathInfos []*types.PathInfo

	for path, methodOperation := range paths {
		for method, operation := range methodOperation {
			pi := &types.PathInfo{
				Method:     strings.ToUpper(method),
				Path:       path,
				Summary:    operation.Summary,
				Parameters: operation.Parameters,
				Responses:  operation.Responses,
			}

			if len(operation.Tags) > 0 {
				pi.Tag = strings.TrimSpace(operation.Tags[0])
			}

			if len(operation.Security) > 0 {
				pi.NeedAuth = true
			}

			if len(operation.Consumes) > 0 {
				pi.Consume = strings.TrimSpace(operation.Consumes[0])
			}

			if len(operation.Produces) > 0 {
				pi.Produce = strings.TrimSpace(operation.Produces[0])
			}

			pathInfos = append(pathInfos, pi)
		}
	}

	sort.Slice(pathInfos, func(i, j int) bool {
		pi, pj := pathInfos[i], pathInfos[j]
		if pi.Path != pj.Path {
			return pi.Path < pj.Path
		}
		li, lj := len(pi.Method), len(pj.Method)
		if li != lj {
			return li < lj
		}
		return pi.Method < pj.Method
	})

	return pathInfos
}

// ParsePathGroup 解析路径分组
func ParsePathGroup(pathInfos []*types.PathInfo) ([]string, map[string][]*types.PathInfo) {
	var keys []string
	m := make(map[string][]*types.PathInfo)

	for _, pi := range pathInfos {
		if _, ok := m[pi.Tag]; !ok {
			keys = append(keys, pi.Tag)
		}

		m[pi.Tag] = append(m[pi.Tag], pi)
	}
	sort.Strings(keys)

	return keys, m
}

// getRefAndIsArray 获取概要引用并判断其是否为数组
func getRefAndIsArray(schema *types.Schema) (ref string, isArray bool) {
	if schema.Type == types.SchemaTypeArray && schema.Items.Ref != "" {
		ref = schema.Items.Ref
		isArray = true
	} else if schema.Type == types.SchemaTypeObject || schema.Type == "" {
		ref = schema.Ref
		isArray = false
	}

	return
}

// convertType 转换类型
func convertType(schema *types.Schema) string {
	if schema.Ref != "" {
		return TrimTypePrefix(schema.Ref)
	}
	if schema.AdditionalProperties != nil {
		return "map\\[string\\] " + convertType(schema.AdditionalProperties)
	}
	switch t := schema.Type; t {
	case "":
		return types.SchemaTypeObject
	case types.SchemaTypeArray:
		if it := schema.Items.Type; it != "" {
			return it + " " + types.SchemaTypeArray
		}
		if schema.Items.Ref != "" {
			return TrimTypePrefix(schema.Items.Ref) + " " + types.SchemaTypeArray
		}
		return types.SchemaTypeObject + " " + types.SchemaTypeArray
	}
	if schema.Type != "" {
		return schema.Type
	}
	return types.SchemaTypeObject
}

func TrimTypePrefix(ty string) string {
	ans := strings.TrimPrefix(ty, schemaReferPrefix)
	return fmt.Sprintf("[%s](#%s)", ans, strings.ReplaceAll(ans, ".", ""))
}

// convertDescription 转换描述
func convertDescription(d string, flag bool) string {
	left, right := d, ""
	start := strings.Index(d, "（")
	end := strings.LastIndex(d, "）")
	if start != -1 && end != -1 {
		left, right = d[:start], d[start+3:end]
	}
	if flag {
		return left
	}
	return right
}
