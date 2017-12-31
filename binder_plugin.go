// Copyright 2017 HenryLee. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ants

import (
	"errors"
	"fmt"
	"math"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
)

// StructArgsBinder a plugin that binds and validates structure type parameters.
type StructArgsBinder struct {
	binders        map[string]*Params
	bindErrCode    int32
	bindErrMessage string
}

var (
	_ tp.PostRegPlugin          = new(StructArgsBinder)
	_ tp.PostReadPullBodyPlugin = new(StructArgsBinder)
)

// NewStructArgsBinder creates a plugin that binds and validates structure type parameters.
func NewStructArgsBinder(bindErrCode int32, bindErrMessage string) *StructArgsBinder {
	return &StructArgsBinder{
		binders:        make(map[string]*Params),
		bindErrCode:    bindErrCode,
		bindErrMessage: bindErrMessage,
	}
}

// Name returns the plugin name.
func (*StructArgsBinder) Name() string {
	return "StructArgsBinder"
}

// PostReg preprocessing struct handler.
func (s *StructArgsBinder) PostReg(h *tp.Handler) *tp.Rerror {
	if h.ArgElemType().Kind() != reflect.Struct {
		return nil
	}
	params := newParams(h.Name())
	err := params.addFields([]int{}, h.ArgElemType(), h.NewArgValue())
	if err != nil {
		tp.Fatalf("%v", err)
	}
	s.binders[h.Name()] = params
	return nil
}

// PostReadPullBody binds and validates the registered struct handler.
func (s *StructArgsBinder) PostReadPullBody(ctx tp.ReadCtx) *tp.Rerror {
	params, ok := s.binders[ctx.Path()]
	if !ok {
		return nil
	}
	bodyValue := reflect.ValueOf(ctx.Input().Body())
	err := params.bindAndValidate(bodyValue, ctx.Query())
	if err != nil {
		return tp.NewRerror(s.bindErrCode, s.bindErrMessage, err.Error())
	}
	return nil
}

// Params struct handler information for binding and validation
type Params struct {
	handlerName string
	params      []*Param
}

// struct binder parameters'tag
const (
	TAG_PARAM        = "param"   // request param tag name
	TAG_IGNORE_PARAM = "-"       // ignore request param tag value
	KEY_QUERY        = "query"   // query param(optional), value means parameter(optional)
	KEY_DESC         = "desc"    // request param description
	KEY_LEN          = "len"     // length range of param's value
	KEY_RANGE        = "range"   // numerical range of param's value
	KEY_NONZERO      = "nonzero" // param`s value can not be zero
	KEY_REGEXP       = "regexp"  // verify the value of the param with a regular expression(param value can not be null)
	KEY_ERR          = "err"     // the custom error for binding or validating
)

func newParams(handlerName string) *Params {
	return &Params{
		handlerName: handlerName,
		params:      make([]*Param, 0),
	}
}

func (p *Params) addFields(parentIndexPath []int, t reflect.Type, v reflect.Value) error {
	var err error
	var deep = len(parentIndexPath) + 1
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		indexPath := make([]int, deep)
		copy(indexPath, parentIndexPath)
		indexPath[deep-1] = i

		var field = t.Field(i)
		var value = v.Field(i)
		canSet := v.Field(i).CanSet()

		tag, ok := field.Tag.Lookup(TAG_PARAM)
		if !ok {
			if field.Type.Kind() == reflect.Struct {
				if err = p.addFields(indexPath, field.Type, value); err != nil {
					return err
				}
			}
			continue
		}

		if tag == TAG_IGNORE_PARAM {
			continue
		}

		if !canSet {
			return fmt.Errorf("%s.%s can not be a non-settable field", t.String(), field.Name)
		}

		if field.Type.Kind() == reflect.Ptr {
			return fmt.Errorf("%s.%s can not be a pointer field", t.String(), field.Name)
		}

		var parsedTags = ParseTags(tag)
		var paramTypeString = field.Type.String()
		var kind = field.Type.Kind()

		if _, ok := parsedTags[KEY_LEN]; ok {
			if kind != reflect.String && kind != reflect.Slice && kind != reflect.Map && kind != reflect.Array {
				return fmt.Errorf("%s.%s invalid `len` tag for field value", t.String(), field.Name)
			}
		}
		if _, ok := parsedTags[KEY_RANGE]; ok {
			switch paramTypeString {
			case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64":
			case "[]int", "[]int8", "[]int16", "[]int32", "[]int64", "[]uint", "[]uint8", "[]uint16", "[]uint32", "[]uint64", "[]float32", "[]float64":
			default:
				return fmt.Errorf("%s.%s invalid `range` tag for non-number field", t.String(), field.Name)
			}
		}
		if _, ok := parsedTags[KEY_REGEXP]; ok {
			if paramTypeString != "string" && paramTypeString != "[]string" {
				return fmt.Errorf("%s.%s invalid `regexp` tag for non-string field", t.String(), field.Name)
			}
		}

		fd := &Param{
			handlerName: p.handlerName,
			indexPath:   indexPath,
			tags:        parsedTags,
			rawTag:      field.Tag,
			rawValue:    value,
		}

		fd.err = fd.tags[KEY_ERR]

		fd.name, fd.isQuery = parsedTags[KEY_QUERY]
		if fd.name == "" {
			fd.name = goutil.SnakeString(field.Name)
		}

		if err = fd.makeVerifyFuncs(); err != nil {
			return fmt.Errorf("%s.%s invalid validation failed: %s", t.String(), field.Name, err.Error())
		}

		p.params = append(p.params, fd)
	}

	return nil
}

