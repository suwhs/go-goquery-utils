package pipes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/advancedlogic/goquery"
	// "golang.org/x/net/html"
	//	"log"
	"reflect"
	// "regexp"
	"strconv"
	// "strings"
	// "text/scanner"
)

type PipeRuntime struct {
	doc       *goquery.Selection
	vars      map[string]interface{}
	immutable bool
	funcs     map[string]*PipeAction
	stack     []IPipeArgument
}

func NewPipeRuntime(d *goquery.Selection, im bool) *PipeRuntime {
	return &PipeRuntime{
		doc:       d,
		immutable: im,
		vars:      make(map[string]interface{}),
		stack:     make([]IPipeArgument, 0, 10),
		funcs:     make(map[string]*PipeAction),
	}
}

func (r *PipeRuntime) Copy() *PipeRuntime {
	vars := make(map[string]interface{})
	for k, v := range r.vars {
		vars[k] = v
	}
	return &PipeRuntime{
		doc:       r.Peek().Selection(),
		immutable: false,
		vars:      vars,
		stack:     make([]IPipeArgument, 0, 10),
		funcs:     r.funcs,
	}
}

func (r *PipeRuntime) getVariable(name string) (interface{}, bool) {
	if _, ok := r.vars[name]; ok {
		return r.vars[name], true
	}
	return nil, false
}

func (r *PipeRuntime) String() string {
	var buffer bytes.Buffer
	for k, v := range r.vars {
		buffer.WriteString(fmt.Sprintf("%s='%s'\n", k, v))
	}
	return buffer.String()
}

func (r *PipeRuntime) Push(a IPipeArgument) IPipeArgument {
	r.stack = append(r.stack, a)
	return a
}

func (r *PipeRuntime) Pop() IPipeArgument {
	entry := r.stack[len(r.stack)-1]
	r.stack = r.stack[:len(r.stack)-1]
	return entry
}

func (r *PipeRuntime) Peek() IPipeArgument {
	if len(r.stack) < 1 {
		return NewSelectionArgument(r.doc)
	}
	return r.stack[len(r.stack)-1]
}

func (r *PipeRuntime) Call(funcname string, arg IPipeArgument) IPipeArgument {
	if fn, ok := r.funcs[funcname]; ok {
		return fn.Exec(r, arg)
	}
	panic(fmt.Sprintf("function '%s' undefined", funcname))
}

/**
	operator (arguments) {block}
	PipeEntry
		PipeExpression
		PipeArgument(PipeExpression)
		PipeAction
			PipeArgument
			[]PipeEntry

**/

type IPipeStructWrapper interface {
	WrapStruct(runtime *PipeRuntime, p *PipeStruct, s *goquery.Selection) interface{}
}

// TODO:
// change:
// result of each ExecWith... pushed to runtime.Stack
// for Next().Exec() top entry on Stack used as argument
// chain result == top entry on Stack

func Exec(p IPipeEntry, runtime *PipeRuntime, arg IPipeArgument) IPipeArgument {
	return p.Exec(runtime, arg)
}

/*
   doc | PipeMatch -> object | list
   if result are list - rest of chain will be applied for each item
   doc | find("article") | first |
       find first entry of "article" tag, and returns it's content
   for habrahabr, for example:
       extract publish date:    `doc | find("div.published") | first | text | unhumanPublishDate `
       extract author: doc | find("div.profile_header") | first | struct("avatar","name") {
           (find("a.author-info_name") | first | text | as("name") )
           (find("img.author-info__image-pic") | first | text | as("avatar"))
       }`
       scoreContent : `doc | find("div.content.html_format") | first | score(100)`
       extractTitle : `doc | find("div.post.shortcuts_item") | first | find("h1.title") | first | text`

       pipe chain
        pipe {
            func
            next
        }

        we call
        param := this.exec(func(runtime,arg))
        and return next.exec(func(runtime,arg))
*/

type Rules struct {
	Host                string
	prepare             []IPipeEntry
	extractDate         []IPipeEntry
	extractAuthor       []IPipeEntry
	extractAuthorAvatar []IPipeEntry
	sanitize            []IPipeEntry
	extractTitle        []IPipeEntry
	scoreContent        []IPipeEntry
	extractContent      []IPipeEntry
	Source              *SingleRuleJSON
}

func (r *Rules) Compare(t *Rules) int {
	return 0
}

type SingleRuleJSON struct {
	Host           []string `json:"host"`
	Aliases        []string `json:"aliases"`
	Prepare        string   `json:"prepare"`
	ExtractDate    string   `json:"extract_date"`
	ExtractAuthor  string   `json:"extract_author"`
	ScoreContent   string   `json:"score_content"`
	ExtractTitle   string   `json:"extract_title"`
	ExtractContent string   `json:"extract_content"`
	Sanitize       string   `json:"sanitizie"`
}

type RulesScript struct {
	prepare             string `json:"prepare"`
	extractDate         string `json:"extract_date"`
	extractAuthor       string `json:"extract_author"`
	extractAuthorAvatar string `json:"extract_author_avatar"`
	sanitize            string `json:"sanitize_doc"`
	extractTitle        string `json:"extract_title"`
	scoreContent        string `json:"score_content"`
	extractContent      string `json:"extract_content"`
}

