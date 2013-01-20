package xmlfilter

import (
	"bytes"
	"encoding/xml"
	"testing"
)

func TestNextNames(t *testing.T) {
	html := `<?xml version="1.0" encoding="utf-8" standalone="yes"?>
		<p>
			<a name="foo"/>
			<small>
				<font face="Arial">
					Foo
					<sup>
						<u>
							<b>
								Bar
							</b>
						</u>
					</sup>
				</font>
			</small>
			<a href="/path/to/somewhere">
				Baz
			</a>
		</p>

		<p>
			Ding
		</p>
	`

	expected := []string{
		"<p><a></a>Foo<sup>Bar</sup><a>Baz</a></p>",
		"<p>Ding</p>",
	}

	r := bytes.NewBufferString(html)
	d := xml.NewDecoder(r)

	for _, v := range expected {
		node, err := NextNames(d, xml.Name{Local:"p"}, xml.Name{Local:"a"},
			xml.Name{Local:"sup"})
		if err != nil {
			t.Fatal(err)
		}
		if node.String() != v {
			t.Errorf("expected %q, got %q", v, node.String())
		}
	}
}
