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

package textile

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"regexp"
	"strings"
	"strconv"
)

var lineRegexp = regexp.MustCompile("[^\\r\\n]*")
var blockQuoteRegexp = regexp.MustCompile("^bq(\\([^)]+\\))?(\\.\\.?) (.*)$")
var preRegexp = regexp.MustCompile("^pre(\\([^)]+\\))?(\\.\\.?) (.*)$")
var codeRegexp = regexp.MustCompile("^bc(\\([^)]+\\))?(\\.\\.?) (.*)$")
var paragraphRegexp = regexp.MustCompile("^p(\\([^)]+\\))?\\. (.*)$")
var headingRegexp = regexp.MustCompile("^h([1-9])(\\([^)]+\\))?\\. (.*)$")
var footnoteRegexp = regexp.MustCompile("^fn([0-9])(\\([^)]+\\))?\\. (.*)$")

// Should change list handling to be able to switch types.
var listRegexp = regexp.MustCompile("^([\\*#]+)(\\([^)]+\\))? (.*)$")

var blockHandlers = [...]struct {
	re *regexp.Regexp
	fn func(*state, []string) bool
}{
	// Match blocks of the form "bq. A blockquote" or "bq.. A blockquote".
	{blockQuoteRegexp, func(st *state, args []string) bool {
		st.checkBlockQuote()
		st.checkPre()
		st.checkCode()
		st.inBlockQuote = true
		st.inLongBlockQuote = args[2] == ".."
		st.currentClassPart = ""
		if args[1] != "" {
			st.currentClassPart = " class=" + args[1][1:len(args[1])-1]
		}
		fmt.Fprintf(st.w, "<blockquote%s><p%s>%s", st.currentClassPart, st.currentClassPart, st.transformLine(args[3]))
		if st.inLongBlockQuote {
			st.w.Write([]byte("</p>"))
		}

		return true
	}},

	// Match blocks of the form "pre. preformatted" or "pre.. preformatted ".
	{preRegexp, func(st *state, args []string) bool {
		st.checkBlockQuote()
		st.checkPre()
		st.checkCode()
		st.inPre = true
		st.inLongPre = args[2] == ".."
		st.currentClassPart = ""
		if args[1] != "" {
			st.currentClassPart = " class=" + args[1][1:len(args[1])-1]
		}
		fmt.Fprintf(st.w, "<pre%s>%s\n", st.currentClassPart, EntityEscapeString(args[3], false, true))

		return true
	}},

	// Match blocks of the form "bc. code" or "bc.. code ".
	{codeRegexp, func(st *state, args []string) bool {
		st.checkBlockQuote()
		st.checkPre()
		st.checkCode()
		st.inCode = true
		st.inLongCode = args[2] == ".."
		st.currentClassPart = ""
		if args[1] != "" {
			st.currentClassPart = " class=" + args[1][1:len(args[1])-1]
		}
		fmt.Fprintf(st.w, "<pre%s><code%s>%s\n", st.currentClassPart, st.currentClassPart, EntityEscapeString(args[3], true, false))

		return true
	}},

	// Match blocks of the form "p. A paragraph".
	{paragraphRegexp, func(st *state, args []string) bool {
		st.checkBlockQuote()
		st.checkPre()
		st.checkCode()
		st.currentClassPart = ""
		if args[1] != "" {
			st.currentClassPart = " class=" + args[1][1:len(args[1])-1]
		}
		fmt.Fprintf(st.w, "<p%s>%s</p>", st.currentClassPart, st.transformLine(args[2]))

		return true
	}},

	// Match blocks of the form "h1. A heading".
	{headingRegexp, func(st *state, args []string) bool {
		st.checkBlockQuote()
		st.checkPre()
		st.checkCode()
		level := args[1]
		st.currentClassPart = ""
		if args[2] != "" {
			st.currentClassPart = " class=" + args[2][1:len(args[2])-1]
		}
		fmt.Fprintf(st.w, "<h%s%s>%s</h%s>", level, st.currentClassPart, st.transformLine(args[3]), level)

		return true
	}},

	// Match blocks of the form "fn1. A footnore".
	{footnoteRegexp, func(st *state, args []string) bool {
		st.checkBlockQuote()
		st.checkPre()
		st.checkCode()
		st.currentClassPart = ""
		if args[2] != "" {
			st.currentClassPart = " class=" + args[2][1:len(args[2])-1]
		}
		fmt.Fprintf(st.w, "<p%s id=fn%s-%s><sup%s>%s</sup> %s&#160;<a href=#fnr%s-%s title=\"Jump back to footnote %s\">&#8617;</a></p>",
			st.currentClassPart, args[1], st.hash, st.currentClassPart, args[1], st.transformLine(args[3]), args[1], st.hash, args[1])
		return true
	}},

	// Match blocks of the form "* item" or "# item".
	{listRegexp, func(st *state, args []string) bool {
		st.checkBlockQuote()
		st.checkPre()

		if st.inCode {
			// Allow lines starting with "*" or "#" in code.
			return false
		}
		st.checkCode()
		st.currentClassPart = ""
		if args[1][0] == '*' {
			st.listType = "ul"
		} else {
			st.listType = "ol"
		}
		if args[2] != "" {
			st.currentClassPart = " class=" + args[2][1:len(args[2])-1]
		}

		level := len(args[1])
		if level < st.currentListLevel {
			for ; level < st.currentListLevel; st.currentListLevel -= 1 {
				fmt.Fprintf(st.w, "</li></%s>", st.listType)
			}
			fmt.Fprintf(st.w, "</li><li>%s", st.transformLine(args[3]))
		} else if level > st.currentListLevel {
			fmt.Fprintf(st.w, "<%s%s><li>%s", st.listType, st.currentClassPart, st.transformLine(args[3]))
		} else {
			fmt.Fprintf(st.w, "</li><li>%s", st.transformLine(args[3]))
		}
		st.currentListLevel = level

		return true
	}},
}

