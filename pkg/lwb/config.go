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
	"http"
)

type BlogConfig struct {
	BlogUrl     *http.URL
	Author      string
	Title       string
	Description string

	// Version number for versioned resources.
	Version int

	// Main Index.
	NumMainIndexPosts int
	MainIndexRegexp   string

	// Sidebar.
	NumRecentPosts int

	// Pages.
	PageRegexp string

	// Posts.
	PostRegexp string

	// Archives.
	MonthlyArchiveRegexp  string
	YearlyArchiveRegexp   string
	TagArchiveRegexp      string
	CategoryArchiveRegexp string

	// Other content.
	StaticRegexp string

	// Rss.
	NumRssFeedPosts int
	RssUrl          string // What to use in rendered content.
	RssFeedRegexp   string

	// Comments.
	DisqusShortname string

	// Cache.
	Cache PageCache
}
