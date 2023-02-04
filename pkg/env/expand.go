// Package env 环境变量
package env

import "os"

// ExpandEnv 使用环境变量中的值替换源字符串中的${var}
func ExpandEnv(s string) string {
	var buf []byte
	i := 0
	for j := 0; j < len(s); j++ {
		if s[j] == '$' && j+2 < len(s) && s[j+1] == '{' { // only ${var} instead of $var is valid
			if buf == nil {
				buf = make([]byte, 0, 2*len(s))
			}
			buf = append(buf, s[i:j]...)
			name, w := getEnvName(s[j+1:])
			if name == "" && w > 0 {
				// invalid matching, remove the $
			} else if name == "" {
				buf = append(buf, s[j]) // keep the $
			} else {
				buf = append(buf, os.Getenv(name)...)
			}
			j += w
			i = j + 1
		}
	}
	if buf == nil {
		return s
	}
	return string(buf) + s[i:]
}

func getEnvName(s string) (string, int) {
	// look for }
	// it's guaranteed that the first char is { and the string has at least two char
	for i := 1; i < len(s); i++ {
		if s[i] == ' ' || s[i] == '\n' || s[i] == '"' { // "xx${xxx"
			return "", 0 // encounter invalid char, keep the $
		}
		if s[i] == '}' {
			if i == 1 { // ${}
				return "", 2 // remove ${}
			}
			return s[1:i], i + 1
		}
	}
	return "", 0 // no }，keep the $
}
