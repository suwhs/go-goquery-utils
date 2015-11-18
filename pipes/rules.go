package pipes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/advancedlogic/goquery"
	"golang.org/x/net/html"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/scanner"
)

type PipeRuntime struct {
	doc       *goquery.Selection
	vars      map[string]interface{}
	immutable bool
}

func NewPipeRuntime(d *goquery.Selection, im bool) *PipeRuntime {
	return &PipeRuntime{
		doc:       d,
		immutable: im,
		vars:      make(map[string]interface{}),
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

type IPipeEntry interface {
	Compile(tok rune, s *scanner.Scanner) IPipeEntry
	// Exec(runtime *PipeRuntime, object interface{}) interface{}
	ExecWithSelection(runtime *PipeRuntime, s *goquery.Selection) interface{}
	ExecWithString(runtime *PipeRuntime, str string) interface{}
	ExecWithPipeEntry(runtime *PipeRuntime, entry *IPipeEntry) interface{}
	String() string
}

type IPipeStructWrapper interface {
	WrapStruct(runtime *PipeRuntime, p *PipeStruct, s *goquery.Selection) interface{}
}

type PipeResult struct {
}

type PipeEmptyResult struct {
	PipeResult
}

type PipeEntry struct {
}

func (p *PipeEntry) String() string { return "PipeEntry" }
func (p *PipeEntry) Compile(tok rune, s *scanner.Scanner) IPipeEntry {
	return p
}

// TODO:
// change:
// result of each ExecWith... pushed to runtime.Stack
// for Next().Exec() top entry on Stack used as argument
// chain result == top entry on Stack

func Exec(p IPipeEntry, runtime *PipeRuntime, object interface{}) interface{} {
	var result interface{}
	if debug {
		log.Printf("Exec(%s,...)", p)
	}
	if s, ok := object.(*goquery.Selection); ok {
		result = p.ExecWithSelection(runtime, s)
	} else if entry, ok := object.(*IPipeEntry); ok {
		result = p.ExecWithPipeEntry(runtime, entry)
	} else if str, ok := object.(string); ok {
		if debug {
			log.Printf("exec with string '%s' ", str)
		}
		result = p.ExecWithString(runtime, str)
	} else if st, ok := object.(*PipeStruct); ok {
		if debug {
			log.Printf("argument are PipeStruct! %+v", st)
		}
		if wrapper, ok := p.(IPipeStructWrapper); ok {
			result = wrapper.WrapStruct(runtime, st, s)
		}

	}
	return result
}

func (p *PipeEntry) ExecWithSelection(runtime *PipeRuntime, s *goquery.Selection) interface{} {
	return nil
}

func (p *PipeEntry) ExecWithPipeEntry(runtime *PipeRuntime, entry *IPipeEntry) interface{} {
	return nil
}

func (p *PipeEntry) ExecWithString(runtime *PipeRuntime, str string) interface{} {
	return nil
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

type PipeMatch struct {
	PipeEntry
}

func (p *PipeMatch) Find(pattern string) interface{} {
	return nil
}

type PipeMatchRegexp struct {
	PipeMatch
	regexp *regexp.Regexp
	input  string
}

type PipeMatchXPath struct {
	PipeMatch
}

type PipeFind struct {
	PipeMatch
	selector string
}

func (p *PipeFind) Compile(tok rune, s *scanner.Scanner) IPipeEntry {
	/* tokens: (, "", ) */
	for s.Scan() != scanner.EOF && "(" != s.TokenText() {
	}

	if "(" == s.TokenText() {
		for s.Scan() != scanner.String {
		}
		p.selector = s.TokenText()
		p.selector = p.selector[1 : len(p.selector)-1]
		for s.TokenText() != ")" && s.Scan() != scanner.EOF {
		}
	}
	return p
}

func (p *PipeFind) String() string { return fmt.Sprintf("PipeFind('%s')", p.selector) }

func (p *PipeFind) ExecWithSelection(r *PipeRuntime, s *goquery.Selection) interface{} {
	result := s.Find(p.selector)
	if result.Length() < 1 {
		if debug {
			log.Printf("PipeFind() return empty list for ('%s')", p.selector)
		}
		html, err := s.Html()
		if err == nil {
			if debug {
				log.Printf("source html:\n%s\n", html)
			}
		}
	}
	return result
}

type PipeSelectGroupByNumber struct {
	PipeEntry
	groupNumber int
}

func (p *PipeSelectGroupByNumber) ExecWithPipeEntry(r *PipeRuntime, entry interface{}) interface{} {
	result := make([]string, 3)
	if rx, ok := entry.(*PipeMatchRegexp); ok {
		res := rx.regexp.FindAllStringSubmatch(rx.input, -1)
		for i := range res {
			sub := res[i]
			if len(sub) < p.groupNumber {
				result = append(result, sub[p.groupNumber])
			}
		}
	}
	return result
}

type PipeStruct struct {
	PipeEntry
	expected []string
	fields   map[string]interface{}
	chains   [][]IPipeEntry
}

func (p *PipeStruct) String() string {
	return fmt.Sprintf("PipeStruct(%+v [%+v])", p.fields, p.chains)
}

func (p *PipeStruct) Compile(tok rune, s *scanner.Scanner) IPipeEntry {
	p.expected = make([]string, 0, 5)
	/* struct format
	   struct(<names>) { <chains> }
	*/
	for s.Scan() != scanner.EOF && "(" != s.TokenText() {
	} // names section begins
	p.CompileNamesList(tok, s)
	for s.Scan() != scanner.EOF && "{" != s.TokenText() {
	} // chains section begins
	p.CompileChainsList(tok, s)
	return p
}

func (p *PipeStruct) CompileNamesList(tok rune, s *scanner.Scanner) {
	/* names format
	   "nameA"[,"nameB"]
	*/
	for s.TokenText() != ")" {
		tok = s.Scan()
		if tok == scanner.String {
			name := s.TokenText()
			name = name[1 : len(name)-1]
			p.expected = append(p.expected, name)
		} else if s.TokenText() == "," {
			continue
		}
	}
	if debug {
		log.Printf("PipeStruct(%v) fields built", p.expected)
	}
}

var debug bool = false

func (p *PipeStruct) CompileChainsList(tok rune, s *scanner.Scanner) {
	// theoretically
	oldPos := s.Pos()
	for s.Scan() != scanner.EOF && "(" != s.TokenText() {
	} // go to chain begin
	pipe := make([]IPipeEntry, 0, 5)
	for s.TokenText() != "}" {

		tok = s.Scan()
		if debug {
			log.Printf("start with: '%s'", s.TokenText())
		}
		entry := compilerFactoryDbg(tok, s)
		if debug {
			log.Printf("pipe entry='%s", entry)
		}
		if entry != nil {
			pipe = append(pipe, entry)
			tok = s.Scan() // shift out from last ')'
			continue
		} else if "|" == s.TokenText() {
			if debug {
				log.Printf("break on |, continue")
			}
			continue
		}
		if debug {
			log.Printf("chain finished...%+v at '%s'", pipe, s.TokenText())
		}
		p.chains = append(p.chains, pipe)
		if s.Pos() == oldPos {
			panic("infinite loop!")
		}
		oldPos = s.Pos()
		pipe = make([]IPipeEntry, 0, 5)
	}
}

func compilerFactoryDbg(tok rune, s *scanner.Scanner) IPipeEntry {
	if debug {
		log.Printf("debug factory: token='%s'", s.TokenText())
	}
	if "find" == s.TokenText() {
		return compileAndReturn(&PipeFind{}, tok, s)
	} else if "text" == s.TokenText() {
		return compileAndReturn(&PipeText{}, tok, s)
	} else if "first" == s.TokenText() {
		return compileAndReturn(&PipeFirst{}, tok, s)
	} else if "struct" == s.TokenText() {
		return compileAndReturn(&PipeStruct{}, tok, s)
	} else if "as" == s.TokenText() {
		return compileAndReturn(&PipeAs{}, tok, s)
		//return (&PipeAs{}).Compile(tok, s).(*PipeEntry)
	} else if "unhumanPublishDate" == s.TokenText() {
		return compileAndReturn(&PipeUnhumanDate{}, tok, s)
	} else if "remove" == s.TokenText() {
		return compileAndReturn(&PipeRemove{}, tok, s)
	} else if "attr" == s.TokenText() {
		return compileAndReturn(&PipeAttr{}, tok, s)
	} else if "store" == s.TokenText() {
		return compileAndReturn(&PipeStore{}, tok, s)
	} else if "clone" == s.TokenText() {
		return compileAndReturn(&PipeClone{}, tok, s)
	} else if "restore" == s.TokenText() {
		return compileAndReturn(&PipeRestore{}, tok, s)
	} else if "|" == s.TokenText() {

	} else if ")" == s.TokenText() {
		return nil
	}
	return nil
}

/**
run each chain from s
result of each chain must be as("<name>")

**/

func (p *PipeStruct) ExecWithSelection(r *PipeRuntime, s *goquery.Selection) interface{} {
	if debug {
		log.Printf("pipe chain: %+v", p.chains)
	}
	p.fields = make(map[string]interface{})
	sr := NewPipeRuntime(s, true)
	for i := range p.chains {

		var result interface{} = s
		for k := range p.chains[i] {
			result = Exec(p.chains[i][k], sr, result)
		}

		if name, ok := result.(*PipeAs); ok {
			found := false
			for l := range p.expected {
				if p.expected[l] == name.name {
					found = true
				}
			}
			if found {
				p.fields[name.name] = sr.vars[name.name]
			} else {
				// unknown name for field
				panic(fmt.Sprintf("unknown field '%s' for struct{%v}", name.name, p.fields))
			}
		} else {
			// fatal error - chain must ends with struct's field name
			panic(fmt.Sprintf("struct(...) chains must finished wiht `| as(\"name\")`"))
		}
		// copy expected fields from sr.vars to p.fields */
	}
	if debug {
		log.Printf("return:'%+v'", sr)
	}
	return p
}

type PipeAction struct {
	PipeEntry
}

type PipeFirst struct {
	PipeAction
}

func (p *PipeFirst) Compile(tok rune, s *scanner.Scanner) IPipeEntry {
	return p
}

func (p *PipeFirst) ExecWithSelection(r *PipeRuntime, s *goquery.Selection) interface{} {
	return s.First()
}

func (p *PipeFirst) String() string { return "PipeFirst()" }

type PipeStore struct {
	PipeAction
	name string
}

func (p *PipeStore) Compile(tok rune, s *scanner.Scanner) IPipeEntry {
	if s.Scan() != scanner.EOF && "(" == s.TokenText() {
		if s.Scan() != scanner.EOF {
			p.name = s.TokenText()
			p.name = p.name[1 : len(p.name)-1]
			if s.Scan() != scanner.EOF && ")" == s.TokenText() {
				return p
			}
		}
	}
	return nil
}

func (p *PipeStore) String() string {
	return fmt.Sprintf("PipeStore('%s')", p.name)
}

func (p *PipeStore) getStore(rt *PipeRuntime) map[string]interface{} {
	var store map[string]interface{}
	st, ok := rt.vars["...store..."]
	if !ok {
		store = make(map[string]interface{})
		rt.vars["...store..."] = store
	} else {
		store, _ = st.(map[string]interface{})
	}
	return store
}

func (p *PipeStore) ExecWithSelection(r *PipeRuntime, s *goquery.Selection) interface{} {
	store := p.getStore(r)
	store[p.name] = s
	return s
}

type PipeClone struct {
	PipeStore
}

func (p *PipeClone) ExecWithSelection(r *PipeRuntime, s *goquery.Selection) interface{} {
	store := p.getStore(r)
	html, err := s.Html()
	if err == nil {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err == nil {
			store[p.name] = doc.Selection
		}
	}
	return s
}

type PipeRestore struct {
	PipeStore
}

func (p *PipeRestore) String() string {
	return fmt.Sprintf("PipeRestore('%s')", p.name)
}

func (p *PipeRestore) ExecWithSelection(r *PipeRuntime, s *goquery.Selection) interface{} {
	store := p.getStore(r)
	result, ok := store[p.name]
	if ok {
		return result
	} else {
		log.Printf("could not find '%s' in store %+v", p.name, store)
		panic("oops")
	}
	return s
}

type PipeScore struct {
	PipeAction
	score int
}

func (p *PipeScore) Compile(tok rune, s *scanner.Scanner) IPipeEntry {
	if s.Next() != scanner.EOF && "(" == s.TokenText() {
		if s.Next() != scanner.EOF {
			intval, err := strconv.Atoi(s.TokenText())
			if err == nil {
				p.score = intval
			}
			if s.Next() != scanner.EOF && ")" == s.TokenText() {
				return p
			}
		}
	}
	return nil
}

func (p *PipeScore) ExecWithSelection(r *PipeRuntime, s *goquery.Selection) interface{} {
	score, _ := s.Attr("gravityScore")
	scoreValue := 0
	if score != "" {
		scoreInt, err := strconv.Atoi(score)
		if err == nil {
			scoreValue = scoreInt
		}
	}
	setAttr(s, "gravityScore", fmt.Sprintf("%d", scoreValue+p.score))
	return s
}

type PipeRemove struct {
	PipeAction
}

func (p *PipeRemove) Compile(tok rune, s *scanner.Scanner) IPipeEntry {
	return p
}

func (p *PipeRemove) ExecWithSelection(r *PipeRuntime, s *goquery.Selection) interface{} {
	s.Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})
	return s
}

type PipeText struct {
	PipeAction
}

func (p *PipeText) Compile(tok rune, s *scanner.Scanner) IPipeEntry {
	return p
}

func (p *PipeText) ExecWithSelection(r *PipeRuntime, s *goquery.Selection) interface{} {
	return s.Text()
}

func (p *PipeText) String() string { return "PipeText()" }

type PipeAttr struct {
	PipeAction
	name  string
	value string
}

func (p *PipeAttr) Compile(tok rune, s *scanner.Scanner) IPipeEntry {
	for s.Scan() != scanner.EOF && "(" != s.TokenText() {
	}

	if "(" == s.TokenText() {
		for s.Scan() != scanner.String {
		}
		p.name = s.TokenText()
		p.name = p.name[1 : len(p.name)-1]
		for s.TokenText() != ")" && s.Scan() != scanner.EOF {
		}
	}
	return p
}

func (p *PipeAttr) ExecWithSelection(r *PipeRuntime, s *goquery.Selection) interface{} {
	value, ok := s.Attr(p.name)
	if ok {
		p.value = value
	}
	return p.value
}

func (p *PipeAttr) String() string { return fmt.Sprintf("PipeAttr(\"%s\")", p.name) }

type PipeUnhumanDate struct {
	PipeAction
}

func (p *PipeUnhumanDate) ExecWithSelection(r *PipeRuntime, s *goquery.Selection) interface{} {
	return s.Text()
}

func (p *PipeUnhumanDate) ExecWithString(r *PipeRuntime, str string) interface{} {
	return str
}

type PipeAs struct {
	PipeAction
	name  string
	value interface{}
}

func (p *PipeAs) ExecWithSelection(r *PipeRuntime, s *goquery.Selection) interface{} {
	p.value = s
	r.vars[p.name] = p.value
	return p
}

func (p *PipeAs) ExecWithString(r *PipeRuntime, str string) interface{} {
	p.value = str
	r.vars[p.name] = p.value
	return p
}

func (p *PipeAs) Compile(tok rune, s *scanner.Scanner) IPipeEntry {
	for s.Scan() != scanner.EOF && s.TokenText() != "(" {
	}
	for s.Scan() != scanner.String {
	}
	p.name = s.TokenText()
	p.name = p.name[1 : len(p.name)-1]
	for s.Scan() != scanner.EOF && s.TokenText() != ")" {
	}
	return p
}

func (p *PipeAs) String() string { return fmt.Sprintf("PipeAs('%s'->'%+v')", p.name, p.value) }
func (p *PipeAs) WrapStruct(runtime *PipeRuntime, st *PipeStruct, s *goquery.Selection) interface{} {
	if _, ok := runtime.vars[p.name]; ok {
		log.Printf("WARNING: field '%s' ALREADY EXISTS", p.name)
	}
	runtime.vars[p.name] = st.fields
	return runtime.String()
}

type PipeFrom struct {
	PipeAction
	name  string
	value interface{}
}

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
	if pipe == "" {
		return nil
	}
	var result = make([]IPipeEntry, 0, 5)
	var s scanner.Scanner
	s.Init(strings.NewReader(pipe))
	var tok rune
	for tok != scanner.EOF {
		tok = s.Scan()
		entry := compilerFactory(tok, &s)
		if entry != nil {
			result = append(result, entry)
			tok = s.Scan() // shift out from last ')'
		}
		if "|" == s.TokenText() {
			continue
		} else if ")" == s.TokenText() {
			return result
		}
	}
	return result
}