func (p *Params) fieldsForBinding(structElem reflect.Value) []reflect.Value {
	count := len(p.params)
	fields := make([]reflect.Value, count)
	for i := 0; i < count; i++ {
		value := structElem
		param := p.params[i]
		for _, index := range param.indexPath {
			value = value.Field(index)
		}
		fields[i] = value
	}
	return fields
}

func (p *Params) bindAndValidate(structValue reflect.Value, queryValues url.Values) (err error) {
	fields := p.fieldsForBinding(reflect.Indirect(structValue))
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("bindAndValidate: " + p.handlerName + " : " + fmt.Sprint(r))
		}
	}()

	for i, param := range p.params {
		value := fields[i]
		if param.isQuery {
			paramValues, ok := queryValues[param.name]
			if ok {
				if err = convertAssign(value, paramValues); err != nil {
					return param.myError(err.Error())
				}
			}
		}
		if err = param.validate(value); err != nil {
			return err
		}
	}
	return
}

// ParseTags returns the key-value in the tag string.
// If the tag does not have the conventional format,
// the value returned by ParseTags is unspecified.
func ParseTags(tag string) map[string]string {
	var values = map[string]string{}

	for tag != "" {
		// Skip leading space.
		i := 0
		for i < len(tag) && tag[i] != '<' {
			i++
		}
		if i >= len(tag) || tag[i] != '<' {
			break
		}
		i++

		// Skip the left Spaces
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		if i >= len(tag) {
			break
		}

		tag = tag[i:]
		if tag == "" {
			break
		}

		var name, value string
		var hadName bool
		i = 0
	PAIR:
		for i < len(tag) {
			switch tag[i] {
			case ':':
				if hadName {
					i++
					continue
				}
				name = strings.TrimRight(tag[:i], " ")
				tag = strings.TrimLeft(tag[i+1:], " ")
				hadName = true
				i = 0
			case '\\':
				i++
				// Fix the escape character of `\\<` or `\\>`
				if tag[i] == '<' || tag[i] == '>' {
					tag = tag[:i-1] + tag[i:]
				} else {
					i++
				}
			case '>':
				if !hadName {
					name = strings.TrimRight(tag[:i], " ")
				} else {
					value = strings.TrimRight(tag[:i], " ")
				}
				values[name] = value
				break PAIR
			default:
				i++
			}
		}
		if i >= len(tag) {
			break
		}
		tag = tag[i+1:]
	}
	return values
}

// Param use the struct field to define a request parameter model
type Param struct {
	handlerName string // handler name
	name        string // param name
	indexPath   []int
	isQuery     bool              // is query param or not
	tags        map[string]string // struct tags for this param
	verifyFuncs []func(reflect.Value) error
	rawTag      reflect.StructTag // the raw tag
	rawValue    reflect.Value     // the raw tag value
	err         string            // the custom error for binding or validating
}

const (
	stringTypeString = "string"
	bytesTypeString  = "[]byte"
	bytes2TypeString = "[]uint8"
)

// Raw gets the param's original value
func (param *Param) Raw() interface{} {
	return param.rawValue.Interface()
}

// Name gets parameter field name
func (param *Param) Name() string {
	return param.name
}

// Description gets the description value for the param
func (param *Param) Description() string {
	return param.tags[KEY_DESC]
}

