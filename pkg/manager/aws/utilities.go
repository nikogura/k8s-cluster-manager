package aws

import (
	"fmt"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
	"reflect"
	"strings"
)

func prettyPrint[T any](str T) string {
	s := reflect.ValueOf(&str).Elem()
	typeOf := s.Type()
	pretty := ""
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		pretty = strings.Join([]string{pretty, fmt.Sprintf("%d: %s %s = %v\n", i,
			typeOf.Field(i).Name, f.Type(), f.Interface())}, "")
	}

	return pretty
}

func prettyPrintMap[T map[string]manager.NodeInfo](str T) string {
	pretty := ""
	i := 0
	for k, v := range str {
		pretty = strings.Join([]string{pretty, fmt.Sprintf("%d: %s =\n-\n | values for key %d:\n%s\n", i,
			k, i, prettyPrint(v))}, "")
		i++
	}

	return pretty
}
