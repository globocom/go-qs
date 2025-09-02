package goqs

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

type QueryItem struct {
	Key   []string
	Value string
}

// Defaults for parse options
var defaults = ParseOptions{
	AllowDots:                false,
	AllowEmptyArrays:         false,
	AllowPrototypes:          false,
	ParseNumbers:             false,
	AllowSparse:              false,
	ArrayLimit:               20,
	Charset:                  "utf-8",
	CharsetSentinel:          false,
	Comma:                    false,
	DecodeDotInKeys:          false,
	Decoder:                  nil,
	Delimiter:                "&",
	Depth:                    5,
	Duplicates:               "combine",
	IgnoreQueryPrefix:        false,
	InterpretNumericEntities: false,
	ParameterLimit:           1000,
	ParseArrays:              true,
	PlainObjects:             false,
	StrictDepth:              false,
	StrictNullHandling:       false,
	AllowNilArrayValues:      false,
	ThrowOnLimitExceeded:     false,
}

// ParseOptions holds options for parsing
type ParseOptions struct {
	AllowDots                bool
	AllowEmptyArrays         bool
	AllowPrototypes          bool
	AllowSparse              bool
	ParseNumbers             bool
	ArrayLimit               int
	Charset                  string
	CharsetSentinel          bool
	Comma                    bool
	DecodeDotInKeys          bool
	Decoder                  DecoderFunc
	Delimiter                string
	Depth                    int
	Duplicates               string
	IgnoreQueryPrefix        bool
	InterpretNumericEntities bool
	ParameterLimit           int
	ParseArrays              bool
	PlainObjects             bool
	StrictDepth              bool
	StrictNullHandling       bool
	AllowNilArrayValues      bool
	ThrowOnLimitExceeded     bool
}

// DecodeFunc defines a function type for string decoding
type DecodeFunc func(string) string

// DecoderFunc defines a function type for string decoding with context
type DecoderFunc func(string, DecodeFunc, string, string) string

// defaultDecoder is the default implementation for the Decoder option
func defaultDecoder(s string, decodeFunc DecodeFunc, charset string, typ string) string {
	return decodeFunc(s)
}

func normalizeParseOptions(opts *ParseOptions) ParseOptions {
	if opts == nil {
		result := defaults
		if result.Decoder == nil {
			result.Decoder = defaultDecoder
		}
		return result
	}

	o := *opts
	if o.Charset == "" {
		o.Charset = defaults.Charset
	}
	if o.Duplicates == "" {
		o.Duplicates = defaults.Duplicates
	}
	if o.Delimiter == "" {
		o.Delimiter = defaults.Delimiter
	}
	if o.Decoder == nil {
		o.Decoder = defaultDecoder
	}
	return o
}

func parseArrayValue(val string, options ParseOptions, currentArrayLength int) interface{} {

	if val == "true" || val == "false" {
		return val
	}

	if val != "" && options.Comma && strings.Contains(val, ",") {
		return strings.Split(val, ",")
	}

	if options.ThrowOnLimitExceeded && currentArrayLength >= options.ArrayLimit {
		panic(fmt.Errorf("Array limit exceeded. Only %d element%s allowed in an array.", options.ArrayLimit, func() string {
			if options.ArrayLimit == 1 {
				return ""
			}
			return "s"
		}()))
	}

	return val
}

func PostProcessParsedObject(obj map[string]interface{}, options *ParseOptions) map[string]interface{} {
	result := processNestedStructures(obj, options)
	if m, ok := result.(map[string]interface{}); ok {
		return m
	}
	return obj
}

func needsToBeObject(arr []interface{}) bool {
	hasMapElements := false
	hasSpecificProperties := false

	for _, v := range arr {
		if _, ok := v.(map[string]interface{}); ok {
			hasMapElements = true

			// Verifica se o mapa tem propriedades como "contains", "some", "every", etc.
			if m, ok := v.(map[string]interface{}); ok {
				for k := range m {
					if k == "contains" || k == "some" || k == "every" || k == "none" || k == "length" {
						hasSpecificProperties = true
						break
					}
				}
			}
		}
	}

	return hasMapElements && hasSpecificProperties
}

func convertToQsArrayFormat(m map[string]interface{}) []interface{} {
	result := []interface{}{}

	for k, v := range m {
		if _, err := strconv.Atoi(k); err == nil {
			if b, ok := v.(bool); ok && b {
				result = append(result, k)
			}
		}
	}

	opMap := map[string]interface{}{}
	for k, v := range m {
		if k == "lt" || k == "gt" || k == "like" || k == "neq" || k == "in" || k == "between" {
			opMap[k] = v
		}
	}

	if len(opMap) > 0 {
		result = append(result, opMap)
	}

	return result
}