// validate tests if the param conforms to it's validation constraints specified
// int the KEY_REGEXP struct tag
func (param *Param) validate(value reflect.Value) (err error) {
	defer func() {
		p := recover()
		if p != nil {
			err = param.myError(fmt.Sprint(p))
		} else if err != nil {
			err = param.myError(err.Error())
		}
	}()
	for _, fn := range param.verifyFuncs {
		if err = fn(value); err != nil {
			return err
		}
	}
	return nil
}

func (param *Param) makeVerifyFuncs() (err error) {
	defer func() {
		p := recover()
		if p != nil {
			err = fmt.Errorf("%v", p)
		}
	}()
	// length
	if tuple, ok := param.tags[KEY_LEN]; ok {
		if fn, err := validateLen(tuple); err == nil {
			param.verifyFuncs = append(param.verifyFuncs, fn)
		} else {
			return err
		}
	}
	// range
	if tuple, ok := param.tags[KEY_RANGE]; ok {
		if fn, err := validateRange(tuple); err == nil {
			param.verifyFuncs = append(param.verifyFuncs, fn)
		} else {
			return err
		}
	}
	// nonzero
	if _, ok := param.tags[KEY_NONZERO]; ok {
		if fn, err := validateNonZero(); err == nil {
			param.verifyFuncs = append(param.verifyFuncs, fn)
		} else {
			return err
		}
	}
	// regexp
	if reg, ok := param.tags[KEY_REGEXP]; ok {
		var isStrings = param.rawValue.Kind() == reflect.Slice
		if fn, err := validateRegexp(isStrings, reg); err == nil {
			param.verifyFuncs = append(param.verifyFuncs, fn)
		} else {
			return err
		}
	}
	return
}

func parseTuple(tuple string) (string, string) {
	c := strings.Split(tuple, ":")
	var a, b string
	switch len(c) {
	case 1:
		a = c[0]
		if len(a) > 0 {
			return a, a
		}
	case 2:
		a = c[0]
		b = c[1]
		if len(a) > 0 || len(b) > 0 {
			return a, b
		}
	}
	panic("invalid validation tuple")
}

func validateNonZero() (func(value reflect.Value) error, error) {
	return func(value reflect.Value) error {
		obj := value.Interface()
		if obj == reflect.Zero(value.Type()).Interface() {
			return errors.New("not set")
		}
		return nil
	}, nil
}

func validateLen(tuple string) (func(value reflect.Value) error, error) {
	var a, b = parseTuple(tuple)
	var min, max int
	var err error
	if len(a) > 0 {
		min, err = strconv.Atoi(a)
		if err != nil {
			return nil, err
		}
	}
	if len(b) > 0 {
		max, err = strconv.Atoi(b)
		if err != nil {
			return nil, err
		}
	}
	return func(value reflect.Value) error {
		length := value.Len()
		if len(a) > 0 {
			if length < min {
				return fmt.Errorf("shorter than %s: %v", a, value.Interface())
			}
		}
		if len(b) > 0 {
			if length > max {
				return fmt.Errorf("longer than %s: %v", b, value.Interface())
			}
		}
		return nil
	}, nil
}

const accuracy = 0.0000001

func validateRange(tuple string) (func(value reflect.Value) error, error) {
	var a, b = parseTuple(tuple)
	var min, max float64
	var err error
	if len(a) > 0 {
		min, err = strconv.ParseFloat(a, 64)
		if err != nil {
			return nil, err
		}
	}
	if len(b) > 0 {
		max, err = strconv.ParseFloat(b, 64)
		if err != nil {
			return nil, err
		}
	}
	return func(value reflect.Value) error {
		var f64 float64
		switch value.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			f64 = float64(value.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			f64 = float64(value.Uint())
		case reflect.Float32, reflect.Float64:
			f64 = value.Float()
		}
		if len(a) > 0 {
			if math.Min(f64, min) == f64 && math.Abs(f64-min) > accuracy {
				return fmt.Errorf("smaller than %s: %v", a, value.Interface())
			}
		}
		if len(b) > 0 {
			if math.Max(f64, max) == f64 && math.Abs(f64-max) > accuracy {
				return fmt.Errorf("bigger than %s: %v", b, value.Interface())
			}
		}
		return nil
	}, nil
}

func validateRegexp(isStrings bool, reg string) (func(value reflect.Value) error, error) {
	re, err := regexp.Compile(reg)
	if err != nil {
		return nil, err
	}
	if !isStrings {
		return func(value reflect.Value) error {
			s := value.String()
			if !re.MatchString(s) {
				return fmt.Errorf("not match %s: %s", reg, s)
			}
			return nil
		}, nil
	} else {
		return func(value reflect.Value) error {
			for _, s := range value.Interface().([]string) {
				if !re.MatchString(s) {
					return fmt.Errorf("not match %s: %s", reg, s)
				}
			}
			return nil
		}, nil
	}
}