var emphasisRegexp = regexp.MustCompile("^(.*)_([^_]*)_(.*)$")
var strongRegexp = regexp.MustCompile("^(.*)\\*([^*]*)\\*(.*)$")
var subRegexp = regexp.MustCompile("^(.*)~([^~]*)~(.*)$")
var supRegexp = regexp.MustCompile("^(.*)\\^([^\\^]*)\\^(.*)$")

var regexpTags = [...]struct {
	tagRegexp *regexp.Regexp
	tagValue  string
}{
	{emphasisRegexp, "em"},
	{strongRegexp, "strong"},
	{supRegexp, "sup"},
	{subRegexp, "sub"},
}

var urlRegexpStr = "[A-Za-z0-9\\-._~:/?#\\(\\)@!\\$&*+,;=%]+[^,!.' \"\\)\\]]"           // Lame.
var urlCanEndInParenRegexpStr = "[A-Za-z0-9\\-._~:/?#\\(\\)@!\\$&*+,;=%]+[^,!.' \"\\]]" // Lame.
var urlNoParensRegexpStr = "[A-Za-z0-9\\-._~:/?#@!\\$&*+,;=%]+[^,!.' \"\\)\\]]"         // Lame.
var imgRegexp = regexp.MustCompile(fmt.Sprintf("^(.*)!(%s)!(.*)$", urlNoParensRegexpStr))
var altImgRegexp = regexp.MustCompile(fmt.Sprintf("^(.*)!(%s) ?\\(([^)]*)\\)!(.*)$", urlNoParensRegexpStr))
var linkRegexp = regexp.MustCompile(fmt.Sprintf("^(.*)\"([^\"]*)\":(%s)(.*)$", urlRegexpStr))
var titleLinkRegexp = regexp.MustCompile(fmt.Sprintf("^(.*)\"([^\"]*)\\(([^\"]*)\\)\":(%s)(.*)$", urlRegexpStr))
var quoteLinkRegexp = regexp.MustCompile(fmt.Sprintf("^(.*)\\[\"([^\"]*)\":(%s)\\](.*)$", urlCanEndInParenRegexpStr))
var doubleQuoteRegexp = regexp.MustCompile("^(.*)\"([^\"]*)\"(.*)$")
var footnoteRefRegexp = regexp.MustCompile("^(.*)\\[([0-9]*)\\](.*)$")
var bareTagRegexp = regexp.MustCompile("^(.*)(<[^>]*>)(.*)$")

type state struct {
	ch               chan string
	w                io.Writer
	inBlockQuote     bool
	inLongBlockQuote bool
	inPre            bool
	inLongPre        bool
	inCode           bool
	inLongCode       bool
	listType         string
	currentListLevel int
	currentClassPart string
	root_url         string
	hash             string
}

