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

package lwb

import (
	"bytes"
	"github.com/garyburd/twister/web"
	"io"
	"os"
)

// PageCache is a simple interface that caches url -> rendered page.
type PageCache interface {
	Run(*web.Request, func(io.Writer) bool)
}

// Cache is an implementation of PageCache.
type Cache struct {
	items map[string][]byte
}

func NewCache() *Cache {
	return &Cache{make(map[string][]byte)}
}

// Run looks up the page in the cache and generates it if it does not exist,
// placing it in the cache afterwards.
func (c *Cache) Run(req *web.Request, fnGenerate func(w io.Writer) bool) {
	cached, found := c.items[req.URL.String()]
	if !found {
		var buf = &bytes.Buffer{}
		if !fnGenerate(buf) {
			req.Error(web.StatusNotFound, os.NewError("Not Found."))
			return
		}

		cached = buf.Bytes()
		c.items[req.URL.String()] = cached
	}

	req.Respond(web.StatusOK, web.HeaderContentType, "text/html").Write(cached)
}

// DummyCache is a noop Cache.
type DummyCache struct{}

func NewDummyCache() *DummyCache {
	return &DummyCache{}
}

// Run simply runs the generator.
func (c *DummyCache) Run(req *web.Request, fnGenerate func(w io.Writer) bool) {
	var buf = &bytes.Buffer{}
	if fnGenerate(buf) {
		req.Respond(web.StatusOK, web.HeaderContentType, "text/html").Write(buf.Bytes())
	} else {
		req.Error(web.StatusNotFound, os.NewError("Not Found."))
	}
}
