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
	"testing"
)

var linetests = []struct {
	in  string
	out string
}{
	{"Hello", "<p>Hello</p>"},
	{"\"Hello\"", "<p>&#8220;Hello&#8221;</p>"},
	{"Hello\nthere", "<p>Hello</p><p>there</p>"},
	{"p. Hello", "<p>Hello</p>"},
	{"p(foo). Hello", "<p class=foo>Hello</p>"},
	{"p. Hello\np. there", "<p>Hello</p><p>there</p>"},
	{"p.Hello\np.there", "<p>p.Hello</p><p>p.there</p>"},
	{"bq. Hello there", "<blockquote><p>Hello there</p></blockquote>"},
	{"bq(foo). Hello there", "<blockquote class=foo><p class=foo>Hello there</p></blockquote>"},
	{"bq. Hello\nthere", "<blockquote><p>Hello<br>there</p></blockquote>"},
	{"bq. Hello\n\nthere", "<blockquote><p>Hello</p></blockquote><p>there</p>"},
	{"bq. Hello\np. there", "<blockquote><p>Hello</p></blockquote><p>there</p>"},
	{"bq.. Hello there", "<blockquote><p>Hello there</p></blockquote>"},
	{"bq.. Hello\n\nthere", "<blockquote><p>Hello</p><p>there</p></blockquote>"},
	{"bq(foo).. Hello\n\nthere", "<blockquote class=foo><p class=foo>Hello</p><p class=foo>there</p></blockquote>"},
	{"bq.. Hello\np. there", "<blockquote><p>Hello</p></blockquote><p>there</p>"},
	{"pre. Hello there", "<pre>Hello there\n</pre>"},
	{"pre(foo). Hello there", "<pre class=foo>Hello there\n</pre>"},
	{"pre. \"Hello\" & there", "<pre>&#8220;Hello&#8221; &#38; there\n</pre>"},
	{"pre.   Hello there", "<pre>  Hello there\n</pre>"},
	{"pre. Hello\nthere", "<pre>Hello\nthere\n</pre>"},
	{"pre. Hello\nthere\n\nxxx", "<pre>Hello\nthere\n</pre><p>xxx</p>"},
	{"pre. Hello\n\nthere", "<pre>Hello\n</pre><p>there</p>"},
	{"pre. Hello\np. there", "<pre>Hello\n</pre><p>there</p>"},
	{"pre.. Hello there", "<pre>Hello there\n</pre>"},
	{"pre(foo).. Hello there", "<pre class=foo>Hello there\n</pre>"},
	{"pre.. Hello\n\nthere", "<pre>Hello\n\nthere\n</pre>"},
	{"pre.. Hello\n\n  there", "<pre>Hello\n\n  there\n</pre>"},
	{"pre.. Hello\np. there", "<pre>Hello\n</pre><p>there</p>"},
	{"bc. Hello there", "<pre><code>Hello there\n</code></pre>"},
	{"bc(foo). Hello there", "<pre class=foo><code class=foo>Hello there\n</code></pre>"},
	{"bc. \"Hello\" & there", "<pre><code>&#34;Hello&#34; &#38; there\n</code></pre>"},
	{"bc.   Hello there", "<pre><code>  Hello there\n</code></pre>"},
	{"bc. Hello\nthere", "<pre><code>Hello\nthere\n</code></pre>"},
	{"bc. Hello\nthere\n\nxxx", "<pre><code>Hello\nthere\n</code></pre><p>xxx</p>"},
	{"bc. Hello\n\nthere", "<pre><code>Hello\n</code></pre><p>there</p>"},
	{"bc. Hello\np. there", "<pre><code>Hello\n</code></pre><p>there</p>"},
	{"bc.. Hello there", "<pre><code>Hello there\n</code></pre>"},
	{"bc(foo).. Hello there", "<pre class=foo><code class=foo>Hello there\n</code></pre>"},
	{"bc.. Hello\n\nthere", "<pre><code>Hello\n\nthere\n</code></pre>"},
	{"bc.. Hello\n\n  there", "<pre><code>Hello\n\n  there\n</code></pre>"},
	{"bc.. Hello\np. there", "<pre><code>Hello\n</code></pre><p>there</p>"},
	{"h1. Hello", "<h1>Hello</h1>"},
	{"h1. Hello\nthere", "<h1>Hello</h1><p>there</p>"},
	{"h1(foo). Hello", "<h1 class=foo>Hello</h1>"},
	{"Hello:there", "<p>Hello:there</p>"},
	{"\"Hello\":there", "<p><a href=\"there\">Hello</a></p>"},
	{"[\"Hello\":there_(foo_bar)]", "<p><a href=\"there_(foo_bar)\">Hello</a></p>"},
	{"[\"Hello\":there_(foo_bar)]", "<p><a href=\"there_(foo_bar)\">Hello</a></p>"},
	{"*[wewe \"Hello\":there]*", "<p><strong>[wewe <a href=\"there\">Hello</a>]</strong></p>"},
	{"\"Hello\":there!", "<p><a href=\"there\">Hello</a>!</p>"},
	{"\"Hello\":there!", "<p><a href=\"there\">Hello</a>!</p>"},
	{"\"Hello\":there_", "<p><a href=\"there_\">Hello</a></p>"},
	{"\"Hello\":there's", "<p><a href=\"there\">Hello</a>'s</p>"},
	{"_\"\"Hello\":there!\"_", "<p><em>&#8220;<a href=\"there\">Hello</a>!&#8221;</em></p>"},
	{"\"Hello\":there,", "<p><a href=\"there\">Hello</a>,</p>"},
	{"\"Hello(foo)\":there", "<p><a href=\"there\" title=\"foo\">Hello</a></p>"},
	{"foo \"Hello\":there bar", "<p>foo <a href=\"there\">Hello</a> bar</p>"},
	{"foo \"Hello\":there%20again bar", "<p>foo <a href=\"there%20again\">Hello</a> bar</p>"},
	{"foo \"Hello there\":http://foo.com/bar bar", "<p>foo <a href=\"http://foo.com/bar\">Hello there</a> bar</p>"},
	{"foo \"Hello there\":http://foo.com/bar.html bar", "<p>foo <a href=\"http://foo.com/bar.html\">Hello there</a> bar</p>"},
	{"foo \"Hello there\":http://foo.com/bar.html bar", "<p>foo <a href=\"http://foo.com/bar.html\">Hello there</a> bar</p>"},
	{"foo \"Hello there\":http://foo.com/bar.html, bar", "<p>foo <a href=\"http://foo.com/bar.html\">Hello there</a>, bar</p>"},
	{"_foo", "<p>_foo</p>"},
	{"_foo_", "<p><em>foo</em></p>"},
	{"x y _foo bar_", "<p>x y <em>foo bar</em></p>"},
	{"_x y_ _foo bar_ blogs", "<p><em>x y</em> <em>foo bar</em> blogs</p>"},
	{"*foo", "<p>*foo</p>"},
	{"*foo*", "<p><strong>foo</strong></p>"},
	{"_*foo*_", "<p><em><strong>foo</strong></em></p>"},
	{"_foo \"bar\":blogs bang_", "<p><em>foo <a href=\"blogs\">bar</a> bang</em></p>"},

	{"(\"RJ11\":http://en.wikipedia.org/wiki/Telephone_plug)", "<p>(<a href=\"http://en.wikipedia.org/wiki/Telephone_plug\">RJ11</a>)</p>"},
	{"x y *foo bar*", "<p>x y <strong>foo bar</strong></p>"},
	{"*x y* *foo bar* blogs", "<p><strong>x y</strong> <strong>foo bar</strong> blogs</p>"},
	{"!foo!", "<img src=\"foo\">"},
	{"!http://foo.com/img.jpg!", "<img src=\"http://foo.com/img.jpg\">"},
	{"!foo_bar!", "<img src=\"foo_bar\">"},
	{"! foo!", "<p>! foo!</p>"},
	{"!foo(bar)!", "<img src=\"foo\" alt=\"bar\">"},
	{"!foo(bar!)!", "<img src=\"foo\" alt=\"bar!\">"},
	{"!foo_bar_(bar the bar)!", "<img src=\"foo_bar_\" alt=\"bar the bar\">"},
	{"foo^bar^", "<p>foo<sup>bar</sup></p>"},
	{"foo~bar~", "<p>foo<sub>bar</sub></p>"},
	{"! foo!", "<p>! foo!</p>"},
	{"\"!foo!\":bar", "<p><a href=\"bar\"><img src=\"foo\"></a></p>"},
	{"\"!foo_bar_foo(bam!)!(boom!)\":bar_wib_wob", "<p><a href=\"bar_wib_wob\" title=\"boom!\"><img src=\"foo_bar_foo\" alt=\"bam!\"></a></p>"},
	{"\"!http://x/y_y_m.jpg(CJ)!(CJ)\":http://x/y_y_m.jpg", "<p><a href=\"http://x/y_y_m.jpg\" title=\"CJ\"><img src=\"http://x/y_y_m.jpg\" alt=\"CJ\"></a></p>"},
	{"\"!http://static.flickr.com/75/183891719_5dff6c3106_m.jpg(bar!)!\":bar/", "<p><a href=\"bar/\"><img src=\"http://static.flickr.com/75/183891719_5dff6c3106_m.jpg\" alt=\"bar!\"></a></p>"},
	{"* foo", "<ul><li>foo</li></ul>"},
	{"* foo\n* bar", "<ul><li>foo</li><li>bar</li></ul>"},
	{"* foo\n\n* bar", "<ul><li>foo</li></ul><ul><li>bar</li></ul>"},
	{"* foo\n** bar", "<ul><li>foo<ul><li>bar</li></ul></li></ul>"},
	{"* foo\n** bob\n* boo\nbar", "<ul><li>foo<ul><li>bob</li></ul></li><li>boo</li></ul><p>bar</p>"},
	{"* foo\n** bob\n*** boo", "<ul><li>foo<ul><li>bob<ul><li>boo</li></ul></li></ul></li></ul>"},
	{"* foo\n** bob\n** boo\n* fred\n", "<ul><li>foo<ul><li>bob</li><li>boo</li></ul></li><li>fred</li></ul>"},
	{"* foo\n** bob\n*** boo\n* fred\n", "<ul><li>foo<ul><li>bob<ul><li>boo</li></ul></li></ul></li><li>fred</li></ul>"},
	{"* foo\n** bob\n** boo\nbar", "<ul><li>foo<ul><li>bob</li><li>boo</li></ul></li></ul><p>bar</p>"},
	{"* l1\n** l2\n** l2\n\nbar", "<ul><li>l1<ul><li>l2</li><li>l2</li></ul></li></ul><p>bar</p>"},
	{"* <foo>& more...<bar>", "<ul><li><foo>&#38; more&#8230;<bar></li></ul>"},
	{"# foo", "<ol><li>foo</li></ol>"},
	{"# foo\n# bar", "<ol><li>foo</li><li>bar</li></ol>"},
	{"# foo\n\n# bar", "<ol><li>foo</li></ol><ol><li>bar</li></ol>"},
	{"# foo\n## bar", "<ol><li>foo<ol><li>bar</li></ol></li></ol>"},
	{"# foo\n## bob\n# boo\nbar", "<ol><li>foo<ol><li>bob</li></ol></li><li>boo</li></ol><p>bar</p>"},
	{"# foo\n## bob\n## boo\nbar", "<ol><li>foo<ol><li>bob</li><li>boo</li></ol></li></ol><p>bar</p>"},
	{"# l1\n## l2\n## l2\n\nbar", "<ol><li>l1<ol><li>l2</li><li>l2</li></ol></li></ol><p>bar</p>"},
	{"# <foo>& more...<bar>", "<ol><li><foo>&#38; more&#8230;<bar></li></ol>"},
	{"some foo[1]", "<p>some foo<a id=fnr1-d41d8cd98f00b204e9800998ecf8427e href=#fn1-d41d8cd98f00b204e9800998ecf8427e title=\"Jump to footnote 1\"><sup class=footnote>1</sup></a></p>"},
	{"fn1. some footnote", "<p id=fn1-d41d8cd98f00b204e9800998ecf8427e><sup>1</sup> some footnote&#160;<a href=#fnr1-d41d8cd98f00b204e9800998ecf8427e title=\"Jump back to footnote 1\">&#8617;</a></p>"},
	{"fn1(footnote). some footnote", "<p class=footnote id=fn1-d41d8cd98f00b204e9800998ecf8427e><sup class=footnote>1</sup> some footnote&#160;<a href=#fnr1-d41d8cd98f00b204e9800998ecf8427e title=\"Jump back to footnote 1\">&#8617;</a></p>"},
	{"* \"foo\":bar wibble", "<ul><li><a href=\"bar\">foo</a> wibble</li></ul>"},
	{"\"foo\":bar_wib", "<p><a href=\"bar_wib\">foo</a></p>"},
	{"\"foo\":bar_wib", "<p><a href=\"bar_wib\">foo</a></p>"},
	{"\"foo\":bar_wib \"wee\":bar_wib", "<p><a href=\"bar_wib\">foo</a> <a href=\"bar_wib\">wee</a></p>"},
	{"<foo>", "<p><foo></p>"},
	{"<foo title=\"bar\">", "<p><foo title=\"bar\"></p>"},
	{"This is \"<foo title=\"bar\">\"", "<p>This is &#8220;<foo title=\"bar\">&#8221;</p>"},
	{"<a href=\"http://foo_b_bob.jpg\" rel=\"lightbox\" title=\"Fun\"><img src=\"http://foo_b_bob.jpg\" alt=\"Fun\" /></a>", "<p><a href=\"http://foo_b_bob.jpg\" rel=\"lightbox\" title=\"Fun\"><img src=\"http://foo_b_bob.jpg\" alt=\"Fun\" /></a></p>"},
	{"\"\"foo\":bar\"", "<p>&#8220;<a href=\"bar\">foo</a>&#8221;</p>"},
}

var abslinetests = []struct {
	in  string
	out string
}{
	{"\"foo\":/bar", "<p><a href=\"http://site/bar\">foo</a></p>"},
	{"!/bar!", "<img src=\"http://site/bar\">"},
}


func TestTransformLine(t *testing.T) {
	for _, lt := range linetests {
		var buf bytes.Buffer
		TextileFormatter(&buf, "", lt.in)
		bs := buf.String()
		if bs != lt.out {
			t.Errorf("%s = '%s' want '%s'", lt.in, bs, lt.out)
		}
	}

	for _, lt := range abslinetests {
		var buf bytes.Buffer
		GetTextileFullLinkFormatter("http://site")(&buf, "", lt.in)
		bs := buf.String()
		if bs != lt.out {
			t.Errorf("%s = '%s' want '%s'", lt.in, bs, lt.out)
		}
	}
}
