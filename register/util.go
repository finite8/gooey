package register

import (
	"fmt"
	"strings"
)

var attribOrder = []string{"rel", "href"}

func MapToAttributes(m map[string]interface{}) (retArr string) {
	sb := strings.Builder{}
	alreadyLoaded := make(map[string]bool)
	for _, a := range attribOrder {
		if v, ok := m[a]; ok && v != nil {
			sb.WriteString(fmt.Sprintf(`%s="%s"`, a, getInterfaceString(v)))
			alreadyLoaded[a] = true
		}
	}
	for k, v := range m {
		if _, ok := alreadyLoaded[k]; !ok && v != nil {
			sb.WriteString(fmt.Sprintf(`%s="%s"`, k, getInterfaceString(v)))
		}
	}
	if sb.Len() > 0 {
		return " " + sb.String()
	}
	return ""
}

func getInterfaceString(v interface{}) string {
	switch vt := v.(type) {
	case *string:
		return *vt
	case string:
		return vt
	default:
		return fmt.Sprintf("%v", v)
	}
}
