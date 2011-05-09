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

package store

import (
	"http"
	"time"
)

type Archive struct {
	Year        int
	Month       int
	Description string
	Path        string
}

type Archives []Archive

// sort.Interface
func (a Archives) Len() int {
	return len(a)
}

func (a Archives) Less(i, j int) bool {
	// Reverse sort.
	if a[i].Year > a[j].Year {
		return true
	}

	if a[i].Year < a[j].Year {
		return false
	}

	return a[i].Month > a[j].Month
}

func (a Archives) Swap(i, j int) {
	tmp := a[i]
	a[i] = a[j]
	a[j] = tmp
}


// Store represents an interface to the underlying storage.
type Store interface {
	GetRecentPosts(numPosts int) []*Post
	GetPage(name string) (*Post, bool)
	GetPostByPath(path string) (*Post, bool)
	GetPostsByYear(year int) ([]*Post, bool)
	GetPostsByYearMonth(year, month int) ([]*Post, bool)
	GetPostsByTag(tag string) ([]*Post, bool)
	GetPostsByCategory(tag string) ([]*Post, bool)
	GetTags() []string
	GetCategories() []string
	GetArchives() Archives
}

// Post represents a post in the system.
type Post struct {
	// The title of the post.
	Title string

	// The unprocessed content of the post.
	Body string

	// The base name of the post.
	Basename string

	// The format of the post ("none", "convertbreaks" or "textile").
	Format string

	// The status of the post ("publish" or "draft").
	Status string

	// The type of the post ("post" or "page").
	Type string

	// The string uuid of the post.
	Uuid string

	// The tags.
	Tags []string

	// The categories.
	Categories []string

	// For back compat, whether this post came from the old blog.
	IsOldEntry bool

	// Allow comments on pages.
	CommentOnPage bool

	// The following are computed...

	// The path of the post from the root of the archives directory.
	Path string

	// The path of the previous post.
	PreviousPath string

	// The title of the previous post.
	PreviousTitle string

	// The path of the next post.
	NextPath string

	// The title of the next post.
	NextTitle string

	// Dates.
	LastModified *time.Time
	Published    *time.Time

	LastModifiedDate string
	PublishedDate    string

	// The path to use when talking to external sites.
	CanonicalPath string

	// The blog url to use when talking to external sites.
	CanonicalBlogUrl *http.URL

	// Cached data.
	CachedPost             []byte
	CachedPostWithFeedback []byte
}

// IsFormatTextile returns whether the post should be formatted using the TextileFormatter.
func (p *Post) IsFormatTextile() bool {
	return p.Format == "textile"
}

// IsFormatConvertBreaks returns whether the post should be formatted using the ConvertBreaksFormatter.
func (p *Post) IsFormatConvertBreaks() bool {
	return p.Format == "convertbreaks"
}

// IsPublished returns whether the post has been published.
func (p *Post) IsPublished() bool {
	return p.Status == "publish"
}

// IsPost returns whether the post is a post..
func (p *Post) IsPost() bool {
	return p.Type == "post"
}

// IsPage returns whether the post is a page..
func (p *Post) IsPage() bool {
	return p.Type == "page"
}

// PublishedShort returns a date string of the form "Sunday, January 2 2006".
func (p *Post) PublishedShort() string {
	return p.Published.Format("Monday, January 2 2006")
}

// PublishedRFC1123 returns a date string in RFC1123 format.
func (p *Post) PublishedRFC1123() string {
	return p.Published.Format(time.RFC1123)
}