// BindOrValidateErrorFunc creates an relational error.
type BindOrValidateErrorFunc func(handler, param, reason string) error

var bindOrValidateErrorFunc = func(handler, param, reason string) error {
	return fmt.Errorf(`{"handler": %q, "param": %q, "reason": %q}`, handler, param, reason)
}

// SetBindOrValidateErrorFunc
func SetBindOrValidateErrorFunc(fn BindOrValidateErrorFunc) {
	bindOrValidateErrorFunc = fn
}

func (param *Param) myError(reason string) error {
	if param.err != "" {
		reason = param.err
	}
	return bindOrValidateErrorFunc(param.handlerName, param.name, reason)
}

func convertAssign(dest reflect.Value, src []string) (err error) {
	if len(src) == 0 {
		return nil
	}

	dest = reflect.Indirect(dest)
	if !dest.CanSet() {
		return fmt.Errorf("%s can not be setted", dest.Type().Name())
	}

	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("%v", p)
		}
	}()

	switch dest.Interface().(type) {
	case string:
		dest.Set(reflect.ValueOf(src[0]))
		return nil

	case []string:
		dest.Set(reflect.ValueOf(src))
		return nil

	case []byte:
		dest.Set(reflect.ValueOf([]byte(src[0])))
		return nil

	case [][]byte:
		b := make([][]byte, 0, len(src))
		for _, s := range src {
			b = append(b, []byte(s))
		}
		dest.Set(reflect.ValueOf(b))
		return nil

	case bool:
		dest.Set(reflect.ValueOf(parseBool(src[0])))
		return nil

	case []bool:
		b := make([]bool, 0, len(src))
		for _, s := range src {
			b = append(b, parseBool(s))
		}
		dest.Set(reflect.ValueOf(b))
		return nil
	}

	switch dest.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i64, err := strconv.ParseInt(src[0], 10, dest.Type().Bits())
		if err != nil {
			err = strconvErr(err)
			return fmt.Errorf("converting type %T (%q) to a %s: %v", src, src[0], dest.Kind(), err)
		}
		dest.SetInt(i64)
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u64, err := strconv.ParseUint(src[0], 10, dest.Type().Bits())
		if err != nil {
			err = strconvErr(err)
			return fmt.Errorf("converting type %T (%q) to a %s: %v", src, src[0], dest.Kind(), err)
		}
		dest.SetUint(u64)
		return nil

	case reflect.Float32, reflect.Float64:
		f64, err := strconv.ParseFloat(src[0], dest.Type().Bits())
		if err != nil {
			err = strconvErr(err)
			return fmt.Errorf("converting type %T (%q) to a %s: %v", src, src[0], dest.Kind(), err)
		}
		dest.SetFloat(f64)
		return nil

	case reflect.Slice:
		member := dest.Type().Elem()
		switch member.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			for _, s := range src {
				i64, err := strconv.ParseInt(s, 10, member.Bits())
				if err != nil {
					err = strconvErr(err)
					return fmt.Errorf("converting type %T (%q) to a %s: %v", src, s, dest.Kind(), err)
				}
				dest.Set(reflect.Append(dest, reflect.ValueOf(i64).Convert(member)))
			}
			return nil

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			for _, s := range src {
				u64, err := strconv.ParseUint(s, 10, member.Bits())
				if err != nil {
					err = strconvErr(err)
					return fmt.Errorf("converting type %T (%q) to a %s: %v", src, s, dest.Kind(), err)
				}
				dest.Set(reflect.Append(dest, reflect.ValueOf(u64).Convert(member)))
			}
			return nil

		case reflect.Float32, reflect.Float64:
			for _, s := range src {
				f64, err := strconv.ParseFloat(s, member.Bits())
				if err != nil {
					err = strconvErr(err)
					return fmt.Errorf("converting type %T (%q) to a %s: %v", src, s, dest.Kind(), err)
				}
				dest.Set(reflect.Append(dest, reflect.ValueOf(f64).Convert(member)))
			}
			return nil
		}
	}

	return fmt.Errorf("unsupported storing type %T into type %s", src, dest.Kind())
}

func parseBool(val string) bool {
	switch strings.TrimSpace(strings.ToLower(val)) {
	case "true", "on", "1":
		return true
	}
	return false
}

func strconvErr(err error) error {
	if ne, ok := err.(*strconv.NumError); ok {
		return ne.Err
	}
	return err
}
