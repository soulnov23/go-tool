package main

const templateErrors = `
// Code generated by github.com/soulnov23/go-tool/pkg/errors/generator. DO NOT EDIT.
// source: {{.source}}

package {{.package}}

import (
	"github.com/soulnov23/go-tool/pkg/errors"
)

var (
{{- range $config := .configs}}
	{{$config.Name}} = &errors.Error{
		Code:    {{$config.Code}},
		Status:  "{{$config.Status}}",
		Name:    "{{$config.Name}}",
		Message: "{{$config.Message}}",
	}
{{- end}}
)
`
