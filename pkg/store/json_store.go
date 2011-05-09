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
	"flag"
	"fmt"
	"github.com/stevela/lwb/lwb"
	"io/ioutil"
	"json"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

var flagJsonPath *string = flag.String("json_dir", "json_store", "Path to the json store")

const (
	postSuffix       = ".post"
	pageSuffix       = ".page"
	timeFormat       = "Mon Jan 2 15:04:05 MST 2006"
	numMonthsPerYear = 12
)

var monthNames = [...]string{"January", "February", "March", "April",
	"May", "June", "July", "August", "September", "October", "November", "December"}

type monthlyPosts struct {
	byMonth [numMonthsPerYear][]*Post
}

type jsonStore struct {
	// The pages in the store.
	pages map[string]*Post

	// The posts in the store.
	posts []*Post

	// A map of path -> post.
	postsByPath map[string]*Post

	// A map of year -> array of posts by month.
	postsByYear map[int]*monthlyPosts

	// A map of tag -> array of posts.
	postsByTag map[string][]*Post

	// A map of category -> array of posts.
	postsByCategory map[string][]*Post
}

// sort.Interface
func (js *jsonStore) Len() int {
	return len(js.posts)
}

func (js *jsonStore) Less(i, j int) bool {
	// Reverse sort.
	return js.posts[i].Published.Seconds() > js.posts[j].Published.Seconds()
}

func (js *jsonStore) Swap(i, j int) {
	tmp := js.posts[i]
	js.posts[i] = js.posts[j]
	js.posts[j] = tmp
}

// NewJsonStore creates a new store. If not nil, postLoadHook can be used to make modifications to
// the post after it has been loaded.
func NewJsonStore(config *lwb.BlogConfig, postLoadHook func(*Post)) (js *jsonStore, err os.Error) {
	// Load all the posts and pages.
	fileInfos, err := ioutil.ReadDir(*flagJsonPath)
	if err != nil {
		panic("Failed to scan for posts: " + err.String())
	}

	js = new(jsonStore)

	// Build the store.
	js = &jsonStore{
		pages:           make(map[string]*Post),
		postsByPath:     make(map[string]*Post),
		postsByYear:     make(map[int]*monthlyPosts),
		postsByTag:      make(map[string][]*Post),
		postsByCategory: make(map[string][]*Post),
	}

	for _, fileInfo := range fileInfos {
		if !strings.HasSuffix(fileInfo.Name, postSuffix) &&
			!strings.HasSuffix(fileInfo.Name, pageSuffix) {
			continue
		}

		data, err := ioutil.ReadFile(path.Join(*flagJsonPath, fileInfo.Name))
		if err != nil {
			panic("Failed to read file: " + err.String())
		}

		item := new(Post)
		if err = json.Unmarshal(data, item); err != nil {
			panic("Failed to parse item: " + err.String())
		}

		if !item.IsPublished() {
			continue
		}

		if len(item.Body) == 0 {
			// Try loading from external page.
			ext := path.Ext(fileInfo.Name)
			base := fileInfo.Name[0 : len(fileInfo.Name)-len(ext)]
			data, err := ioutil.ReadFile(path.Join(*flagJsonPath, fmt.Sprintf("%s.body", base)))
			if err == nil {
				item.Body = string(data)
			}
		}

		if len(item.Body) == 0 {
			panic("No body in post for " + fileInfo.Name)
		}

		// Convert times.
		if item.LastModified, err = time.Parse(timeFormat, item.LastModifiedDate); err != nil {
			panic("Failed to parse last modified time: " + err.String())
		}
		if item.Published, err = time.Parse(timeFormat, item.PublishedDate); err != nil {
			panic("Failed to parse published time: " + err.String())
		}

		// Type.
		switch item.Type {
		case "post":
			item.Path = fmt.Sprintf("/%d/%.02d/%s",
				item.Published.Year, item.Published.Month, item.Basename)

			js.posts = append(js.posts, item)
			js.postsByPath[item.Path] = item
		case "page":
			item.Path = fmt.Sprintf("/page/%s", item.Basename)
			js.pages[item.Path] = item
		default:
			panic("Unknown item type: " + item.Type)
		}

		item.CanonicalBlogUrl = config.BlogUrl
		item.CanonicalPath = item.Path

		// Run hook.
		if postLoadHook != nil {
			postLoadHook(item)
		}
	}

	sort.Sort(js)

	for i := 0; i < len(js.posts); i += 1 {
		// Navigation.
		post := js.posts[i]
		if i != 0 {
			post.PreviousPath = js.posts[i-1].Path
			post.PreviousTitle = js.posts[i-1].Title
		}

		if i != len(js.posts)-1 {
			post.NextPath = js.posts[i+1].Path
			post.NextTitle = js.posts[i+1].Title
		}

		// Date based archives.
		year := int(post.Published.Year)
		zeroBasedMonth := int(post.Published.Month) - 1

		monthPosts, foundYear := js.postsByYear[year]
		if !foundYear {
			js.postsByYear[year] = new(monthlyPosts)
			monthPosts = js.postsByYear[year]
		}
		monthPosts.byMonth[zeroBasedMonth] = append(monthPosts.byMonth[zeroBasedMonth], post)

		// Other archives.
		for _, tag := range post.Tags {
			js.postsByTag[tag] = append(js.postsByTag[tag], post)
		}

		for _, category := range post.Categories {
			js.postsByCategory[category] = append(js.postsByCategory[category], post)
		}
	}

	return
}

