/*
Copyright 2011 Steve Lacey

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package handlers

import (
	"flag"
	"github.com/garyburd/twister/web"
	"github.com/stevela/lwb/lwb"
	"github.com/stevela/lwb/store"
	"github.com/stevela/lwb/textile"
	"io/ioutil"
	"path"
	"strings"
	"template"
)

var flagTemplatePath *string = flag.String("tmpl", "tmpl", "Path to the templates")

type RenderContext struct {
	Db          store.Store
	Config      *lwb.BlogConfig
	Generator   string
	RecentPosts []*store.Post
	Tags        []string
	Categories  []string
	Archives    store.Archives
	UseCache    bool
	Title       string
	Path        string
}

type templateEntry struct {
	Path      string
	Timestamp int64
	*template.Template
}

var templates = make(map[string]*templateEntry)

const templateSuffix = ".tmpl"

func ReloadTemplates(config *lwb.BlogConfig) {
	fileInfos, err := ioutil.ReadDir(*flagTemplatePath)
	if err != nil {
		panic("Failed to scan for templates: " + err.String())
	}

	for _, fileInfo := range fileInfos {
		if !strings.HasSuffix(fileInfo.Name, templateSuffix) {
			continue
		}
		basename := fileInfo.Name[:len(fileInfo.Name)-len(templateSuffix)]
		_, present := templates[basename]
		if !present || templates[basename].Timestamp < fileInfo.Mtime_ns {
			path := path.Join(*flagTemplatePath, fileInfo.Name)
			tmpl := template.New(
				template.FormatterMap{
					"textile":          textile.TextileFormatter,
					"textileFullLinks": textile.GetTextileFullLinkFormatter(config.BlogUrl.String()),
					"entities":         textile.EncodeEntitiesFormatter,
					"spaces":           lwb.EncodeSpacesFormatter,
					"convertbreaks":    lwb.ConvertBreaksFormatter,
				})
			tmpl.SetDelims("{{", "}}")

			b, err := ioutil.ReadFile(path)
			if err != nil {
				panic("failed to read " + path + ": " + err.String())
			}
			err = tmpl.Parse(string(b))
			if err != nil {
				panic("failed to parse " + path + ": " + err.String())
			}

			templates[basename] = &templateEntry{
				Path:      path,
				Timestamp: fileInfo.Mtime_ns,
				Template:  tmpl,
			}
		}
	}
}

// DebugFilter does various things (like reloading templates on each request if
// in debug mode).
func DebugFilter(enabled bool, config *lwb.BlogConfig, handler web.Handler) web.Handler {
	if !enabled {
		return handler
	}

	return web.HandlerFunc(func(req *web.Request) {
		ReloadTemplates(config)
		handler.ServeWeb(req)
	})
}