func EscapeQueryString(rawQuery string) string {
	rawQuery = strings.ReplaceAll(rawQuery, "\t", "")
	rawQuery = strings.ReplaceAll(rawQuery, "\n", "")

	parts := strings.Split(rawQuery, "&")

	for i, part := range parts {
		if idx := strings.Index(part, "="); idx != -1 {
			key := part[:idx]
			value := part[idx+1:]

			if unescapedValue, err := url.QueryUnescape(value); err == nil {
				value = unescapedValue
			}

			escapedValue := url.QueryEscape(value)
			key, _ = url.QueryUnescape(key)
			parts[i] = key + "=" + escapedValue
		}
	}
	return strings.Join(parts, "&")
}

func DecodeQuery(str string, opts *ParseOptions) (result map[string]interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = v
			case string:
				err = errors.New(v)
			default:
				err = fmt.Errorf("unexpected query decoder panic: %v", r)
			}
			result = nil
		}
	}()

	str = EscapeQueryString(str)
	options := normalizeParseOptions(opts)
	if str == "" {
		if options.PlainObjects {
			return map[string]interface{}{}, nil
		}
		return map[string]interface{}{}, nil
	}
	var urlValues [][]string
	if s, ok := interface{}(str).(string); ok {
		urlValues = parseValues(s, options)
	} else {
		return nil, errors.New("input must be a string")
	}
	obj := map[string]interface{}{}
	for _, pair := range urlValues {
		if len(pair) != 3 {
			continue
		}
		// realIndex := pair[0]
		key := pair[1]
		value := pair[2]
		newObj := parseKeys(key, value, options, true)
		merged := Merge(obj, newObj, options)
		if m, ok := merged.(map[string]interface{}); ok {
			obj = m
		} else {
			return nil, errors.New("failed to merge objects")
		}
	}
	// if options.AllowSparse {
	// 	processed := PostProcessParsedObject(obj)
	// 	return processed, nil
	// }

	compacted, ok := Compact(obj).(map[string]interface{})
	if !ok {
		return nil, errors.New("failed to compact object")
	}
	processed := PostProcessParsedObject(compacted, opts)
	return processed, nil
}

func parseKeys(givenKey string, val interface{}, options ParseOptions, valuesParsed bool) interface{} {
	if givenKey == "" {
		return nil
	}
	key := givenKey
	if options.AllowDots {
		re := regexp.MustCompile(`\.([^.[]+)`)
		key = re.ReplaceAllString(key, "[$1]")
	}
	brackets := regexp.MustCompile(`(\[[^[\]]*])`)
	child := regexp.MustCompile(`(\[[^[\]]*])`)
	segment := brackets.FindStringSubmatchIndex(key)
	var parent string
	if len(segment) >= 2 {
		parent = key[:segment[0]]
	} else {
		parent = key
	}
	keys := []string{}
	if parent != "" {
		if !options.PlainObjects && parent == "__proto__" && !options.AllowPrototypes {
			return nil
		}
		keys = append(keys, parent)
	}
	i := 0
	seg := child.FindStringSubmatchIndex(key)
	for options.Depth > 0 && seg != nil && len(seg) >= 2 && i < options.Depth {
		i++
		segmentStr := key[seg[0]:seg[1]]
		if !options.PlainObjects && len(segmentStr) > 2 && segmentStr[1:len(segmentStr)-1] == "__proto__" && !options.AllowPrototypes {
			return nil
		}
		keys = append(keys, segmentStr)
		key = key[seg[1]:]
		seg = child.FindStringSubmatchIndex(key)
	}
	if seg != nil && len(key) > 0 {
		if options.StrictDepth {
			panic(fmt.Errorf("Input depth exceeded depth option of %d and strictDepth is true", options.Depth))
		}
		keys = append(keys, "["+key+"]")
	}
	return parseObject(keys, val, options, valuesParsed)
}