// GetRecentPosts returns the most recent numPosts posts.
func (js *jsonStore) GetRecentPosts(numPosts int) (posts []*Post) {
	if numPosts > len(js.posts) {
		numPosts = len(js.posts)
	}

	posts = js.posts[:numPosts]

	return
}

// GetPage returns a page with the given name.
func (js *jsonStore) GetPage(name string) (post *Post, found bool) {
	post, found = js.pages[name]

	return
}

// GetPostsByPath returns a post given a date based path.
func (js *jsonStore) GetPostByPath(path string) (post *Post, found bool) {
	post, found = js.postsByPath[path]

	return
}

// GetPostsByYear returns all the posts for a given year.
func (js *jsonStore) GetPostsByYear(year int) (posts []*Post, found bool) {
	if monthPosts, foundPosts := js.postsByYear[year]; foundPosts {
		for month := numMonthsPerYear - 1; month >= 0; month -= 1 {
			for _, post := range monthPosts.byMonth[month] {
				posts = append(posts, post)
			}
		}
	}

	found = len(posts) > 0

	return
}

// GetPostsByYearMonth returns all the posts for a given year and month.
func (js *jsonStore) GetPostsByYearMonth(year, month int) (posts []*Post, found bool) {
	if month < 1 || month > numMonthsPerYear {
		found = false
	} else {
		if monthPosts, foundPosts := js.postsByYear[year]; foundPosts {
			posts = monthPosts.byMonth[month-1]
		}

		found = len(posts) != 0
	}

	return
}

// GetPostsByTag returns all the posts for a given tag.
func (js *jsonStore) GetPostsByTag(tag string) (posts []*Post, found bool) {
	posts, found = js.postsByTag[tag]

	return
}

// GetPostsByCategory returns all the posts for a given category.
func (js *jsonStore) GetPostsByCategory(category string) (posts []*Post, found bool) {
	posts, found = js.postsByCategory[category]

	return
}

// GetTags returns all the tags.
func (js *jsonStore) GetTags() (tags []string) {
	for tag, _ := range js.postsByTag {
		tags = append(tags, tag)
	}

	return
}

// GetCategories returns all the categories.
func (js *jsonStore) GetCategories() (categories []string) {
	for category, _ := range js.postsByCategory {
		categories = append(categories, category)
	}

	return
}

// GetArchives returns the yearly archives.
func (js *jsonStore) GetArchives() (archives Archives) {
	for year, posts := range js.postsByYear {
		for month := 0; month < numMonthsPerYear; month += 1 {
			if len(posts.byMonth[month]) != 0 {
				archives = append(archives, Archive{year, month,
					fmt.Sprintf("%s %d", monthNames[month], year),
					fmt.Sprintf("/%d/%02d/", year, month+1)})
			}
		}
	}

	sort.Sort(archives)

	return
}
