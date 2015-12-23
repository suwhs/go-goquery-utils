package pipes

import (
	"fmt"
	"github.com/advancedlogic/goquery"
	"strings"
)

/*
   generate pipe to extract content from rawHtml to calculate result similar to targetHtml
*/
func SuggestRules(rawHtml string, targetHtml string) []IPipeEntry {

	targetDoc, err := goquery.NewDocumentFromReader(strings.NewReader(targetHtml))
	if err != nil {
		panic(fmt.Sprintf("error: %s", err.Error()))
	}
	sourceDoc, err := goquery.NewDocumentFromReader(strings.NewReader(rawHtml))
	if err != nil {
		panic(fmt.Sprintf("error: %s", err.Error()))
	}
	if targetDoc == sourceDoc {
		return nil
	}
	result := make([]IPipeEntry, 0, 10)
	return result
}
