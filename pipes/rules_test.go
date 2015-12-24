package pipes

import (
	"fmt"
	"github.com/advancedlogic/goquery"
	"log"
	"reflect"
	"strings"
	"testing"
)

type Author struct {
	Avatar  string `json:"avatar"`
	Name    string `json:"name"`
	Profile string `json:"profile"`
}

var (
	DATA string = "<html><head><title>title</title><body><div class=\"date\">yesterday 8:15</div><div class=\"article\"><p>article body</p>" +
		"<p class=\"author\"><img src=\"avatar.jpg\"/><a href=\"/profile\">Name Surname</a></p>" +
		"</div></body></html>"
)

func TestCompiler(t *testing.T) {
	/* expE[find<expA["title"]>,first,text,as<expA>*/
	PIPE0 := "find(\"title\") | first | text | as(\"title\")"                                                  // extract title
	PIPE1 := "find(\"div.date\") | first | text | as(\"date\")"                                                // extract date
	PIPE2 := "find(\"div.article\") | store(\"content\") | find(\"p.author\") | remove | restore(\"content\")" // extract article body
	PIPE3 := "find(\"p.author\") | struct(\"avatar\",\"name\")" +
		"{ " +
		"(find(\"img\") | first | attr(\"src\") | as(\"avatar\")) " +
		"(find(\"a\") | first | text | as(\"name\"))" +
		"} | as(\"author\")"
	PIPE4 := "store(\"content\") | find(\"p.author\") | clone(\"profile\") | remove | restore(\"content\") | find(\"div.article\")"
	PIPE5 := "restore(\"profile\") | find(\"img\") | first | attr(\"src\")"

	compilePipe(PIPE0)
	compilePipe(PIPE3)
	r := &Rules{}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(DATA))
	if err != nil {
		log.Printf("error reading html doc")
		panic(":(")
	}
	saved := doc.Clone()
	d := doc.Selection
	rt := NewPipeRuntime(d, false)
	p := r.Compile(PIPE0)
	if p == nil {
		t.Fail()
		return
	}
	fmt.Printf("compiled pipe0:\n%+v\n", p)
	r.execChain(rt, d, p)
	log.Printf("PIPE0 result: %s", rt.String())
	if title, ok := rt.getVariable("title"); ok {
		if title == "title" {
			log.Printf("PIPE0 test PASSED")
		} else {
			log.Printf("PIPE0 test FAILED: WRONG RESULT '%s' (%v)", title, reflect.TypeOf(title))
			t.Fail()
		}
	} else {
		log.Printf("PIPE0 test FAILED: NO RESULT ")
		log.Printf("exists results: %+v", rt.vars)
		t.Fail()
	}

	p = r.Compile(PIPE1)
	if p == nil {
		t.Fail()
		return
	}
	d = saved.Clone()
	r.execChain(rt, d, p)
	log.Printf("PIPE1 result: %s", rt.String())
	p = r.Compile(PIPE2)
	if p == nil {
		t.Fail()
		return
	}
	d = saved.Clone()
	result2 := r.execChain(rt, d, p)
	if result2.getType() == "selection" {
		html, err := result2.Selection().Html()
		if err == nil {
			log.Printf("PIPE2 result (sanitize): %+v -> %s", result2.Selection(), html)
		} else {
			log.Printf("PIPE2 test failed")
			t.Fail()
		}
	} else {
		log.Printf("PIPE2 return not selection!")
	}
	p = r.Compile(PIPE3)
	if p == nil {
		t.Fail()
		return
	}
	d = saved.Clone()
	res := r.execChain(rt, d, p)
	log.Printf("PIPE3 result: %s\nruntime:%v", res.String(), rt.vars)
	p = r.Compile(PIPE4)
	d = saved.Clone()
	var html string
	result3 := r.execChain(rt, d, p)
	if result3.getType() == "selection" {
		html, err = result3.Selection().Html()
		if err != nil {
			log.Printf("PIPE4 WRONG RESULT")
		}
	} else {
		html = ":("
		t.Fail()
	}

	if err == nil {
		log.Printf("PIPE4 result (sanitize): %s\nruntime: %v\n", html, rt.String())
	} else {
		log.Printf("PIPE4 test failed")
		t.Fail()
	}

	p = r.Compile(PIPE5)
	d = saved.Clone()
	result := r.execChain(rt, d, p)
	log.Printf("PIPE5 result='%s', vars=%+v", result, rt.vars)

}

func TestPipeExpressions(t *testing.T) {
	compileAndRun("(find(\"div.article\")|first|push|(find(\"p.author\")|remove|pop)", DATA, t)

}

func TestPipeConcat(t *testing.T) {
	compileAndRun(
		"concat{ (clone(\"doc\")|find(\"div.article\")|first|push|(find(\"p.author\")|remove|pop) (restore(\"doc\")|find(\"p.author\")|first) }", DATA, t)
}

func compileAndRun(pipe string, data string, t *testing.T) {
	r := &Rules{}
	p := r.Compile(pipe)
	if p == nil {
		t.Fail()
	}
	d := getDoc(data, t)
	rt := NewPipeRuntime(d.Selection, false)
	result := r.execChain(rt, d.Selection, p)
	log.Printf("result: '''%s'''", docHtml(result, t))
}

func docHtml(sel IPipeArgument, t *testing.T) string {

	if sel.getType() == "selection" {
		html, _ := sel.Selection().Html()
		return html
	}
	t.Fail()
	return ""
}

func getDoc(data string, t *testing.T) *goquery.Document {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(data))
	if err == nil {
		return doc
	}
	t.Fail()
	return nil
}