func parseValues(str string, options ParseOptions) [][]string {
	result := [][]string{}

	cleanStr := str
	if options.IgnoreQueryPrefix {
		cleanStr = strings.TrimPrefix(cleanStr, "?")
	}
	cleanStr = strings.ReplaceAll(cleanStr, "%5B", "[")
	cleanStr = strings.ReplaceAll(cleanStr, "%5D", "]")

	limit := options.ParameterLimit
	if limit == 0 {
		limit = 1000
	}
	parts := strings.SplitN(cleanStr, options.Delimiter, limit+1)
	if options.ThrowOnLimitExceeded && len(parts) > limit {
		panic(fmt.Errorf("Parameter limit exceeded. Only %d parameter%s allowed.", limit, func() string {
			if limit == 1 {
				return ""
			}
			return "s"
		}()))
	}

	skipIndex := -1
	charset := options.Charset
	if options.CharsetSentinel {
		for i, part := range parts {
			if strings.HasPrefix(part, "utf8=") {
				charsetSentinel := ""
				isoSentinel := ""
				if part == charsetSentinel {
					charset = "utf-8"
				} else if part == isoSentinel {
					charset = "iso-8859-1"
				}
				skipIndex = i
				break
			}
		}
	}

	decoder := options.Decoder
	if decoder == nil {
		decoder = defaultDecoder
	}

	paramIndex := 0
	var pos int
	for i, part := range parts {
		if i == skipIndex {
			continue
		}

		bracketEqualsPos := strings.Index(part, "]=")
		if bracketEqualsPos == -1 {
			pos = strings.Index(part, "=")
		} else {
			pos = bracketEqualsPos + 1
		}

		var key string
		var val interface{}
		if pos == -1 {
			key = decoder(part, Decode, charset, "key")
			if options.StrictNullHandling {
				val = nil
			} else {
				val = ""
			}
		} else {
			key = decoder(part[:pos], Decode, charset, "key")
			if key != "" {
				rawVal := part[pos+1:]
				// For simplicity, we'll convert the value directly to string
				// since we're storing everything as strings in the result
				val = decoder(rawVal, Decode, charset, "value") //nolint:staticcheck

				if val != nil && options.InterpretNumericEntities && charset == "iso-8859-1" { //nolint:staticcheck
					if s, ok := val.(string); ok {
						val = interpretNumericEntities(s)
					}
				}
			}
		}

		if key != "" {
			// Convert the value to string for storage in our result slice
			valStr := ""
			if val != nil {
				valStr = AsString(val)
			}

			// Add the parameter to our result slice
			result = append(result, []string{
				strconv.Itoa(paramIndex),
				key,
				valStr,
			})

			paramIndex++
		}
	}

	return result
}

func interpretNumericEntities(str string) string {
	re := regexp.MustCompile(`&#(\d+);`)
	return re.ReplaceAllStringFunc(str, func(s string) string {
		m := re.FindStringSubmatch(s)
		if len(m) == 2 {
			n, _ := strconv.Atoi(m[1])
			return string(rune(n))
		}
		return s
	})
}

type RFCFormat string

const (
	RFC1738          RFCFormat = "RFC1738"
	RFC3986          RFCFormat = "RFC3986"
	DefaultRFCFormat           = RFC3986
)

var Formatters = map[RFCFormat]func(string) string{
	RFC1738: func(value string) string {
		// Replaces %20 por +
		return strings.ReplaceAll(value, "%20", "+")
	},
	RFC3986: func(value string) string {
		return value
	},
}

var hexTable [256]string

func init() {
	for i := 0; i < 256; i++ {
		hexTable[i] = "%" + strings.ToUpper(strconv.FormatInt(int64(i), 16))
		if i < 16 {
			hexTable[i] = "%0" + strings.ToUpper(strconv.FormatInt(int64(i), 16))
		}
	}
}

func CompactQueue(queue []queueItem) {
	for len(queue) > 1 {
		item := queue[len(queue)-1]
		queue = queue[:len(queue)-1]
		obj := reflect.ValueOf(item.obj).Elem().FieldByName(item.prop)
		if obj.Kind() == reflect.Slice {
			compacted := reflect.MakeSlice(obj.Type(), 0, obj.Len())
			for j := 0; j < obj.Len(); j++ {
				if !obj.Index(j).IsNil() {
					compacted = reflect.Append(compacted, obj.Index(j))
				}
			}
			obj.Set(compacted)
		}
	}
}

type queueItem struct {
	obj  interface{}
	prop string
}

