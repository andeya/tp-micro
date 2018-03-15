package create

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/henrylee2cn/ant"
	"github.com/henrylee2cn/ant/cmd/ant/info"
	"github.com/henrylee2cn/goutil"
)

const (
	TOKEN_MIN                  = iota
	TOKEN_ASSIGN               // =
	TOKEN_KEYWORD_TYPE         // type
	TOKEN_KEYWORD_API          // api
	TOKEN_KEYWORD_PULL         // pull
	TOKEN_KEYWORD_PUSH         // push
	TOKEN_KEYWORD_STRING       // string
	TOKEN_KEYWORD_INT          // int
	TOKEN_KEYWORD_LONG         // long
	TOKEN_KEYWORD_FLOAT        // float
	TOKEN_KEYWORD_DOUBLE       // double
	TOKEN_KEYWORD_LIST         // list
	TOKEN_KEYWORD_MAP          // map
	TOKEN_BRACE_LEFT           // {
	TOKEN_BRACE_RIGHT          // }
	TOKEN_BRACKETS_LEFT        // (
	TOKEN_BRACKETS_RIGHT       // )
	TOKEN_ANGLE_BRACKETS_LEFT  // <
	TOKEN_ANGLE_BRACKETS_RIGHT // >
	TOKEN_COLON                // :
	TOKEN_COMMA                // ,
	TOKEN_RETURN               // ->
	TOKEN_COMMENT              // //
	TOKEN_TYPE_TAG             // 'xx'
	TOKEN_SYMBOL
	TOKEN_MAX
)

var token_rules = map[int]string{
	//TOKEN_ASSIGN : 					"=",
	TOKEN_COMMENT:              `\s*//\s*`,
	TOKEN_KEYWORD_TYPE:         `^\s*type\s+`,
	TOKEN_KEYWORD_API:          `^\s*api\s+`,
	TOKEN_KEYWORD_PULL:         `^\s*pull\s+`,
	TOKEN_KEYWORD_PUSH:         `^\s*push\s+`,
	TOKEN_COLON:                `\s*:\s*`,
	TOKEN_COMMA:                `\s*\,\s*`,
	TOKEN_RETURN:               `\s*->\s*`,
	TOKEN_BRACE_LEFT:           `\s*\{\s*`,
	TOKEN_BRACE_RIGHT:          `\s*\}\s*`,
	TOKEN_BRACKETS_LEFT:        `\s*\(\s*`,
	TOKEN_BRACKETS_RIGHT:       `\s*\)\s*`,
	TOKEN_ANGLE_BRACKETS_LEFT:  `\s*<\s*`,
	TOKEN_ANGLE_BRACKETS_RIGHT: `\s*>\s*`,
	TOKEN_KEYWORD_STRING:       `\s*string\s*`,
	TOKEN_KEYWORD_INT:          `\s*int32\s*`,
	TOKEN_KEYWORD_LONG:         `\s*int64\s*`,
	TOKEN_KEYWORD_FLOAT:        `\s*float32\s*`,
	TOKEN_KEYWORD_DOUBLE:       `\s*float64\s*`,
	TOKEN_KEYWORD_LIST:         `\s*list\s*`,
	TOKEN_KEYWORD_MAP:          `\s*map\s*`,
	TOKEN_TYPE_TAG:             "\\s*`.+`\\s*",
	TOKEN_SYMBOL:               `\s*[\w]+\s*`,
}

type Token struct {
	lineno    int
	tokenType int
	text      string
}

type Lexer struct {
	lines        []string
	rules        map[int]*regexp.Regexp
	tokens       []*Token
	currTokenIdx int
}

func (lexer *Lexer) init(tplReader io.Reader) {
	lexer.currTokenIdx = 0
	lexer.rules = map[int]*regexp.Regexp{}
	for k, v := range token_rules {
		reg := regexp.MustCompile(v)
		lexer.rules[k] = reg
	}

	scanner := bufio.NewScanner(tplReader)
	for scanner.Scan() {
		line := scanner.Text()
		lexer.lines = append(lexer.lines, line)
	}
	for k, v := range lexer.lines {
		lexer.parseLine(k+1, v)
	}

	//for _, v := range lexer.tokens {
	//	fmt.Println("line ", v.lineno,  v.text)
	//}
}

