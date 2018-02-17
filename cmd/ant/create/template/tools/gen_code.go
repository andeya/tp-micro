package main

import (
	"fmt"
	//"github.com/urfave/cli"
	"bufio"
	"log"
	"os"
	"regexp"
	"strings"
)

const (
	TOKEN_MIN                  = iota
	TOKEN_ASSIGN               // =
	TOKEN_KEYWORD_TYPE         // type
	TOKEN_KEYWORD_API          // api
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
	TOKEN_SYMBOL
	TOKEN_MAX
)

var token_rules = map[int]string{
	//TOKEN_ASSIGN : 					"=",
	TOKEN_COMMENT:              "\\s*//\\s*",
	TOKEN_KEYWORD_TYPE:         "^\\s*type\\s+",
	TOKEN_KEYWORD_API:          "^\\s*api\\s+",
	TOKEN_COLON:                "\\s*:\\s*",
	TOKEN_COMMA:                "\\s*\\,\\s*",
	TOKEN_RETURN:               "\\s*->\\s*",
	TOKEN_BRACE_LEFT:           "\\s*\\{\\s*",
	TOKEN_BRACE_RIGHT:          "\\s*\\}\\s*",
	TOKEN_BRACKETS_LEFT:        "\\s*\\(\\s*",
	TOKEN_BRACKETS_RIGHT:       "\\s*\\)\\s*",
	TOKEN_ANGLE_BRACKETS_LEFT:  "\\s*<\\s*",
	TOKEN_ANGLE_BRACKETS_RIGHT: "\\s*>\\s*",
	TOKEN_KEYWORD_STRING:       "\\s*string\\s*",
	TOKEN_KEYWORD_INT:          "\\s*int\\s*",
	TOKEN_KEYWORD_LONG:         "\\s*long\\s*",
	TOKEN_KEYWORD_FLOAT:        "\\s*float\\s*",
	TOKEN_KEYWORD_DOUBLE:       "\\s*double\\s*",
	TOKEN_KEYWORD_LIST:         "\\s*list\\s*",
	TOKEN_KEYWORD_MAP:          "\\s*map\\s*",
	TOKEN_SYMBOL:               "\\s*[\\w]+\\s*",
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

func (lexer *Lexer) init(filePath string) {
	lexer.currTokenIdx = 0
	lexer.rules = map[int]*regexp.Regexp{}
	for k, v := range token_rules {
		reg := regexp.MustCompile(v)
		lexer.rules[k] = reg
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lexer.lines = append(lexer.lines, line)
	}
	for k, v := range lexer.lines {
		lexer.parseLine(k+1, v)
	}

	//for _, v := range lexer.tokens {
	//	fmt.Println("line ", v.lineno, ": ", v.text)
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
			log.Panic("There is some error in config file with line ", lineno, ": ", line)
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
	<api_define> ::= api { <api_declare> }
	<identifier> ::= "\w+"
	<type> ::= int | long | float | double | string | list<<type>> | map<<type>, <type>> | <identifier>
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
}

type CustomType struct {
	typeName string
	members  []*Variable
}

type CustomAPI struct {
	name   string
	params []*Variable
	ret    string
}

func (parser *Parser) init(lexer *Lexer) {
	parser.lexer = lexer
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
		log.Panic("invaid identifier in line ", identifier.lineno, ": ", identifier.text)
	}

	braceLeft := parser.lexer.takeToken()
	if braceLeft.tokenType != TOKEN_BRACE_LEFT {
		log.Panic("expect '{' in line ", braceLeft.lineno, ": ", braceLeft.text)
	}
	members := []*Variable{}
	parser.variable_declare(&members)
	braceRight := parser.lexer.takeToken()
	if braceRight.tokenType != TOKEN_BRACE_RIGHT {
		log.Panic("expect '}' in line ", braceRight.lineno, ": ", braceRight.text)
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
		log.Panic("invaid identifier in line ", identifier.lineno, ": ", identifier.text)
	}
	colon := parser.lexer.takeToken()
	if colon.tokenType != TOKEN_COLON {
		log.Panic("expect ':' in line ", colon.lineno, ": ", colon.text)
	}
	type_identifier := parser.lexer.takeToken()
	if type_identifier.tokenType == TOKEN_KEYWORD_LIST {
		listTypeName := parser.parseListType()
		*members = append(*members, &Variable{
			variableName: identifier.text,
			variableType: type_identifier.text,
			other1:       listTypeName,
			other2:       "",
		})
	} else if type_identifier.tokenType == TOKEN_KEYWORD_MAP {
		mapKeyTypeName, mapValueTypeName := parser.parseMapType()
		*members = append(*members, &Variable{
			variableName: identifier.text,
			variableType: type_identifier.text,
			other1:       mapKeyTypeName,
			other2:       mapValueTypeName,
		})
	} else {
		if parser.isDeclaredType(type_identifier.text) == false {
			log.Panic("undeclared type in line ", type_identifier.lineno, ": ", type_identifier.text)
		}
		*members = append(*members, &Variable{
			variableName: identifier.text,
			variableType: type_identifier.text,
			other1:       "",
			other2:       "",
		})
	}

	//fmt.Println("member: ", identifier.text, ":", type_identifier.text)

	// recursion for processing variable declare
	parser.variable_declare(members)
}