func ArrayToObject(source []interface{}, plainObjects bool) map[int]interface{} {
	obj := make(map[int]interface{})
	for i, v := range source {
		if v != nil {
			// Convert booleans to strings
			if b, ok := v.(bool); ok {
				if b {
					obj[i] = "true"
				} else {
					obj[i] = "false"
				}
			} else {
				obj[i] = v
			}
		}
	}
	return obj
}
func Merge(target, source any, options ParseOptions) any {
	if source == nil {
		return target
	}

	sourceVal := reflect.ValueOf(source)
	if sourceVal.Kind() != reflect.Map && sourceVal.Kind() != reflect.Slice && sourceVal.Kind() != reflect.Array {
		switch t := target.(type) {
		case []any:
			return append(t, source)
		case map[string]any:
			if options.PlainObjects || options.AllowPrototypes {
				key := AsString(source)
				t[key] = true
			}
			return t
		default:
			return []any{target, source}
		}
	}

	if target == nil || (reflect.ValueOf(target).Kind() != reflect.Map &&
		reflect.ValueOf(target).Kind() != reflect.Slice &&
		reflect.ValueOf(target).Kind() != reflect.Array) {
		return append([]any{target}, source)
	}

	mergeTarget := target
	targetArr, targetIsArray := target.([]any)
	sourceArr, sourceIsArray := source.([]any)

	if targetIsArray && !sourceIsArray {
		mergeTarget = ArrayToObject(targetArr, options.PlainObjects)
	}

	if targetIsArray && sourceIsArray {
		for i, item := range sourceArr {
			if item == nil {
				continue
			}

			if i < len(targetArr) {
				targetItem := targetArr[i]
				if targetItem != nil && reflect.TypeOf(targetItem).Kind() == reflect.Map &&
					item != nil && reflect.TypeOf(item).Kind() == reflect.Map {
					targetArr[i] = Merge(targetItem, item, options)
				} else {
					targetArr = append(targetArr, item)
				}
			} else {
				targetArr = append(targetArr, item)
			}
		}
		return targetArr
	}

	if mt, ok := mergeTarget.(map[string]any); ok {
		if s, ok := source.(map[string]any); ok {
			for key, value := range s {
				if value == nil {
					continue
				}

				if existingVal, exists := mt[key]; exists {
					mt[key] = Merge(existingVal, value, options)
				} else {
					mt[key] = value
				}
			}
		}
	} else if mt, ok := mergeTarget.(map[int]any); ok {
		result := make(map[string]any)
		for k, v := range mt {
			result[strconv.Itoa(k)] = v
		}

		if s, ok := source.(map[string]any); ok {
			for key, value := range s {
				if value == nil {
					continue
				}

				if existingVal, exists := result[key]; exists {
					result[key] = Merge(existingVal, value, options)
				} else {
					result[key] = value
				}
			}
		}
		return result
	}

	return mergeTarget
}

