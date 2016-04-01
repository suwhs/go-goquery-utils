package pipes

// import (
// 	"encoding/json"
// 	"io/ioutil"
// 	"log"
// 	"strings"
// )

// /**
//     configuration holds hosts parsing rules
//     map[host]Rules

//     global map

//     requests to configuration:
//         get(host,func(Rules))
// **/

// type PipesConfig struct {
// 	cfguChannel chan interface{}
// 	cfguStop    chan int
// 	hosts       map[string]*Rules
// 	rules       []*Rules
// 	aliases     map[string]string
// 	path        string
// }

// type RulesApplyFunc func(rules *Rules)
// type RulesGetFunc func(rules []*RulesScript)
// type RulesSetFunc func()
// type RulesListFunc func(jstr []string)

// type loadQuery struct {
// 	host string
// 	fn   RulesApplyFunc
// 	resp chan interface{}
// }

// type getQuery struct {
// 	selector string
// 	fn       RulesGetFunc
// 	resp     chan interface{}
// }

// type setQuery struct {
// 	rule *RulesScript
// 	fn   RulesSetFunc
// }

// type listQuery struct {
// 	fn RulesListFunc
// }

// func (pc *PipesConfig) ApplyRules(host string, fn RulesApplyFunc) chan interface{} {
// 	response := make(chan interface{})
// 	pc.cfguChannel <- &loadQuery{host: host, fn: fn, resp: response}
// 	return response
// }

// func (pc *PipesConfig) GetRules(selector string, fn RulesGetFunc) chan interface{} {
// 	response := make(chan interface{})
// 	pc.cfguChannel <- &getQuery{selector: selector, fn: fn, resp: response}
// 	return response
// }

// func (pc *PipesConfig) SetRules(rule *SingleRuleJSON, fn RulesSetFunc) {
// 	pc.cfguChannel <- &setQuery{rule: rule, fn: fn}
// }

// func (pc *PipesConfig) ListRules(fn RulesListFunc) {
// 	pc.cfguChannel <- &listQuery{fn: fn}
// }

// func NewPipesConfig(cfgPath string) *PipesConfig {
// 	pc := &PipesConfig{}
// 	pc.path = cfgPath
// 	pc.cfguChannel = make(chan interface{}, 100)
// 	pc.cfguStop = make(chan int, 10)
// 	pc.hosts = make(map[string]*Rules)
// 	pc.aliases = make(map[string]string)
// 	pc.rules = make([]Rules, 0, 100)
// 	go pc.cfguMainLoop()
// 	go pc.cfguLoadRules()
// 	return pc
// }

// func (pc *PipesConfig) DestroyCFGU() {
// 	// log.Printf("destroy pipes config")
// 	pc.cfguStop <- 1
// 	pc.cfguSaveRules()
// }

// func (pc *PipesConfig) cfguMainLoop() {
// 	defer func() {
// 		pc.cfguSafeShutdown()
// 	}()
// 	for {
// 		select {
// 		case query := <-pc.cfguChannel:
// 			pc.cfguHandleQuery(query)
// 		case _ = <-pc.cfguStop:
// 			log.Printf("CFGU shutdown")
// 			return
// 		}
// 	}
// }

// func (pc *PipesConfig) cfguSafeShutdown() {
// 	log.Printf("CFGU shutdowned")
// }

// /*
//    dispatch loop - avoid race with access to rules list
// */

// func (pc *PipesConfig) cfguHandleQuery(query interface{}) {
// 	if lq, ok := query.(*loadQuery); ok {
// 		pc.cfguHandleLoadQuery(lq)
// 	} else if gq, ok := query.(*getQuery); ok {
// 		pc.cfguHandleGetQuery(gq)
// 	} else if sq, ok := query.(*setQuery); ok {
// 		pc.cfguHandleSetQuery(sq)
// 	} else if lq, ok := query.(*listQuery); ok {
// 		pc.cfguHandleListQuery(lq)
// 	}
// }

// func (pc *PipesConfig) cfguHandleLoadQuery(q *loadQuery) {
// 	h := strings.ToLower(q.host)
// 	if alias, ok := pc.aliases[h]; ok {
// 		// log.Printf("replace '%s' -> '%s' ", h, alias)
// 		h = alias
// 	}
// 	if r, ok := pc.hosts[h]; ok {
// 		r.Host = h
// 		go func() {
// 			q.fn(r)
// 			q.resp <- "f"
// 		}()
// 	} else {
// 		go func() { q.fn(nil); q.resp <- "f" }()
// 	}
// }

