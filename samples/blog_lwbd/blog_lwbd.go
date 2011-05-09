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

package main

import (
	"flag"
	"fmt"
	"github.com/garyburd/twister/expvar"
	"github.com/garyburd/twister/server"
	"github.com/garyburd/twister/web"
	"github.com/stevela/lwb/handlers"
	"github.com/stevela/lwb/lwb"
	"github.com/stevela/lwb/store"
	"http"
	"log"
	"net"
	"os"
	"sort"
)

var flagDebug *bool = flag.Bool("debug", false, "Run in debug mode")
var flagDebugLog *bool = flag.Bool("debuglog", false, "Output debug logs")
var flagCache *bool = flag.Bool("cache", true, "Run with a cache")
var flagGenerator *string = flag.String("generator", "Light Weight Blogging (http://github.com/stevela/lwb)",
	"A link to the software that generated this site")
var flagHost *string = flag.String("host", "example.com", "Host to run this server as")
var flagLog *string = flag.String("log", "access.log", "Path to access.log")
var flagPort *int = flag.Int("port", 8080, "Port to run the server on")
var flagProtocol *string = flag.String("protocol", "http", "Protocol to run this server on")

var config = &lwb.BlogConfig{
	Author:      "Your Name",
	Title:       "Your Blog Title",
	Description: "Your blog description...",

	// Version number for static content.
        Version: 1,

	// Main Index.
	NumMainIndexPosts: 20,
	MainIndexRegexp:   "/",

	// Sidebar.
	NumRecentPosts: 10,

	// Pages
	PageRegexp: "/<path:page/[^/]*>",

	// Posts.
	PostRegexp: "/<year:[0-9][0-9][0-9][0-9]>/<month:[0-9][0-9]>/<basename:[^/]*>",

	// Archives.
	MonthlyArchiveRegexp:  "/<year:[0-9][0-9][0-9][0-9]>/<month:[0-9][0-9]>/",
	YearlyArchiveRegexp:   "/<year:[0-9][0-9][0-9][0-9]>/",
	TagArchiveRegexp:      "/tag/<tag:[^/]*>/",
	CategoryArchiveRegexp: "/category/<tag:[^/]*>/",

	// Other content.
	StaticRegexp: "/<path:.*>",

	// Rss.
	NumRssFeedPosts: 20,
	RssUrl:          "/index.xml", // What to use in rendered content.
	RssFeedRegexp:   "/index.xml", // What to actual serve. Maybe different if using something like feedburner.

	// Comments.
	DisqusShortname: "xxx", // Replace with your own disqus shortname.
}

func pathHandler(req *web.Request, targetPattern string) {
	if newPath := req.Param.Get("path"); newPath == "" {
		req.Error(web.StatusNotFound, os.NewError("Not Found."))
	} else {
		newUrl := fmt.Sprintf(targetPattern, req.URL.Scheme, req.URL.Host, newPath)
		req.Redirect(newUrl, true)
	}
}

func main() {
	flag.Parse()

	var err os.Error
	if config.BlogUrl, err = http.ParseURL(fmt.Sprintf("%s://%s", *flagProtocol, *flagHost)); err != nil {
		panic("Invalid protocol and/or host")
	}

	// Cache.
	if *flagCache {
		config.Cache = lwb.NewCache()
	} else {
		config.Cache = lwb.NewDummyCache()
	}

	// Initialize the database.
	db, _ := store.NewJsonStore(config, nil)

	// Context for rendering.
	tags := db.GetTags()
	categories := db.GetCategories()
	sort.SortStrings(tags)
	sort.SortStrings(categories)

	context := &handlers.RenderContext{
		Db:          db,
		Config:      config,
		Generator:   *flagGenerator,
		RecentPosts: db.GetRecentPosts(config.NumRecentPosts),
		Tags:        tags,
		Categories:  categories,
		Archives:    db.GetArchives(),
		UseCache:    *flagCache,
		Title:       config.Title,
		Path:        config.BlogUrl.String(),
	}

	// Templates...
	handlers.ReloadTemplates(config)

	// Example of adding file extension types for static file serving.
	fileMimeTypes := map[string]string {
	        ".eot": "application/vnd.ms-fontobject",
		".otf": "application/octet-stream",
		".ttf": "application/x-font-ttf",
		".woff": "application/x-font-woff",
	}

	// Expiry for static content.
	const maxAge = 60 * 60 * 24 * 365 * 10
	fileHeaders := web.HeaderMap{
		web.HeaderExpires: {fmt.Sprintf("%d", maxAge)},
		web.HeaderCacheControl: {fmt.Sprintf("max-age=%d", maxAge)},
	}

	serveFileOptions := &web.ServeFileOptions{fileMimeTypes, fileHeaders}

	// Register all path handlers.
	rh := handlers.DebugFilter(*flagDebug, config, web.NewRouter().
		// Stats.
		Register("/expvar", "GET", web.HandlerFunc(expvar.ServeWeb)).

		// Example of redirecting old blog urls to shiny new ones.
		Register("/index.shtml", "GET", web.RedirectHandler("/", true)).
		Register("/index.html", "GET", web.RedirectHandler("/", true)).
		Register("/bionew.shtml", "GET", web.RedirectHandler("/page/bio", true)).
		Register("/blogarchives/<path:.*>.shtml", "GET", func(req *web.Request) { pathHandler(req, "%s://%s/%s") }).
		Register("/blogarchives/<path:[^0-9].*>", "GET", func(req *web.Request) { pathHandler(req, "%s://%s/category/%s") }).
		Register("/blogarchives/<path:[0-9].*>", "GET", func(req *web.Request) { pathHandler(req, "%s://%s/%s") }).

		// Example of redirecting to feedburner.
		//Register("/index.xml", "GET", web.RedirectHandler("http://feeds.feedburner.com/steve-lacey-main", false)).

		// Handlers.
		Register(config.MainIndexRegexp, "GET", handlers.MainIndexHandler(context)).
		Register(config.RssFeedRegexp, "GET", handlers.RssFeedHandler(context)).
		Register(config.MonthlyArchiveRegexp, "GET", handlers.DateArchiveHandler(context)).
		Register(config.YearlyArchiveRegexp, "GET", handlers.DateArchiveHandler(context)).
		Register(config.TagArchiveRegexp, "GET", handlers.TagArchiveHandler(context,
			func(key string) ([]*store.Post, bool) { return db.GetPostsByTag(key) })).
		Register(config.CategoryArchiveRegexp, "GET", handlers.TagArchiveHandler(context,
			func(key string) ([]*store.Post, bool) { return db.GetPostsByCategory(key) })).
		Register(config.PostRegexp, "GET", handlers.SinglePostHandler(context)).
		Register(config.PageRegexp, "GET", handlers.PageHandler(context)).
		Register(config.StaticRegexp, "GET", web.DirectoryHandler("static/", serveFileOptions)))

	// Create a logger.
	logFile, err := os.OpenFile(*flagLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(fmt.Sprintf("Failed to open \"%s\": %s", *flagLog, err.String()))
	}

	defer logFile.Close()
	logger := server.NewApacheCombinedLogger(logFile)

	// Go!
	addr := fmt.Sprintf(":%d", *flagPort)
	fmt.Printf("Running on %s\n", addr)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("Listen", err)
		return
	}

	defer listener.Close()
	err = (&server.Server{Listener: listener, Handler: rh, Logger: logger}).Serve()
	if err != nil {
		log.Fatal("Server", err)
	}
}
