package lib

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/interpreter/functions"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
	"log"
	"math/rand"
	"net/url"
	"regexp"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewEnv(c *CustomLib) (*cel.Env, error) {
	return cel.NewEnv(cel.Lib(c))
}

func Calculate(env *cel.Env, expression string, params map[string]interface{}) (ref.Val, error) {
	ast, iss := env.Compile(expression)
	if iss.Err() != nil {
		log.Println(iss.Err())
		return nil, iss.Err()
	}

	prg, err := env.Program(ast)
	if err != nil {
		log.Printf("Program creation error: %v\n", err)
		return nil, err
	}

	out, _, err := prg.Eval(params)
	if err != nil {
		log.Printf("Evaluation error: %v\n", err)
		return nil, err
	}
	return out, nil
}

func ParseUrlType(u *UrlType) string {
	var buf strings.Builder
	if u.Scheme != "" {
		buf.WriteString(u.Scheme)
		buf.WriteByte(':')
	}
	if u.Scheme != "" || u.Host != "" {
		if u.Host != "" || u.Path != "" {
			buf.WriteString("//")
		}
		if h := u.Host; h != "" {
			buf.WriteString(u.Host)
		}
	}
	path := u.Path
	if path != "" && path[0] != '/' && u.Host != "" {
		buf.WriteByte('/')
	}
	if buf.Len() == 0 {
		if i := strings.IndexByte(path, ':'); i > -1 && strings.IndexByte(path[:i], '/') == -1 {
			buf.WriteString("./")
		}
	}
	buf.WriteString(path)

	if u.Query != "" {
		buf.WriteByte('?')
		buf.WriteString(u.Query)
	}
	if u.Fragment != "" {
		buf.WriteByte('#')
		buf.WriteString(u.Fragment)
	}
	return buf.String()
}

type CustomLib struct {
	envOptions     []cel.EnvOption
	programOptions []cel.ProgramOption
}