func (lexer *Lexer) parseLine(lineno int, lineText string) {
	line := lineText
	for len(line) > 0 {
		isMatch := false
		for tokenType := TOKEN_MIN + 1; tokenType < TOKEN_MAX; tokenType++ {
			reg := lexer.rules[tokenType]
			if reg == nil {
				continue
			}
			ret := reg.FindStringIndex(line)
			if len(ret) == 2 && ret[0] == 0 {
				// do not process comment words
				if tokenType == TOKEN_COMMENT {
					return
				}
				text := strings.TrimSpace(line[ret[0]:ret[1]])
				lexer.tokens = append(lexer.tokens, &Token{lineno: lineno, tokenType: tokenType, text: text})
				line = line[ret[1]:]
				isMatch = true
				break
			}
		}
		if isMatch == false {
			ant.Fatalf("[ant] There is some error in config file with line %v: %v", lineno, line)
			break
		}
	}
}

func (lexer *Lexer) takeToken() *Token {
	if lexer.currTokenIdx >= len(lexer.tokens) {
		return nil
	}
	ret := lexer.tokens[lexer.currTokenIdx]
	lexer.currTokenIdx++
	//fmt.Println("takeToken: ", ret.text)
	return ret
}

func (lexer *Lexer) nextTokenType() int {
	return lexer.tokens[lexer.currTokenIdx].tokenType
}

/*
BNF Design:
	<type_define> ::= type <identifier> { <variable_declare> }
	<api_define> ::= api pull handlerName { <api_declare> } | api push handlerName { <api_declare> }
	<identifier> ::= "\w+"
	<type> ::= int32 | int64 | float32 | float64 | string | list<<type>> | map<<type>, <type>> | <identifier>
	<variable_declare> ::= <identifier> : <type> | <variable_declare>
	<parameter_declare> ::= <identifier> : <type> | ,<parameter_declare>
	<api_declare> ::= <identifier> (<parameter_declare>) -> <type>
*/
type Parser struct {
	lexer *Lexer
	types []*CustomType
	apis  []*CustomAPI
}

type Variable struct {
	variableName string
	variableType string
	other1       string // used by list or map
	other2       string // used by map
	tag          string // type member tag
}

type CustomType struct {
	typeName string
	members  []*Variable
}

type CustomAPI struct {
	access string // push or pull
	group  string
	name   string
	params []*Variable
	ret    string
}

func (parser *Parser) init(lexer *Lexer) {
	parser.lexer = lexer
}

func (parser *Parser) getCustomType(typeName string) *CustomType {
	if typeName == "string" || typeName == "int32" || typeName == "int64" || typeName == "float32" || typeName == "float64" {
		return nil
	}
	for _, v := range parser.types {
		if v.typeName == typeName {
			return v
		}
	}
	return nil
}

func (parser *Parser) parse() {
	for {
		token := parser.lexer.takeToken()
		if token == nil {
			break
		}
		switch token.tokenType {
		case TOKEN_KEYWORD_TYPE:
			parser.type_define()
		case TOKEN_KEYWORD_API:
			parser.api_define()
		}
	}
	parser.printTypes()
	parser.printApis()
}

func (parser *Parser) removeComments() {
}

func (parser *Parser) type_define() {
	identifier := parser.lexer.takeToken()
	//fmt.Println("\n------ type define begin ------")
	//fmt.Println("type: ", identifier.text)
	if identifier.tokenType != TOKEN_SYMBOL {
		ant.Fatalf("[ant] invaid identifier in line %v: %v", identifier.lineno, identifier.text)
	}

	braceLeft := parser.lexer.takeToken()
	if braceLeft.tokenType != TOKEN_BRACE_LEFT {
		ant.Fatalf("[ant] expect '{' in line %v: %v", braceLeft.lineno, braceLeft.text)
	}
	members := []*Variable{}
	parser.variable_declare(&members)
	braceRight := parser.lexer.takeToken()
	if braceRight.tokenType != TOKEN_BRACE_RIGHT {
		ant.Fatalf("[ant] expect '}' in line %v: %v", braceRight.lineno, braceRight.text)
	}
	//fmt.Println("------------ end ----------\n")
	customType := &CustomType{typeName: identifier.text, members: members}
	parser.types = append(parser.types, customType)
}

