package dictionary

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/html"
)

// PageDictionary is the output format of a page dictionary
type PageDictionary struct {
	Generated    time.Time             `json:"_generated"`
	ProcessingMs int64                 `json:"_processing_ms"`
	Routes       []pageDictionaryEntry `json:"routes"`
}

// pageDictionaryTemplate is the config file of a page dictionary
type pageDictionaryTemplate struct {
	Path     string                   `json:"path"`
	File     string                   `json:"file"`
	Children []pageDictionaryTemplate `json:"children"`
}

// pageDictionaryEntry is an entry of a page dictionary
type pageDictionaryEntry struct {
	Path      string                  `json:"path"`
	Component pageDictionaryComponent `json:"component"`
	Children  []pageDictionaryEntry   `json:"children,omitempty"`
}

// pageDictionaryComponent is a component which contains a template
type pageDictionaryComponent struct {
	Template string `json:"template"`
}

// expiredPage walks through children to find pages that recently updated
func expiredPage(expiredAt time.Time, template pageDictionaryTemplate) bool {
	for _, child := range template.Children {
		if expiredPage(expiredAt, child) {
			return true
		}
	}
	info, err := os.Stat(template.File)
	if err != nil {
		return false
	}
	expired := info.ModTime().Sub(expiredAt)
	if expired > 0 {
		println(template.File, "updated", expired.String(), "ago")
	}
	return expired > 0
}

// GeneratePageDictionary converts a file dictionary into an appropriate vue-router definition.
// Using a file dictionary is useful as it allows you to point where files are and this function
// will read all these files and pass them accordingly in a format that is acceptable.
// [ { "path": "/", "file": "web/static/dashboard/index.html" } ]
// Would read from the file and return in a content similar to
// [ { "path": "/", "component": { "template": "filecontents..." } } ]
// It also supports children. It will also check file modification times to not create a new
// dictionary when nothing has changed and will instead use the cached dictionary. The page
// dictionary also contains the ms taken to generate the file and the time it was made at (used
// internally to track file changes) in the format
// { "_generated": 0, "_processing_ms": 1, "routes": [ ... ] }
func GeneratePageDictionary(dictionaryPath string, outputPath string) (body *PageDictionary, err error) {
	now := time.Now().UTC()

	file, err := ioutil.ReadFile(dictionaryPath)
	if err != nil {
		return
	}
	inputDictionary := make([]pageDictionaryTemplate, 0)
	if err = json.Unmarshal(file, &inputDictionary); err != nil {
		return
	}

	parent := pageDictionaryTemplate{
		Path:     "ROOT",
		Children: inputDictionary,
	}

	if ofile, err := ioutil.ReadFile(outputPath); err == nil {
		dictionary := &PageDictionary{}
		if err = json.Unmarshal(ofile, &dictionary); err == nil {
			lastUpdate := dictionary.Generated
			expired := expiredPage(lastUpdate, parent)
			if !expired {
				return dictionary, nil
			}
		}
	}

	routes, _ := walkTemplate(parent, true)
	dict := &PageDictionary{
		Generated: now,
		Routes:    routes.Children,
	}

	dict.ProcessingMs = time.Now().UTC().Sub(now).Milliseconds()
	if body, err := json.Marshal(dict); err == nil {
		err := ioutil.WriteFile(outputPath, body, 0644)
		if err != nil {
			println(err.Error())
		}
	}

	return dict, nil
}

// walkTemplate walks through children
func walkTemplate(template pageDictionaryTemplate, skip bool) (pageDictionaryEntry, bool) {
	entry := pageDictionaryEntry{
		Path: template.Path,
	}

	if !skip {
		println("Opening", template.File)
		body, err := ioutil.ReadFile(template.File)
		if err != nil {
			println("Could not open", template.File, err.Error())
			return entry, false
		}

		m := minify.New()
		m.AddFunc("text/html", html.Minify)
		minbody, err := m.Bytes("text/html", body)

		if err != nil {
			entry.Component = pageDictionaryComponent{string(body)}
		} else {
			entry.Component = pageDictionaryComponent{string(minbody)}
		}
	}

	children := make([]pageDictionaryEntry, 0)
	for _, child := range template.Children {
		if childEntry, ok := walkTemplate(child, false); ok {
			children = append(children, childEntry)
		}
	}
	entry.Children = children
	return entry, true
}