func NewEnvOption() CustomLib {
	c := CustomLib{}

	c.envOptions = []cel.EnvOption{
		cel.Container("lib"),
		cel.Types(
			&UrlType{},
			&Request{},
			&Response{},
		),
		cel.Declarations(
			decls.NewIdent("request", decls.NewObjectType("lib.Request"), nil),
			decls.NewIdent("response", decls.NewObjectType("lib.Response"), nil),
		),
		cel.Declarations(
			// request
			//decls.NewIdent("request.url.scheme", decls.String, nil),
			//decls.NewIdent("request.url.domain", decls.String, nil),
			//decls.NewIdent("request.url.host", decls.String, nil),
			//decls.NewIdent("request.url.port", decls.String, nil),
			//decls.NewIdent("request.url.path", decls.String, nil),
			//decls.NewIdent("request.url.query", decls.String, nil),
			//decls.NewIdent("request.url.fragment", decls.String, nil),
			//decls.NewIdent("request.method", decls.String, nil),
			//decls.NewIdent("request.headers", decls.NewMapType(decls.String, decls.String), nil),
			//decls.NewIdent("request.content_type", decls.String, nil),
			//decls.NewIdent("request.body", decls.Bytes, nil),
			//// response
			//decls.NewIdent("response.url.scheme", decls.String, nil),
			//decls.NewIdent("response.url.domain", decls.String, nil),
			//decls.NewIdent("response.url.host", decls.String, nil),
			//decls.NewIdent("response.url.port", decls.String, nil),
			//decls.NewIdent("response.url.path", decls.String, nil),
			//decls.NewIdent("response.url.query", decls.String, nil),
			//decls.NewIdent("response.url.fragment", decls.String, nil),
			//decls.NewIdent("response.status", decls.Int, nil),
			//decls.NewIdent("response.headers", decls.NewMapType(decls.String, decls.String), nil),
			//decls.NewIdent("response.content_type", decls.String, nil),
			//decls.NewIdent("response.body", decls.Bytes, nil),
			// functions
			decls.NewFunction("bcontains",
				decls.NewInstanceOverload("bytes_bcontains_bytes",
					[]*exprpb.Type{decls.Bytes, decls.Bytes},
					decls.Bool)),
			decls.NewFunction("bmatchs",
				decls.NewInstanceOverload("string_bmatchs_bytes",
					[]*exprpb.Type{decls.String, decls.Bytes},
					decls.Bool)),
			decls.NewFunction("md5",
				decls.NewOverload("md5_string",
					[]*exprpb.Type{decls.String},
					decls.String)),
			decls.NewFunction("randomInt",
				decls.NewOverload("randomInt_int_int",
					[]*exprpb.Type{decls.Int, decls.Int},
					decls.Int)),
			decls.NewFunction("randomLowercase",
				decls.NewOverload("randomLowercase_int",
					[]*exprpb.Type{decls.Int},
					decls.String)),
			decls.NewFunction("base64",
				decls.NewOverload("base64_string",
					[]*exprpb.Type{decls.String},
					decls.String)),
			decls.NewFunction("base64",
				decls.NewOverload("base64_bytes",
					[]*exprpb.Type{decls.Bytes},
					decls.String)),
			decls.NewFunction("base64Decode",
				decls.NewOverload("base64Decode_string",
					[]*exprpb.Type{decls.String},
					decls.String)),
			decls.NewFunction("base64Decode",
				decls.NewOverload("base64Decode_bytes",
					[]*exprpb.Type{decls.Bytes},
					decls.String)),
			decls.NewFunction("urlencode",
				decls.NewOverload("urlencode_string",
					[]*exprpb.Type{decls.String},
					decls.String)),
			decls.NewFunction("urlencode",
				decls.NewOverload("urlencode_bytes",
					[]*exprpb.Type{decls.Bytes},
					decls.String)),
			decls.NewFunction("urldecode",
				decls.NewOverload("urldecode_string",
					[]*exprpb.Type{decls.String},
					decls.String)),
			decls.NewFunction("urldecode",
				decls.NewOverload("urldecode_bytes",
					[]*exprpb.Type{decls.Bytes},
					decls.String)),
			decls.NewFunction("substr",
				decls.NewOverload("substr_string_int_int",
					[]*exprpb.Type{decls.String, decls.Int, decls.Int},
					decls.String)),
		),
	}
	c.programOptions = []cel.ProgramOption{
		cel.Functions(
			&functions.Overload{
				Operator: "bytes_bcontains_bytes",
				Binary: func(lhs ref.Val, rhs ref.Val) ref.Val {
					v1, ok := lhs.(types.Bytes)
					if !ok {
						return types.ValOrErr(lhs, "unexpected type '%v' passed to bmatch", lhs.Type())
					}
					v2, ok := rhs.(types.Bytes)
					if !ok {
						return types.ValOrErr(rhs, "unexpected type '%v' passed to bmatch", rhs.Type())
					}
					return types.Bool(bytes.Contains(v1, v2))
				},
			},
		),
		cel.Functions(
			&functions.Overload{
				Operator: "string_bmatch_bytes",
				Binary: func(lhs ref.Val, rhs ref.Val) ref.Val {
					v1, ok := lhs.(types.String)
					if !ok {
						return types.ValOrErr(lhs, "unexpected type '%v' passed to bmatch", lhs.Type())
					}
					v2, ok := rhs.(types.Bytes)
					if !ok {
						return types.ValOrErr(rhs, "unexpected type '%v' passed to bmatch", rhs.Type())
					}
					ok, err := regexp.Match(string(v1), v2)
					if err != nil {
						return types.NewErr("%v", err)
					}
					return types.Bool(ok)
				},
			},
		),
		cel.Functions(
			&functions.Overload{
				Operator: "md5_string",
				Unary: func(value ref.Val) ref.Val {
					v, ok := value.(types.String)
					if !ok {
						return types.ValOrErr(value, "unexpected type '%v' passed to md5_string", value.Type())
					}
					return types.String(fmt.Sprintf("%x", md5.Sum([]byte(v))))
				},
			},
		),
		cel.Functions(
			&functions.Overload{
				Operator: "randomInt_int_int",
				Binary: func(lhs ref.Val, rhs ref.Val) ref.Val {
					to, ok := lhs.(types.Int)
					if !ok {
						return types.ValOrErr(lhs, "unexpected type '%v' passed to randomInt", lhs.Type())
					}
					from, ok := rhs.(types.Int)
					if !ok {
						return types.ValOrErr(rhs, "unexpected type '%v' passed to randomInt", rhs.Type())
					}
					return types.Int(rand.Intn(int(to)) + int(from))
				},
			}),
		cel.Functions(&functions.Overload{
			Operator: "randomLowercase_int",
			Unary: func(value ref.Val) ref.Val {
				n, ok := value.(types.Int)
				if !ok {
					return types.ValOrErr(value, "unexpected type '%v' passed to randomLowercase", value.Type())
				}
				return types.String(randomLowercase(int(n)))
			},
		}),
		cel.Functions(&functions.Overload{
			Operator: "base64_string",
			Unary: func(value ref.Val) ref.Val {
				v, ok := value.(types.String)
				if !ok {
					return types.ValOrErr(value, "unexpected type '%v' passed to base64_string", value.Type())
				}
				return types.String(base64.StdEncoding.EncodeToString([]byte(v)))
			},
		}),
		cel.Functions(&functions.Overload{
			Operator: "base64_bytes",
			Unary: func(value ref.Val) ref.Val {
				v, ok := value.(types.Bytes)
				if !ok {
					return types.ValOrErr(value, "unexpected type '%v' passed to base64_bytes", value.Type())
				}
				return types.String(base64.StdEncoding.EncodeToString(v))
			},
		}),
		cel.Functions(&functions.Overload{
			Operator: "base64Decode_string",
			Unary: func(value ref.Val) ref.Val {
				v, ok := value.(types.String)
				if !ok {
					return types.ValOrErr(value, "unexpected type '%v' passed to base64Decode_string", value.Type())
				}
				decodeBytes, err := base64.StdEncoding.DecodeString(string(v))
				if err != nil {
					return types.NewErr("%v", err)
				}
				return types.String(string(decodeBytes))
			},
		}),
		cel.Functions(&functions.Overload{
			Operator: "base64Decode_bytes",
			Unary: func(value ref.Val) ref.Val {
				v, ok := value.(types.Bytes)
				if !ok {
					return types.ValOrErr(value, "unexpected type '%v' passed to base64Decode_bytes", value.Type())
				}
				decodeBytes, err := base64.StdEncoding.DecodeString(string(v))
				if err != nil {
					return types.NewErr("%v", err)
				}
				return types.String(string(decodeBytes))
			},
		}),
		cel.Functions(&functions.Overload{
			Operator: "urlencode_string",
			Unary: func(value ref.Val) ref.Val {
				v, ok := value.(types.String)
				if !ok {
					return types.ValOrErr(value, "unexpected type '%v' passed to urlencode_string", value.Type())
				}
				return types.String(url.QueryEscape(string(v)))
			},
		}),
		cel.Functions(&functions.Overload{
			Operator: "urlencode_bytes",
			Unary: func(value ref.Val) ref.Val {
				v, ok := value.(types.Bytes)
				if !ok {
					return types.ValOrErr(value, "unexpected type '%v' passed to urlencode_bytes", value.Type())
				}
				return types.String(url.QueryEscape(string(v)))
			},
		}),
		cel.Functions(&functions.Overload{
			Operator: "urldecode_string",
			Unary: func(value ref.Val) ref.Val {
				v, ok := value.(types.String)
				if !ok {
					return types.ValOrErr(value, "unexpected type '%v' passed to urldecode_string", value.Type())
				}
				decodeString, err := url.QueryUnescape(string(v))
				if err != nil {
					return types.NewErr("%v", err)
				}
				return types.String(decodeString)
			},
		}),
		cel.Functions(&functions.Overload{
			Operator: "urldecode_bytes",
			Unary: func(value ref.Val) ref.Val {
				v, ok := value.(types.Bytes)
				if !ok {
					return types.ValOrErr(value, "unexpected type '%v' passed to urldecode_bytes", value.Type())
				}
				decodeString, err := url.QueryUnescape(string(v))
				if err != nil {
					return types.NewErr("%v", err)
				}
				return types.String(decodeString)
			},
		}),
		cel.Functions(&functions.Overload{
			Operator: "substr_string_int_int",
			Function: func(values ...ref.Val) ref.Val {
				if len(values) == 3 {
					str, ok := values[0].(types.String)
					if !ok {
						return types.NewErr("invalid string to 'substr'")
					}
					start, ok := values[1].(types.Int)
					if !ok {
						return types.NewErr("invalid start to 'substr'")
					}
					length, ok := values[1].(types.Int)
					if !ok {
						return types.NewErr("invalid length to 'substr'")
					}
					runes := []rune(str)
					if start < 0 || length < 0 || int(start+length) > len(runes) {
						return types.NewErr("invalid start or length to 'substr'")
					}
					return types.String(runes[start : start+length])
				} else {
					return types.NewErr("too many arguments to 'substr'")
				}
			},
		}),
	}
	return c
}

// 声明环境中的变量类型和函数
func (c *CustomLib) CompileOptions() []cel.EnvOption {
	return c.envOptions
}

func (c *CustomLib) ProgramOptions() []cel.ProgramOption {
	return c.programOptions
}

func (c *CustomLib) UpdateCompileOptions(args map[string]string) {
	for k := range args {
		d := decls.NewIdent(k, decls.String, nil)
		c.envOptions = append(c.envOptions, cel.Declarations(d))
	}
}

//func (c *CustomLib) UpdateProgramOptions() {
//
//}

func randomLowercase(n int) string {
	lowercase := "abcdefghijklmnopqrstuvwxyz"
	randSource := rand.New(rand.NewSource(time.Now().Unix()))
	return randomStr(randSource, lowercase, n)
}

func randomStr(randSource *rand.Rand, letterBytes string, n int) string {
	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)
	randBytes := make([]byte, n)
	for i, cache, remain := n-1, randSource.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = randSource.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			randBytes[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(randBytes)
}