func (parser *Parser) variable_declare(members *[]*Variable) {
	if tokenType := parser.lexer.nextTokenType(); tokenType == TOKEN_BRACE_RIGHT {
		return
	}
	identifier := parser.lexer.takeToken()
	if identifier.tokenType != TOKEN_SYMBOL {
		ant.Fatalf("[ant] invaid identifier in line %v: %v", identifier.lineno, identifier.text)
	}
	colon := parser.lexer.takeToken()
	if colon.tokenType != TOKEN_COLON {
		ant.Fatalf("[ant] expect ':' in line %v: %v", colon.lineno, colon.text)
	}
	variableName := ""
	variableType := ""
	other1 := ""
	other2 := ""
	tag := ""
	type_identifier := parser.lexer.takeToken()
	if type_identifier.tokenType == TOKEN_KEYWORD_LIST {
		listTypeName := parser.parseListType()
		variableName = identifier.text
		variableType = type_identifier.text
		other1 = listTypeName
		other2 = ""
	} else if type_identifier.tokenType == TOKEN_KEYWORD_MAP {
		mapKeyTypeName, mapValueTypeName := parser.parseMapType()
		variableName = identifier.text
		variableType = type_identifier.text
		other1 = mapKeyTypeName
		other2 = mapValueTypeName
	} else {
		if parser.isDeclaredType(type_identifier.text) == false {
			ant.Fatalf("[ant] undeclared type in line %v: %v", type_identifier.lineno, type_identifier.text)
		}
		variableName = identifier.text
		variableType = type_identifier.text
		other1 = ""
		other2 = ""
	}

	// member tag
	if parser.lexer.nextTokenType() == TOKEN_TYPE_TAG {
		tagToken := parser.lexer.takeToken()
		tag = tagToken.text
	}
	*members = append(*members, &Variable{
		variableName: variableName,
		variableType: variableType,
		other1:       other1,
		other2:       other2,
		tag:          tag,
	})
	//fmt.Println("member: ", identifier.text, ":", type_identifier.text)

	// recursion for processing variable declare
	parser.variable_declare(members)
}

func (parser *Parser) api_define() {
	access := parser.lexer.takeToken()
	if access.tokenType != TOKEN_KEYWORD_PULL && access.tokenType != TOKEN_KEYWORD_PUSH {
		ant.Fatalf("[ant] expect access type 'pull' or 'push' in line %v: %v", access.lineno, access.text)
	}
	group := parser.lexer.takeToken()
	if group.tokenType != TOKEN_SYMBOL {
		ant.Fatalf("[ant] invaid group name dentifier in line %v: %v", group.lineno, group.text)
	}
	braceLeft := parser.lexer.takeToken()
	if braceLeft.tokenType != TOKEN_BRACE_LEFT {
		ant.Fatalf("[ant] expect '{' in line %v: %v", braceLeft.lineno, braceLeft.text)
	}
	parser.api_declare(access.text, group.text)
	braceRight := parser.lexer.takeToken()
	if braceRight.tokenType != TOKEN_BRACE_RIGHT {
		ant.Fatalf("[ant] expect '}' in line %v: %v", braceRight.lineno, braceRight.text)
	}
}