// func (pc *PipesConfig) cfguHandleGetQuery(q *getQuery) {
// 	result := make([]*SingleRuleJSON, 0, 10)
// 	for k, v := range pc.hosts {
// 		// log.Printf("host:%s", k)
// 		if q.selector == "*" || q.selector == "" || strings.Contains(k, q.selector) {
// 			result = append(result, v.Source)
// 		} else {
// 		}
// 	}
// 	q.fn(result)
// 	q.resp <- "f"
// }

// func (pc *PipesConfig) cfguHandleSetQuery(q *setQuery) {
// 	// compile rule and
// 	a := &Rules{}
// 	a.CompileAll(q.rule)
// 	a.Source = q.rule
// 	var lastHost = ""
// 	// need correct handle replace rule
// 	var replaceTaret *Rules
// 	// var replaceHost string
// 	for i := range q.rule.Host {
// 		h := q.rule.Host[i]
// 		if replaceTaret == nil {
// 			if record, ok := pc.hosts[h]; ok {
// 				// replaceHost = h
// 				replaceTaret = record
// 				// log.Printf("found rule to replace, '%s'", replaceHost)
// 				for k := range replaceTaret.Source.Host {
// 					delete(pc.hosts, replaceTaret.Source.Host[k])
// 				}
// 				for k := range replaceTaret.Source.Aliases {
// 					delete(pc.aliases, replaceTaret.Source.Aliases[k])
// 				}
// 			}
// 		}
// 		pc.hosts[h] = a
// 		lastHost = h
// 	}
// 	if lastHost != "" {
// 		for i := range q.rule.Aliases {
// 			pc.aliases[q.rule.Aliases[i]] = lastHost
// 		}
// 	}
// 	if replaceTaret != nil {
// 		removes := make([]int, 0, 10)
// 		for idx := pc.nextEntry(0, replaceTaret); idx > -1; idx = pc.nextEntry(idx+1, replaceTaret) {
// 			// log.Printf("entry index: '%d'", idx)
// 			removes = append(removes, idx)
// 		}
// 		for i := len(removes) - 1; i > -1; i-- {
// 			if removes[i]+1 < len(pc.rules) {
// 				pc.rules = append(pc.rules[:removes[i]], pc.rules[removes[i]+1])
// 			} else {
// 				pc.rules = pc.rules[:removes[i]]
// 			}
// 			// log.Printf("remove entry '%d'", removes[i])
// 		}
// 		//for i := range replaceTaret.Source.Aliases {
// 		//	pc.rules[i] = pc.rules[len(pc.rules)-1]
// 		//	pc.rules = pc.rules[:len(pc.rules)-1]
// 		// }
// 		*replaceTaret = *a
// 	} else {
// 		pc.rules = append(pc.rules, *a)
// 	}
// 	q.fn()
// }

// func (pc *PipesConfig) nextEntry(start int, r *Rules) int {
// 	return SliceIndex(start, len(pc.rules), func(pos int) bool { return r.Compare(&pc.rules[pos]) == 0 })
// }

// func SliceIndex(start int, limit int, predicate func(i int) bool) int {
// 	for i := start; i < limit; i++ {
// 		if predicate(i) {
// 			return i
// 		}
// 	}
// 	return -1
// }

// func (pc *PipesConfig) cfguHandleListQuery(q *listQuery) {
// 	q.fn([]string{})
// }

// func (r *SingleRuleJSON) load(pc *PipesConfig) {
// 	pc.SetRules(r, func() {})
// }

// type RulesJSON struct {
// 	RulesList   []SingleRuleJSON `json:"rules"`
// 	VersionCode string           `json:"version"`
// }

// func (r *RulesJSON) load(pc *PipesConfig) {
// 	for i := range r.RulesList {
// 		r.RulesList[i].load(pc)
// 	}
// }

// func (pc *PipesConfig) cfguLoadRules() {
// 	file, e := ioutil.ReadFile(pc.path)
// 	if e == nil {
// 		j := &RulesJSON{}
// 		err := json.Unmarshal(file, j)
// 		if err != nil {
// 			log.Printf("Error:%s", err)
// 			return
// 		}
// 		// log.Printf("loaded: %+v", j)
// 		j.load(pc)
// 		// log.Printf("loaded %d rules", len(pc.hosts))
// 		return
// 	}
// 	// log.Printf("file '%s' could not found", pc.path)
// }

// func (pc *PipesConfig) cfguSaveRules() {
// 	// log.Printf("save pipes..")

// 	result := make([]SingleRuleJSON, 0, 10)
// 	for _, v := range pc.hosts {
// 		result = append(result, *v.Source)
// 	}

// 	rj := RulesJSON{RulesList: result}

// 	json, err := json.Marshal(rj)
// 	if err == nil {
// 		err = ioutil.WriteFile(pc.path, json, 0644)
// 		if err == nil {
// 			// log.Printf("rules saved")
// 			return
// 		}
// 	}
// 	log.Printf("error saving rules: %s", err.Error())
// }