func (parser *Parser) api_define() {

	braceLeft := parser.lexer.takeToken()
	if braceLeft.tokenType != TOKEN_BRACE_LEFT {
		log.Panic("expect '{' in line ", braceLeft.lineno, ": ", braceLeft.text)
	}
	parser.api_declare()
	braceRight := parser.lexer.takeToken()
	if braceRight.tokenType != TOKEN_BRACE_RIGHT {
		log.Panic("expect '}' in line ", braceRight.lineno, ": ", braceRight.text)
	}

}

func (parser *Parser) api_declare() {
	if tokenType := parser.lexer.nextTokenType(); tokenType == TOKEN_BRACE_RIGHT {
		return
	}

	// 1. api function name
	function := parser.lexer.takeToken()
	if function.tokenType != TOKEN_SYMBOL {
		log.Panic("invaid api name dentifier in line ", function.lineno, ": ", function.text)
	}

	// 2. function '('
	bracketsLeft := parser.lexer.takeToken()
	if bracketsLeft.tokenType != TOKEN_BRACKETS_LEFT {
		log.Panic("expect '(' in line ", bracketsLeft.lineno, ": ", bracketsLeft.text)
	}

	// 3. function parameter
	params := []*Variable{}
	parser.parameter_declare(&params)

	// 4. function ')'
	bracketsRight := parser.lexer.takeToken()
	if bracketsRight.tokenType != TOKEN_BRACKETS_RIGHT {
		log.Panic("expect ')' in line ", bracketsRight.lineno, ": ", bracketsRight.text)
	}

	// 5. function -> return type
	retTagToken := parser.lexer.takeToken()
	if retTagToken.tokenType != TOKEN_RETURN {
		log.Panic("expect function return tag '->' in line ", retTagToken.lineno, ": ", retTagToken.text)
	}
	retToken := parser.lexer.takeToken()
	if parser.isDeclaredType(retToken.text) == false {
		log.Panic("undefined ret type in line ", retToken.lineno, ": ", retToken.text)
	}
	api := &CustomAPI{name: function.text, params: params, ret: retToken.text}
	parser.apis = append(parser.apis, api)

	// recursion for processing other api declare
	parser.api_declare()
}

func (parser *Parser) parameter_declare(params *[]*Variable) {
	paramName := parser.lexer.takeToken()
	if paramName.tokenType != TOKEN_SYMBOL {
		log.Panic("invaid parameter name dentifier in line ", paramName.lineno, ": ", paramName.text)
	}
	colon := parser.lexer.takeToken()
	if colon.tokenType != TOKEN_COLON {
		log.Panic("expect ':' in line ", colon.lineno, ": ", colon.text)
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
			log.Panic("undeclared param type in line ", paramType.lineno, ": ", paramType.text)
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
			log.Panic("expect ',' in line ", paramType.lineno, ": ", paramType.text)
		}
		parser.parameter_declare(params)
	}
}

func (parser *Parser) parseListType() string {
	bracketsLeft := parser.lexer.takeToken() // '<'
	if bracketsLeft.tokenType != TOKEN_ANGLE_BRACKETS_LEFT {
		log.Panic("expect '<' in line ", bracketsLeft.lineno, ": ", bracketsLeft.text)
	}
	listType := parser.lexer.takeToken()
	if parser.isDeclaredType(listType.text) == false {
		log.Panic("undeclared type in line ", listType.lineno, ": ", listType.text)
	}
	bracketsRight := parser.lexer.takeToken() // '>'
	if bracketsRight.tokenType != TOKEN_ANGLE_BRACKETS_RIGHT {
		log.Panic("expect '>' in line ", bracketsRight.lineno, ": ", bracketsRight.text)
	}
	return listType.text
}

func (parser *Parser) parseMapType() (string, string) {
	bracketsLeft := parser.lexer.takeToken() // '<'
	if bracketsLeft.tokenType != TOKEN_ANGLE_BRACKETS_LEFT {
		log.Panic("expect '<' in line ", bracketsLeft.lineno, ": ", bracketsLeft.text)
	}
	mapKeyType := parser.lexer.takeToken()
	if mapKeyType.tokenType != TOKEN_KEYWORD_INT && mapKeyType.tokenType != TOKEN_KEYWORD_STRING &&
		mapKeyType.tokenType != TOKEN_KEYWORD_LONG {
		log.Panic("undeclared type in line ", mapKeyType.lineno, ": ", mapKeyType.text)
	}
	comma := parser.lexer.takeToken() // ','
	if comma.tokenType != TOKEN_COMMA {
		log.Panic("expect ',' for map define in line ", comma.lineno, ": ", comma.text)
	}
	mapValueType := parser.lexer.takeToken()
	if parser.isDeclaredType(mapValueType.text) == false {
		log.Panic("undeclared type in line ", mapValueType.lineno, ": ", mapValueType.text)
	}
	bracketsRight := parser.lexer.takeToken() // '>'
	if bracketsRight.tokenType != TOKEN_ANGLE_BRACKETS_RIGHT {
		log.Panic("expect '>' in line ", bracketsRight.lineno, ": ", bracketsRight.text)
	}
	return mapKeyType.text, mapValueType.text
}