func (parser *Parser) api_declare(access string, group string) {
	if tokenType := parser.lexer.nextTokenType(); tokenType == TOKEN_BRACE_RIGHT {
		return
	}

	// 1. api function name
	function := parser.lexer.takeToken()
	if function.tokenType != TOKEN_SYMBOL {
		ant.Fatalf("[ant] invaid api name dentifier in line %v: %v", function.lineno, function.text)
	}
	function.text = strings.ToUpper(function.text[0:1]) + function.text[1:]

	// 2. function '('
	bracketsLeft := parser.lexer.takeToken()
	if bracketsLeft.tokenType != TOKEN_BRACKETS_LEFT {
		ant.Fatalf("[ant] expect '(' in line %v: %v", bracketsLeft.lineno, bracketsLeft.text)
	}

	// 3. function parameter
	params := []*Variable{}
	parser.parameter_declare(&params)

	// 4. function ')'
	bracketsRight := parser.lexer.takeToken()
	if bracketsRight.tokenType != TOKEN_BRACKETS_RIGHT {
		ant.Fatalf("[ant] expect ')' in line %v: %v", bracketsRight.lineno, bracketsRight.text)
	}

	// 5. function -> return type
	if access == "pull" {
		retTagToken := parser.lexer.takeToken()
		if retTagToken.tokenType != TOKEN_RETURN {
			ant.Fatalf("[ant] expect function return tag '->' in line %v: %v", retTagToken.lineno, retTagToken.text)
		}
		retToken := parser.lexer.takeToken()
		if parser.isDeclaredType(retToken.text) == false {
			ant.Fatalf("[ant] undefined ret type in line %v: %v", retToken.lineno, retToken.text)
		}
		api := &CustomAPI{access: access, group: group, name: function.text, params: params, ret: retToken.text}
		parser.apis = append(parser.apis, api)
	} else {
		api := &CustomAPI{access: access, group: group, name: function.text, params: params, ret: ""}
		parser.apis = append(parser.apis, api)
	}

	// recursion for processing other api declare
	parser.api_declare(access, group)
}

func (parser *Parser) parameter_declare(params *[]*Variable) {
	paramName := parser.lexer.takeToken()
	if paramName.tokenType != TOKEN_SYMBOL {
		ant.Fatalf("[ant] invaid parameter name dentifier in line %v: %v", paramName.lineno, paramName.text)
	}
	colon := parser.lexer.takeToken()
	if colon.tokenType != TOKEN_COLON {
		ant.Fatalf("[ant] expect ':' in line %v: %v", colon.lineno, colon.text)
	}
	paramType := parser.lexer.takeToken()
	if paramType.tokenType == TOKEN_KEYWORD_LIST {
		listTypeName := parser.parseListType()
		*params = append(*params, &Variable{
			variableName: paramName.text,
			variableType: paramType.text,
			other1:       listTypeName,
			other2:       "",
		})
	} else if paramType.tokenType == TOKEN_KEYWORD_MAP {
		mapKeyTypeName, mapValueTypeName := parser.parseMapType()
		*params = append(*params, &Variable{
			variableName: paramName.text,
			variableType: paramType.text,
			other1:       mapKeyTypeName,
			other2:       mapValueTypeName,
		})
	} else {
		if parser.isDeclaredType(paramType.text) == false {
			ant.Fatalf("[ant] undeclared param type in line %v: %v", paramType.lineno, paramType.text)
		}
		*params = append(*params, &Variable{
			variableName: paramName.text,
			variableType: paramType.text,
			other1:       "",
			other2:       "",
		})
	}
	if tokenType := parser.lexer.nextTokenType(); tokenType == TOKEN_COMMA {
		commaToken := parser.lexer.takeToken()
		if commaToken.tokenType != TOKEN_COMMA {
			ant.Fatalf("[ant] expect ',' in line %v: %v", paramType.lineno, paramType.text)
		}
		parser.parameter_declare(params)
	}
}

func (parser *Parser) parseListType() string {
	bracketsLeft := parser.lexer.takeToken() // '<'
	if bracketsLeft.tokenType != TOKEN_ANGLE_BRACKETS_LEFT {
		ant.Fatalf("[ant] expect '<' in line %v: %v", bracketsLeft.lineno, bracketsLeft.text)
	}
	listType := parser.lexer.takeToken()
	if parser.isDeclaredType(listType.text) == false {
		ant.Fatalf("[ant] undeclared type in line %v: %v", listType.lineno, listType.text)
	}
	bracketsRight := parser.lexer.takeToken() // '>'
	if bracketsRight.tokenType != TOKEN_ANGLE_BRACKETS_RIGHT {
		ant.Fatalf("[ant] expect '>' in line %v: %v", bracketsRight.lineno, bracketsRight.text)
	}
	return listType.text
}