func AsString(v interface{}) string {
	switch s := v.(type) {
	case string:
		return s
	case bool:
		if s {
			return "true"
		}
		return "false"
	case int:
		return strconv.Itoa(s)
	case float64:
		return strconv.FormatFloat(s, 'f', -1, 64)
	case fmt.Stringer:
		return s.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func Assign(target, source map[string]interface{}) map[string]interface{} {
	for k, v := range source {
		target[k] = v
	}
	return target
}

var charset = "utf-8"

func Decode(str string) string {
	str = strings.ReplaceAll(str, "+", " ")
	if charset == "iso-8859-1" {

		decoded, err := url.QueryUnescape(str)
		if err != nil {
			return str
		}
		return decoded
	}
	decoded, err := url.QueryUnescape(str)
	if err != nil {
		return str
	}
	return decoded
}

const limit = 1024

// encode percent-encodes a string according to RFC3986 or RFC1738
func Encode(str, charset, kind, format string) string {
	if len(str) == 0 {
		return str
	}
	var out strings.Builder
	for j := 0; j < len(str); j += limit {
		segment := str
		if len(str) >= limit {
			end := j + limit
			if end > len(str) {
				end = len(str)
			}
			segment = str[j:end]
		}
		for _, r := range segment {
			c := r
			if c == '-' || c == '.' || c == '_' || c == '~' ||
				(c >= '0' && c <= '9') ||
				(c >= 'A' && c <= 'Z') ||
				(c >= 'a' && c <= 'z') ||
				(format == string(RFC1738) && (c == '(' || c == ')')) {
				out.WriteRune(c)
				continue
			}
			if c < 0x80 {
				out.WriteString(hexTable[c])
				continue
			}
			buf := make([]byte, 4)
			n := utf8.EncodeRune(buf, c)
			for i := 0; i < n; i++ {
				out.WriteString(hexTable[buf[i]])
			}
		}
	}
	return out.String()
}

func Contains(slice []any, v any) bool {
	for _, item := range slice {
		if reflect.DeepEqual(item, v) {
			return true
		}
	}
	return false
}

// isRegExp checks if obj is a regexp
func IsRegExp(obj interface{}) bool {
	return reflect.TypeOf(obj).String() == "*regexp.Regexp"
}

// isBuffer checks if obj is a []byte
func IsBuffer(obj interface{}) bool {
	_, ok := obj.([]byte)
	return ok
}

// combine concatenates two slices
func Combine(a, b []interface{}) []interface{} {
	return append(a, b...)
}

// maybeMap applies fn to val or each element if val is a slice
func MaybeMap(val interface{}, fn func(interface{}) interface{}) interface{} {
	if arr, ok := val.([]interface{}); ok {
		mapped := make([]interface{}, len(arr))
		for i, v := range arr {
			mapped[i] = fn(v)
		}
		return mapped
	}
	return fn(val)
}

// parseObject builds the nested object from key chain
func parseObject(chain []string, val interface{}, options ParseOptions, valuesParsed bool) interface{} {
	currentArrayLength := 0
	if len(chain) > 0 && chain[len(chain)-1] == "[]" {
		parentKey := strings.Join(chain[:len(chain)-1], "")
		if m, ok := val.(map[string]interface{}); ok {
			if arr, ok := m[parentKey].([]interface{}); ok {
				currentArrayLength = len(arr)
			}
		}
	}
	var leaf interface{}
	if valuesParsed {
		leaf = val
	} else {
		if strVal, ok := val.(string); ok {
			leaf = parseArrayValue(strVal, options, currentArrayLength)
		} else {
			leaf = val
		}
	}

	for i := len(chain) - 1; i >= 0; i-- {
		root := chain[i]
		var obj interface{}
		if root == "[]" && options.ParseArrays {
			if options.AllowEmptyArrays && (leaf == "" || (options.StrictNullHandling && leaf == nil)) {
				obj = []interface{}{}
			} else {
				if arr, ok := leaf.([]interface{}); ok {
					obj = Combine([]interface{}{}, arr)
				} else {
					obj = Combine([]interface{}{}, []interface{}{leaf})
				}
			}
		} else {
			m := map[string]interface{}{}
			cleanRoot := root
			if strings.HasPrefix(root, "[") && strings.HasSuffix(root, "]") {
				cleanRoot = root[1 : len(root)-1]
			}
			decodedRoot := cleanRoot
			if options.DecodeDotInKeys {
				decodedRoot = strings.ReplaceAll(cleanRoot, "%2E", ".")
			}
			index, err := strconv.Atoi(decodedRoot)
			if err == nil && options.ParseArrays && index >= 0 && index <= options.ArrayLimit {
				// shiftedIndex := index
				// if index > 0 {
				// 	shiftedIndex = 0
				// }
				// arr := make([]interface{}, shiftedIndex+1)
				// arr[shiftedIndex] = leaf
				arr := make([]interface{}, index+1)
				arr[index] = leaf
				obj = arr
			} else if decodedRoot != "__proto__" {
				m[decodedRoot] = leaf
				obj = m
			}
		}
		leaf = obj
	}
	return leaf
}

func processNestedStructures(obj interface{}, options *ParseOptions) interface{} {
	if obj == nil {
		return nil
	}
	if m, ok := obj.(map[string]interface{}); ok {
		for k, v := range m {
			m[k] = processNestedStructures(v, options)
		}
		if shouldConvertToArray(m) {
			return convertToQsArrayFormat(m)
		}
		return m
	}
	if arr, ok := obj.([]interface{}); ok {
		newArr := make([]interface{}, 0, len(arr))
		for _, v := range arr {
			if v == nil && !options.AllowNilArrayValues {
				continue // ignora nil
			}
			newArr = append(newArr, processNestedStructures(v, options))
		}

		if needsToBeObject(newArr) {
			return ArrayToObject(newArr, true)
		}
		return newArr
	}

	if b, ok := obj.(bool); ok {
		if b {
			return "true"
		}
		return "false"
	}
	return obj
}

func shouldConvertToArray(m map[string]interface{}) bool {
	hasNum := false
	hasOps := false
	hasBool := false
	hasObj := false
	for k, v := range m {
		if _, err := strconv.Atoi(k); err == nil {
			hasNum = true
		}
		if b, ok := v.(bool); ok && b {
			hasBool = true
		}
		if _, ok := v.(map[string]interface{}); ok {
			hasObj = true
		}
	}
	return (hasNum && hasOps) || (hasBool && hasObj)
}

func Compact(value interface{}) interface{} {
	if arr, ok := value.([]interface{}); ok {
		for len(arr) > 0 && arr[len(arr)-1] == nil {
			arr = arr[:len(arr)-1]
		}
		return arr
	}
	if m, ok := value.(map[string]interface{}); ok {
		for k, v := range m {
			m[k] = Compact(v)
		}
		return m
	}
	return value
}