func compilerFactory(tok rune, s *scanner.Scanner) IPipeEntry {
	if "find" == s.TokenText() {
		return compileAndReturn(&PipeFind{}, tok, s)
	} else if "text" == s.TokenText() {
		return compileAndReturn(&PipeText{}, tok, s)
	} else if "first" == s.TokenText() {
		return compileAndReturn(&PipeFirst{}, tok, s)
	} else if "struct" == s.TokenText() {
		return compileAndReturn(&PipeStruct{}, tok, s)
	} else if "as" == s.TokenText() {
		return compileAndReturn(&PipeAs{}, tok, s)
		//return (&PipeAs{}).Compile(tok, s).(*PipeEntry)
	} else if "unhumanPublishDate" == s.TokenText() {
		return compileAndReturn(&PipeUnhumanDate{}, tok, s)
	} else if "remove" == s.TokenText() {
		return compileAndReturn(&PipeRemove{}, tok, s)
	} else if "attr" == s.TokenText() {
		return compileAndReturn(&PipeAttr{}, tok, s)
	} else if "store" == s.TokenText() {
		return compileAndReturn(&PipeStore{}, tok, s)
	} else if "clone" == s.TokenText() {
		return compileAndReturn(&PipeClone{}, tok, s)
	} else if "restore" == s.TokenText() {
		return compileAndReturn(&PipeRestore{}, tok, s)
	} else if "|" == s.TokenText() {

	} else if ")" == s.TokenText() {
		return nil
	}
	return nil
}

func compileAndReturn(object IPipeEntry, tok rune, s *scanner.Scanner) IPipeEntry {
	object.Compile(tok, s)
	return object
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
		if sel, ok := result.(*goquery.Selection); ok {
			md.content = sel
		}
	}
	return md
}

func (r *Rules) execChain(rt *PipeRuntime, s *goquery.Selection, chain []IPipeEntry) interface{} {
	var result interface{} = s
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
		if sel, ok := result.(*goquery.Selection); ok {
			return sel
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