func (parser *Parser) parseMapType() (string, string) {
	bracketsLeft := parser.lexer.takeToken() // '<'
	if bracketsLeft.tokenType != TOKEN_ANGLE_BRACKETS_LEFT {
		ant.Fatalf("[ant] expect '<' in line %v: %v", bracketsLeft.lineno, bracketsLeft.text)
	}
	mapKeyType := parser.lexer.takeToken()
	if mapKeyType.tokenType != TOKEN_KEYWORD_INT && mapKeyType.tokenType != TOKEN_KEYWORD_STRING &&
		mapKeyType.tokenType != TOKEN_KEYWORD_LONG {
		ant.Fatalf("[ant] undeclared type in line %v: %v", mapKeyType.lineno, mapKeyType.text)
	}
	comma := parser.lexer.takeToken() // ','
	if comma.tokenType != TOKEN_COMMA {
		ant.Fatalf("[ant] expect ',' for map define in line %v: %v", comma.lineno, comma.text)
	}
	mapValueType := parser.lexer.takeToken()
	if parser.isDeclaredType(mapValueType.text) == false {
		ant.Fatalf("[ant] undeclared type in line %v: %v", mapValueType.lineno, mapValueType.text)
	}
	bracketsRight := parser.lexer.takeToken() // '>'
	if bracketsRight.tokenType != TOKEN_ANGLE_BRACKETS_RIGHT {
		ant.Fatalf("[ant] expect '>' in line %v: %v", bracketsRight.lineno, bracketsRight.text)
	}
	return mapKeyType.text, mapValueType.text
}

func (parser *Parser) isDeclaredType(typeName string) bool {
	if typeName == "string" || typeName == "int32" || typeName == "int64" || typeName == "float32" || typeName == "float64" {
		return true
	}
	for _, v := range parser.types {
		if v.typeName == typeName {
			return true
		}
	}
	return false
}

func (parser *Parser) isSystemType(typeName string) bool {
	if typeName == "string" || typeName == "int32" || typeName == "int64" || typeName == "float32" || typeName == "float64" {
		return true
	}
	return false
}

func (parser *Parser) printTypes() {
	fmt.Println("\n------------ type define begin ------------")
	for i := 0; i < len(parser.types); i++ {
		currType := parser.types[i]
		fmt.Println("type: ", currType.typeName)
		for j := 0; j < len(currType.members); j++ {
			v := currType.members[j]
			fmt.Println("member: ", v.variableName, v.variableType, v.other1, v.other2)
		}
		if i != len(parser.types)-1 {
			fmt.Println()
		}
	}
	fmt.Println("------------------ end --------------------\n")
}

func (parser *Parser) printApis() {
	fmt.Println("\n------------ api define begin -------------")
	for i := 0; i < len(parser.apis); i++ {
		currApi := parser.apis[i]
		fmt.Print(currApi.access, " ", currApi.group, " ", currApi.name, " (")
		for j := 0; j < len(currApi.params); j++ {
			param := currApi.params[j]
			if j > 0 {
				fmt.Print(", ")
			}
			if param.other2 != "" {
				fmt.Print(param.variableName, " ", param.variableType, " ", param.other1, " ", param.other2)
			} else if param.other1 != "" {
				fmt.Print(param.variableName, " ", param.variableType, " ", param.other1)
			} else {
				fmt.Print(param.variableName, " ", param.variableType)
			}
		}
		fmt.Println(") ->", currApi.ret)
	}
	fmt.Println("------------------ end --------------------\n")
}

var types_tpl = `
package types

${type_define_list}

`

var handler_tpl = `
package api
import (
    "${app_path}/logic"
    "${app_path}/types"
    tp "github.com/henrylee2cn/teleport"
)
${api_define_group}
${api_define_function}

`

var router_tpl = `
package api
import (
    tp "github.com/henrylee2cn/teleport"
)
// Route registers handlers to router.
func Route(root string, router *tp.Router) {
    // root router group
    rootGroup := router.SubRoute(root)
    // custom router
    customRoute(rootGroup.ToRouter())
    // automatically generated router
    ${register_router_list}
}
`
var sdk_rpc_tpl = `
package sdk
import (
	"github.com/henrylee2cn/ant"
	"${app_path}/types"
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket"
)
var client *ant.Client
// Init init client with config and linker.
func Init(cliConfig ant.CliConfig, linker ant.Linker) {
	client = ant.NewClient(
		cliConfig,
		linker,
	)
}
// InitWithClient init client with current client.
func InitWithClient(cli *ant.Client) {
	client = cli
}
${rpc_call_define}
`