func (parser *Parser) isDeclaredType(typeName string) bool {
	if typeName == "string" || typeName == "int" || typeName == "long" || typeName == "float" || typeName == "double" {
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
	if typeName == "string" || typeName == "int" || typeName == "long" || typeName == "float" || typeName == "double" {
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
		fmt.Print(currApi.name, " (")
		for j := 0; j < len(currApi.params); j++ {
			param := currApi.params[j]
			if j > 0 {
				fmt.Print(", ")
			}
			fmt.Print(param.variableName, param.variableType, param.other1, param.other2)
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
    "logic"
    "types"
    tp "github.com/henrylee2cn/teleport"
)

type ApiHandlers struct {
    tp.PullCtx
}

${api_define_list}

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
    rootGroup.RoutePull(new(ApiHandlers))
}
`

var sdk_rpc_tpl = ``

var sdk_rpc_test_tpl = ``


type CodeGen struct {
	parser *Parser
}

func (codeGen *CodeGen) init(parser *Parser) {
	codeGen.parser = parser
}

func (codeGen *CodeGen) genForGolang() {
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
					typeMember += fmt.Sprintf("    %s []%s\n", v.variableName, v.other1)
				} else {
					typeMember += fmt.Sprintf("    %s []*%s\n", v.variableName, v.other1)
				}
			} else if v.variableType == "map" {
				if parser.isSystemType(v.other2) {
					typeMember += fmt.Sprintf("    %s [%s]%s\n", v.variableName, v.other1, v.other2)
				} else {
					typeMember += fmt.Sprintf("    %s [%s]*%s\n", v.variableName, v.other1, v.other2)
				}
			} else {
				if parser.isSystemType(v.variableType) {
					typeMember += fmt.Sprintf("    %s %s\n", v.variableName, v.variableType)
				} else {
					typeMember += fmt.Sprintf("    %s *%s\n", v.variableName, v.variableType)
				}
			}
		}
		typeTail := fmt.Sprintln("}\n\n")
		typeDefines += typeHeader + typeMember + typeTail
	}
	fileContent := strings.Replace(types_tpl, "${type_define_list}", typeDefines, 1)
	codeGen.saveFile("../types/types.gen.go", &fileContent)

	// 2. gen api/handlers.gen.go
	apiDefines := ""
	for i := 0; i < len(parser.apis); i++ {
		currApi := parser.apis[i]
		apiHeaderStr := fmt.Sprintf("func (handlers *ApiHandlers) %s(", currApi.name)
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
					apiParamsStr += fmt.Sprintf("%s []*%s", param.variableName, param.other1)
				}

			} else if param.variableType == "map" {
				if parser.isSystemType(param.other2) {
					apiParamsStr += fmt.Sprintf("%s [%s]%s", param.variableName, param.other1, param.other2)
				} else {
					apiParamsStr += fmt.Sprintf("%s [%s]*%s", param.variableName, param.other1, param.other2)
				}
			} else {
				if parser.isSystemType(param.variableType) {
					apiParamsStr += fmt.Sprintf("%s %s", param.variableName, param.variableType)
				} else {
					apiParamsStr += fmt.Sprintf("%s *%s", param.variableName, param.variableType)
				}
			}
			apiParamsStr2 += param.variableName
		}
		replyStr := ""
		if parser.isSystemType(currApi.ret) {
			replyStr += fmt.Sprintf(") (%s, *tp.Rerror)", currApi.ret)
		} else {
			replyStr += fmt.Sprintf(") (*%s, *tp.Rerror)", currApi.ret)
		}
		bodyStr := " {\n"
		bodyStr += fmt.Sprintf("    return logic.%s(%s)", currApi.name, apiParamsStr2)
		bodyStr += "\n}\n"
		apiDefines += apiHeaderStr + apiParamsStr + replyStr + bodyStr + "\n"
	}
	fileContent = strings.Replace(handler_tpl, "${api_define_list}", apiDefines, 1)
	codeGen.saveFile("../api/handlers.gen.go", &fileContent)

	// 3. gen api/router.gen.go
	fileContent = router_tpl
	codeGen.saveFile("../api/router.gen.go", &fileContent)
}

func (codeGen *CodeGen) saveFile(fileName string, txt *string) {
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		log.Panic(err)
	}
	defer  f.Close()
	f.WriteString(*txt)
}

func main() {
	lexer := Lexer{}
	lexer.init("../api.proto")

	parser := Parser{}
	parser.init(&lexer)
	parser.parse()

	codeGen := CodeGen{}
	codeGen.init(&parser)
	codeGen.genForGolang()
	fmt.Println("winer winer chicken dinner!")
}