var replacements = [...]struct {
	from string
	to   string
}{
	{"_", "%%%%"},
}

func (st *state) maybeAddRoot(url string) string {
	if len(url) > 0 && url[0] == '/' {
		url = st.root_url + url
	}

	return url
}

func escapeParts(s string) string {
	s = EntityEscapeString(s, false, false)
	for _, r := range replacements {
		s = strings.Replace(s, r.from, r.to, -1)
	}

	return s
}

func unescapeParts(s string) string {
	for _, r := range replacements {
		s = strings.Replace(s, r.to, r.from, -1)
	}

	return s
}

func (st *state) transformLine(s string) string {
	return unescapeParts(st.transformLine2(s))
}

func (st *state) transformLine2(s string) string {
	var subexprs [][]string

	subexprs = bareTagRegexp.FindAllStringSubmatch(s, -1)
	for _, subexpr := range subexprs {
		s = fmt.Sprintf("%s%s%s",
			st.transformLine2(subexpr[1]),
			subexpr[2],
			st.transformLine2(subexpr[3]))
	}
	if len(subexprs) > 0 {
		return s
	}

	// ["foo":http://bar]
	subexprs = quoteLinkRegexp.FindAllStringSubmatch(s, -1)
	for _, subexpr := range subexprs {
		s = fmt.Sprintf("%s<a href=\"%s\">%s</a>%s",
			st.transformLine2(subexpr[1]),
			escapeParts(st.maybeAddRoot(subexpr[3])),
			st.transformLine2(subexpr[2]),
			st.transformLine2(subexpr[4]))
	}

	// "foo(title)":http://bar
	subexprs = titleLinkRegexp.FindAllStringSubmatch(s, -1)
	for _, subexpr := range subexprs {
		s = fmt.Sprintf("%s<a href=\"%s\" title=\"%s\">%s</a>%s",
			st.transformLine2(subexpr[1]),
			escapeParts(st.maybeAddRoot(subexpr[4])),
			st.transformLine2(subexpr[3]),
			st.transformLine2(subexpr[2]),
			st.transformLine2(subexpr[5]))
	}

	// "foo":http://bar
	subexprs = linkRegexp.FindAllStringSubmatch(s, -1)
	for _, subexpr := range subexprs {
		s = fmt.Sprintf("%s<a href=\"%s\">%s</a>%s",
			st.transformLine2(subexpr[1]),
			escapeParts(st.maybeAddRoot(subexpr[3])),
			st.transformLine2(subexpr[2]),
			st.transformLine2(subexpr[4]))
	}

	// !http://image!
	subexprs = imgRegexp.FindAllStringSubmatch(s, -1)
	for _, subexpr := range subexprs {
		s = fmt.Sprintf("%s<img src=\"%s\">%s",
			st.transformLine2(subexpr[1]),
			escapeParts(st.maybeAddRoot(subexpr[2])),
			st.transformLine2(subexpr[3]))
	}

	// !http://image(altstring)!
	subexprs = altImgRegexp.FindAllStringSubmatch(s, -1)
	for _, subexpr := range subexprs {
		s = escapeParts(fmt.Sprintf("%s<img src=\"%s\" alt=\"%s\">%s",
			st.transformLine2(subexpr[1]),
			escapeParts(st.maybeAddRoot(subexpr[2])),
			st.transformLine2(subexpr[3]),
			st.transformLine2(subexpr[4])))
	}

	subexprs = footnoteRefRegexp.FindAllStringSubmatch(s, -1)
	for _, subexpr := range subexprs {
		linkNum, _ := strconv.Atoi(subexpr[2])
		s = fmt.Sprintf("%s%s%s",
			st.transformLine2(subexpr[1]),
			escapeParts(fmt.Sprintf("<a id=fnr%d-%s href=#fn%d-%s title=\"Jump to footnote %d\"><sup class=footnote>%d</sup></a>",
				linkNum, st.hash, linkNum, st.hash, linkNum, linkNum)),
			st.transformLine2(subexpr[3]))
	}

	for _, val := range regexpTags {
		subexprs = val.tagRegexp.FindAllStringSubmatch(s, -1)
		for _, subexpr := range subexprs {
			s = fmt.Sprintf("%s<%s>%s</%s>%s",
				st.transformLine2(subexpr[1]),
				val.tagValue,
				subexpr[2],
				val.tagValue,
				st.transformLine2(subexpr[3]))
		}
	}

	return s
}