var sdk_rpc_test_tpl = `
package sdk
import (
	"testing"
	"github.com/henrylee2cn/ant"
	"${app_path}/types"
)

// TestSdk test SDK.
func TestSdk(t *testing.T) {
	Init(
		ant.CliConfig{
			Failover:        3,
			HeartbeatSecond: 4,
		},
		ant.NewStaticLinker(":9090"),
	)
	${rpc_call_test_define}
}
`

var logic_tpl = `
package logic
import (
	tp "github.com/henrylee2cn/teleport"
	"${app_path}/types"
)
${logic_api_define}
`

type CodeGen struct {
	parser *Parser
}

func (codeGen *CodeGen) init(parser *Parser) {
	codeGen.parser = parser
}

func mustMkdirAll(dir string) {
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		ant.Fatalf("[ant] %v", err)
	}
}

func (codeGen *CodeGen) genForGolang() {
	// make all directory
	mustMkdirAll("./types")
	mustMkdirAll("./api")
	mustMkdirAll("./logic")
	mustMkdirAll("./sdk")

	parser := codeGen.parser

	// 1. gen types/types.gen.go
	typeDefines := "\n"
	for i := 0; i < len(parser.types); i++ {
		currType := parser.types[i]
		typeHeader := fmt.Sprintf("type %s struct {\n", currType.typeName)
		typeMember := ""
		for j := 0; j < len(currType.members); j++ {
			v := currType.members[j]
			if v.variableType == "list" {
				if parser.isSystemType(v.other1) {
					typeMember += fmt.Sprintf("    %s []%s %s\n", v.variableName, v.other1, v.tag)
				} else {
					typeMember += fmt.Sprintf("    %s []*%s %s\n", v.variableName, v.other1, v.tag)
				}
			} else if v.variableType == "map" {
				if parser.isSystemType(v.other2) {
					typeMember += fmt.Sprintf("    %s map[%s]%s %s\n", v.variableName, v.other1, v.other2, v.tag)
				} else {
					typeMember += fmt.Sprintf("    %s map[%s]*%s %s\n", v.variableName, v.other1, v.other2, v.tag)
				}
			} else {
				if parser.isSystemType(v.variableType) {
					typeMember += fmt.Sprintf("    %s %s %s\n", v.variableName, v.variableType, v.tag)
				} else {
					typeMember += fmt.Sprintf("    %s *%s %s\n", v.variableName, v.variableType, v.tag)
				}
			}
		}
		typeTail := fmt.Sprintln("}\n\n")
		typeDefines += typeHeader + typeMember + typeTail
	}
	fileContent := strings.Replace(types_tpl, "${type_define_list}", typeDefines, 1)
	codeGen.saveFile("./types/types.gen.go", &fileContent)

	groups := make(map[string]string)
	for i := 0; i < len(parser.apis); i++ {
		currApi := parser.apis[i]
		if _, ok := groups[currApi.group]; ok == false {
			groups[currApi.group] = currApi.access
		}
	}
	// 2. gen api/handlers.gen.go
	// 2.1 scan all api group
	groupDefines := ""
	for k, v := range groups {
		accessStr := ""
		if v == "pull" {
			accessStr = "tp.PullCtx"
		} else if v == "push" {
			accessStr = "tp.PushCtx"
		}
		groupDefines += fmt.Sprintf("type %s struct {\n    %s\n}\n", k, accessStr)
	}

	// 2.2 scan all api function
	apiDefines := ""
	logicApiDefines := ""
	for i := 0; i < len(parser.apis); i++ {
		currApi := parser.apis[i]
		apiHeaderStr := fmt.Sprintf("func (handlers *%s) %s(", currApi.group, currApi.name)
		logicApiHeaderStr := fmt.Sprintf("func %s(", currApi.name)
		apiParamsStr := ""
		apiParamsStr2 := ""
		for j := 0; j < len(currApi.params); j++ {
			if j >= 1 && j < len(currApi.params) {
				apiParamsStr += fmt.Sprintf(", ")
				apiParamsStr2 += fmt.Sprintf(", ")
			}
			param := currApi.params[j]
			if param.variableType == "list" {
				if parser.isSystemType(param.other1) {
					apiParamsStr += fmt.Sprintf("%s []%s", param.variableName, param.other1)
				} else {
					apiParamsStr += fmt.Sprintf("%s []*types.%s", param.variableName, param.other1)
				}

			} else if param.variableType == "map" {
				if parser.isSystemType(param.other2) {
					apiParamsStr += fmt.Sprintf("%s [%s]%s", param.variableName, param.other1, param.other2)
				} else {
					apiParamsStr += fmt.Sprintf("%s [%s]*types.%s", param.variableName, param.other1, param.other2)
				}
			} else {
				// here translate base type, eg: int32 int64 float32 float64 string
				if parser.isSystemType(param.variableType) {
					apiParamsStr += fmt.Sprintf("%s %s", param.variableName, param.variableType)
				} else {
					apiParamsStr += fmt.Sprintf("%s *types.%s", param.variableName, param.variableType)
				}
			}
			apiParamsStr2 += param.variableName
		}
		replyStr := ""
		if currApi.access == "pull" {
			if parser.isSystemType(currApi.ret) {
				replyStr += fmt.Sprintf(") (%s, *tp.Rerror)", currApi.ret)
			} else {
				replyStr += fmt.Sprintf(") (*types.%s, *tp.Rerror)", currApi.ret)
			}
		} else {
			replyStr += fmt.Sprintf(") (*tp.Rerror)")
		}

		bodyStr := " {\n"
		bodyStr += fmt.Sprintf("    return logic.%s(%s)", currApi.name, apiParamsStr2)
		bodyStr += "\n}\n"
		logicBodySr := " {\n"
		if currApi.access == "pull" {
			logicBodySr += "return nil, nil"
		} else {
			logicBodySr += "return nil"
		}
		logicBodySr += "\n}\n"
		apiDefines += apiHeaderStr + apiParamsStr + replyStr + bodyStr + "\n"
		logicApiDefines += logicApiHeaderStr + apiParamsStr + replyStr + logicBodySr + "\n"
	}
	fileContent = ""
	fileContent = strings.Replace(handler_tpl, "${api_define_group}", groupDefines, 1)
	fileContent = strings.Replace(fileContent, "${api_define_function}", apiDefines, 1)
	codeGen.saveFile("./api/handlers.gen.go", &fileContent)
	fileContent = ""
	fileContent = strings.Replace(logic_tpl, "${logic_api_define}", logicApiDefines, 1)
	codeGen.saveFile("./logic/_logic.gen.go", &fileContent)

	// 3. gen api/router.gen.go
	routerRegisters := ""
	for k, v := range groups {
		if v == "pull" {
			routerRegisters += fmt.Sprintf("rootGroup.RoutePull(new(%s))\n", k)
		} else if v == "push" {
			routerRegisters += fmt.Sprintf("rootGroup.RoutePush(new(%s))\n", k)
		}
	}
	fileContent = router_tpl
	fileContent = strings.Replace(fileContent, "${register_router_list}", routerRegisters, 1)
	codeGen.saveFile("./api/router.gen.go", &fileContent)

	// 4. gen sdk/rpc.gen.go
	sdkRpcDefines := ""
	for i := 0; i < len(parser.apis); i++ {
		currApi := parser.apis[i]
		apiHeaderStr := fmt.Sprintf("func %s(", currApi.name)
		apiParamsStr := ""
		apiParamsStr2 := ""
		for j := 0; j < len(currApi.params); j++ {
			if j >= 1 && j < len(currApi.params) {
				apiParamsStr += fmt.Sprintf(", ")
				apiParamsStr2 += fmt.Sprintf(", ")
			}
			param := currApi.params[j]
			if param.variableType == "list" {
				if parser.isSystemType(param.other1) {
					apiParamsStr += fmt.Sprintf("%s []%s", param.variableName, param.other1)
				} else {
					apiParamsStr += fmt.Sprintf("%s []*types.%s", param.variableName, param.other1)
				}

			} else if param.variableType == "map" {
				if parser.isSystemType(param.other2) {
					apiParamsStr += fmt.Sprintf("%s [%s]%s", param.variableName, param.other1, param.other2)
				} else {
					apiParamsStr += fmt.Sprintf("%s [%s]*types.%s", param.variableName, param.other1, param.other2)
				}
			} else {
				// here translate base type, eg: int32 int64 float32 float64 string
				if parser.isSystemType(param.variableType) {
					apiParamsStr += fmt.Sprintf("%s %s", param.variableName, param.variableType)
				} else {
					apiParamsStr += fmt.Sprintf("%s *types.%s", param.variableName, param.variableType)
				}
			}
			apiParamsStr2 += param.variableName
		}
		apiParamsStr += ", setting ...socket.PacketSetting"
		replyStr := ""
		if currApi.access == "pull" {
			if parser.isSystemType(currApi.ret) {
				replyStr += fmt.Sprintf(") (%s, *tp.Rerror)", currApi.ret)
			} else {
				replyStr += fmt.Sprintf(") (*%s, *tp.Rerror)", currApi.ret)
			}
		} else {
			replyStr += fmt.Sprintf(") (*tp.Rerror)")
		}

		bodyStr := " {\n"
		if currApi.access == "pull" {
			bodyStr += fmt.Sprintf("    reply := new(types.%s)\n", currApi.ret)
			bodyStr += fmt.Sprintf("    rerr := client.Pull(\"/root/%s/%s\", %s, reply, setting...).Rerror()\n", goutil.SnakeString(currApi.group), goutil.SnakeString(currApi.name), currApi.params[0].variableName)
			bodyStr += "    return reply, rerr\n"
		} else if currApi.access == "push" {
			bodyStr += fmt.Sprintf("    rerr := client.Push(\"/root/%s/%s\", %s, setting...).Rerror()\n", goutil.SnakeString(currApi.group), goutil.SnakeString(currApi.name), currApi.params[0].variableName)
			bodyStr += "    return rerr\n"
		}

		bodyStr += "}\n"
		sdkRpcDefines += apiHeaderStr + apiParamsStr + replyStr + bodyStr + "\n"
	}
	fileContent = ""
	fileContent = strings.Replace(sdk_rpc_tpl, "${rpc_call_define}", sdkRpcDefines, 1)
	codeGen.saveFile("./sdk/rpc.gen.go", &fileContent)

	// 5. gen sdk/rpc.gen_test.go
	sdkRpcTestDefines := ""
	for i := 0; i < len(parser.apis); i++ {
		currApi := parser.apis[i]

		// support auto fill params
		//apiParamType := parser.getCustomType(currApi.params[0].variableType)
		//for j := 0; j < len(apiParamType.members); j++ {
		//	v := currType.members[j]
		//}
		colon := ":"
		if i > 0 {
			colon = ""
		}
		if parser.isDeclaredType(currApi.params[0].variableType) {
			callExp := fmt.Sprintf("    reply, rerr %s= %s(&types.%s{})\n", colon, currApi.name, currApi.params[0].variableType)
			callExp += "    if rerr != nil {\n"
			callExp += "    t.Logf(\"rerr: %v\", rerr)\n"
			callExp += "    } else {\n"
			callExp += "    }\n"
			sdkRpcTestDefines += callExp + "\n"
		} else {
			ant.Fatalf("[ant] not declared type  %v", currApi.params[0].variableType)
		}
	}
	fileContent = ""
	fileContent = strings.Replace(sdk_rpc_test_tpl, "${rpc_call_test_define}", sdkRpcTestDefines, 1)
	codeGen.saveFile("./sdk/rpc.gen_test.go", &fileContent)
}

func (codeGen *CodeGen) saveFile(fileName string, txt *string) {
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		ant.Fatalf("[ant] Create files error: %v", err)
	}
	defer f.Close()
	fileContent := strings.Replace(*txt, "${app_path}", info.ProjPath(), -1)
	f.WriteString("// generate by ant command\n\n")
	f.WriteString(fileContent)
}
