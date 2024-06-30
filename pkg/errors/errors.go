package errors

import (
	"strings"
	"sync"
	"text/template"

	"github.com/soulnov23/go-tool/pkg/json/jsoniter"
)

//go:generate protoc --proto_path=. --go_out=paths=source_relative:. --validate_out=lang=go,paths=source_relative:. errors.proto

var templateCache sync.Map

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return jsoniter.Stringify(e)
}

func (e *Error) WithMessageValues(values any) *Error {
	var (
		tpl *template.Template
		err error
	)
	value, ok := templateCache.Load(e.Message)
	if !ok {
		tpl, err = template.New(e.Name).Parse(e.Message)
		if err != nil {
			return e
		}
		templateCache.Store(e.Message, tpl)
	}
	builder := &strings.Builder{}
	if err := value.(*template.Template).Execute(builder, values); err != nil {
		return e
	}
	e.Message = builder.String()
	return e
}

func Parse(err string) *Error {
	e := &Error{}
	if errr := jsoniter.UnmarshalFromString(err, e); errr != nil {
		return nil
	}
	return e
}

func FromError(err error) *Error {
	if err == nil {
		return nil
	}
	if innerErr, ok := err.(*Error); ok && innerErr != nil {
		return innerErr
	}
	return Parse(err.Error())
}

func Equal(err1 error, err2 error) bool {
	verr1, ok1 := err1.(*Error)
	verr2, ok2 := err2.(*Error)

	if ok1 != ok2 {
		return false
	}

	if !ok1 {
		return err1 == err2
	}

	if verr1.Code != verr2.Code {
		return false
	}

	if verr1.Name != verr2.Name {
		return false
	}

	return true
}