func (st *state) checkBlockQuote() {
	if st.inBlockQuote {
		if st.inLongBlockQuote {
			st.w.Write([]byte("</blockquote>"))
		} else {
			st.w.Write([]byte("</p></blockquote>"))
		}
		st.inBlockQuote = false
		st.inLongBlockQuote = false
		st.currentClassPart = ""
	}
}

func (st *state) checkPre() {
	if st.inPre {
		st.w.Write([]byte("</pre>"))
		st.inPre = false
		st.inLongPre = false
		st.currentClassPart = ""
	}
}

func (st *state) checkCode() {
	if st.inCode {
		st.w.Write([]byte("</code></pre>"))
		st.inCode = false
		st.inLongCode = false
		st.currentClassPart = ""
	}
}

func (st *state) clearListLevels() {
	if st.currentListLevel > 0 {
		for ; st.currentListLevel > 0; st.currentListLevel -= 1 {
			fmt.Fprintf(st.w, "</li></%s>", st.listType)
		}
		st.currentClassPart = ""
	}
}

func (st *state) checkAll() {
	st.checkBlockQuote()
	st.checkPre()
	st.checkCode()
	st.clearListLevels()
}

func (st *state) transform() {
	for {
		// Accept a line from the iterator.
		var s, ok = <-st.ch
		if !ok {
			st.checkAll()
			break
		}

		// Match an empty line.
		if len(s) == 0 && !st.inLongBlockQuote && !st.inLongPre && !st.inLongCode {
			st.checkAll()
			continue
		}

		// Match block modifiers.
		matched := false
		for _, handler := range blockHandlers {
			subexpr := handler.re.FindStringSubmatch(s)
			matched = subexpr != nil
			if matched {
				matched = handler.fn(st, subexpr)
				break
			}
		}

		if matched {
			continue
		}

		if st.inPre || st.inCode {
			fmt.Fprintf(st.w, "%s\n", EntityEscapeString(s, st.inCode, st.inPre))
			continue
		}

		// Bare line.
		st.clearListLevels()

		if !(linkRegexp.MatchString(s) || titleLinkRegexp.MatchString(s)) &&
			(imgRegexp.MatchString(s) || altImgRegexp.MatchString(s)) {
			fmt.Fprintf(st.w, "%s", st.transformLine(s))
		} else if st.inBlockQuote {
			if !st.inLongBlockQuote {
				fmt.Fprintf(st.w, "<br>%s", st.transformLine(s))
			} else if len(s) != 0 {
				fmt.Fprintf(st.w, "<p%s>%s</p>", st.currentClassPart, st.transformLine(s))
			}
		} else {
			fmt.Fprintf(st.w, "<p>%s</p>", st.transformLine(s))
		}
	}

	st.clearListLevels()
}

// GetTextileFullLinkFormatter returns a formatter that formats arbitrary values using
// a subset of Textile (http://www.textism.org/tools/textile).
// Its only difference to TextileFormatter is that any relative links are converted
// to absolute links.
func GetTextileFullLinkFormatter(root_url string) func(io.Writer, string, ...interface{}) {
	return func(w io.Writer, format string, value ...interface{}) {
		formatter(w, format, root_url, value...)
	}
}

// TextileFormatter formats arbitrary values using a subset of Textile
// (http://www.textism.org/tools/textile).
func TextileFormatter(w io.Writer, format string, value ...interface{}) {
	formatter(w, format, "", value...)
}

func formatter(w io.Writer, format string, root_url string, value ...interface{}) {
	ok := false
	var b []byte
	if len(value) == 1 {
		b, ok = value[0].([]byte)
	}
	if !ok {
		var buf bytes.Buffer
		fmt.Fprint(&buf, value...)
		b = buf.Bytes()
	}

	// Create an iterator that spits out single lines.
	ch := make(chan string)
	go func() {
		// Split into lines.
		for _, line := range lineRegexp.FindAllString(string(b), -1) {
			ch <- line
		}
		close(ch)
	}()

	// Iterate.
	var buf = &bytes.Buffer{}

	// Generate a unique id for the post that can be used by footnotes.
	h := md5.New()
	h.Write(buf.Bytes())
	hash := fmt.Sprintf("%x", h.Sum())
	(&state{ch: ch, w: buf, root_url: root_url, hash: hash}).transform()
	EntityEscape(w, buf.Bytes(), false, true)
}