func (r *Rules) CompileAll(srj *SingleRuleJSON) {
	if srj.Prepare != "" {
		r.prepare = r.Compile(srj.Prepare)
	}
	if srj.ExtractAuthor != "" {
		r.extractAuthor = r.Compile(srj.ExtractAuthor)
	}
	if srj.ExtractDate != "" {
		r.extractDate = r.Compile(srj.ExtractDate)
	}
	if srj.ExtractTitle != "" {
		r.extractTitle = r.Compile(srj.ExtractTitle)
	}
	if srj.ExtractContent != "" {
		r.extractContent = r.Compile(srj.ExtractContent)
	}
	if srj.Sanitize != "" {
		r.sanitize = r.Compile(srj.Sanitize)
	}
}

func (r *Rules) Compile(pipe string) []IPipeEntry {
	exps := parsePipe(pipe)
	result := make([]IPipeEntry, len(exps))
	for i := range exps {
		result[i] = exps[i].Compile()
	}
	return result
}

func (r *Rules) UnmarshalJSON(data []byte) {
	script := &RulesScript{}
	if err := json.Unmarshal(data, script); err != nil {
		r.prepare = r.Compile(script.prepare)
		r.extractDate = r.Compile(script.extractDate)
		r.extractAuthor = r.Compile(script.extractAuthor)
		r.extractAuthorAvatar = r.Compile(script.extractAuthorAvatar)
		r.sanitize = r.Compile(script.sanitize)
		r.extractTitle = r.Compile(script.extractTitle)
		r.extractContent = r.Compile(script.extractContent)
	}
}

type Metadata struct {
	date         string
	author       string
	authorAvatar string
	title        string
	content      *goquery.Selection
}

func (m *Metadata) GetDate() string                { return m.date }
func (m *Metadata) GetAuthorName() string          { return m.author }
func (m *Metadata) GetAuthorAvatar() string        { return m.authorAvatar }
func (m *Metadata) GetAuthorProfile() string       { return "" }
func (m *Metadata) GetContentProviderName() string { return "" }
func (m *Metadata) GetContentProviderURL() string  { return "" }
func (m *Metadata) GetContent() *goquery.Selection { return m.content }
func (m *Metadata) GetTitle() string               { return m.title }
func (m *Metadata) HasContent() bool               { return m.content != nil }

func (r *Rules) PreExecute(doc *goquery.Selection) *Metadata {
	return r.preExecute(doc)
}

func (r *Rules) preExecute(doc *goquery.Selection) *Metadata {
	runtime := NewPipeRuntime(doc, false)
	md := &Metadata{}
	var result interface{}
	if r.prepare != nil {
		result = r.execChain(runtime, doc, r.prepare)
	}
	if r.extractDate != nil {
		result = r.execChain(runtime, doc, r.extractDate)
		if str, ok := result.(string); ok {
			md.date = str
		}
	}
	if r.extractAuthor != nil {
		result = r.execChain(runtime, doc, r.extractAuthor)
		if sm, ok := runtime.vars["author"].(map[string]interface{}); ok {
			md.author = fmt.Sprintf("%s", sm["name"])
			md.authorAvatar = fmt.Sprintf("%s", sm["avatar"])
		}
	}
	if r.extractTitle != nil {
		result = r.execChain(runtime, doc, r.extractTitle)
		if str, ok := result.(string); ok {
			md.title = str
		}
	}
	if r.scoreContent != nil {
		_ = r.execChain(runtime, doc, r.scoreContent)
	}
	if r.extractContent != nil {
		result := r.execChain(runtime, doc, r.extractContent)
		if result.getType() == "selection" {
			md.content = result.Selection()
		}
	}
	return md
}

func (r *Rules) execChain(rt *PipeRuntime, s *goquery.Selection, chain []IPipeEntry) IPipeArgument {
	var result IPipeArgument = NewSelectionArgument(s)
	for i := range chain {
		result = Exec(chain[i], rt, result)
	}
	return result
}

func (r *Rules) PostExecute(doc *goquery.Selection) *goquery.Selection {
	return r.postExecute(doc)
}

func (r *Rules) postExecute(doc *goquery.Selection) *goquery.Selection {
	runtime := NewPipeRuntime(doc, false)
	if r.sanitize != nil {
		result := r.execChain(runtime, doc, r.sanitize)
		if result.getType() == "selection" {
			return result.Selection()
		}
	}
	return doc
}

func Unmarshal(data map[string]string, S interface{}) {
	st := reflect.TypeOf(S).Elem()
	for i := 0; i < st.NumField(); i++ {
		field := st.Field(i)
		name := field.Tag.Get("json")
		strVal, ok := data[name]
		value := reflect.ValueOf(field)
		if !ok {
			continue
		}
		switch field.Type.Kind() {
		case reflect.Int:
			intVal, err := strconv.Atoi(strVal)
			if err == nil {
				value.SetInt(int64(intVal))
			}
		case reflect.String:
			value.SetString(strVal)
		}
	}
}
