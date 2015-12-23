package pipes

import (
	// "fmt"
	"github.com/advancedlogic/goquery"
	"golang.org/x/net/html"
	"strings"
)

func unquote(s string) string {
	if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		return s[1 : len(s)-1]
	}
	return s
}

func setAttr(s *goquery.Selection, attr string, value string) {
	if s.Size() > 0 {
		node := s.Get(0)
		attrs := make([]html.Attribute, 0)
		for _, a := range node.Attr {
			if a.Key != attr {
				newAttr := new(html.Attribute)
				newAttr.Key = a.Key
				newAttr.Val = a.Val
				attrs = append(attrs, *newAttr)
			}
		}
		newAttr := new(html.Attribute)
		newAttr.Key = attr
		newAttr.Val = value
		attrs = append(attrs, *newAttr)
		node.Attr = attrs
	}
}
