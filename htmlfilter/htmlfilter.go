// Copyright 2013 Rodrigo Moraes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package htmlfilter provides some helpers for exp/html.
package htmlfilter

import (
	"bytes"
	"exp/html"
	"fmt"
	"strings"
)

// Node stores a tree of nodes we are interested in.
type Node struct {
	Token html.Token
	List  []Node
}

// String returns a simplified string version of the node, just for debugging
// and testing purposes. It doesn't include tag attributes.
func (n Node) String() string {
	switch n.Token.Type {
	case html.SelfClosingTagToken:
		return fmt.Sprintf("<%s/>", n.Token.Data)
	case html.StartTagToken:
		b := new(bytes.Buffer)
		fmt.Fprintf(b, "<%s>", n.Token.Data)
		for _, v := range n.List {
			fmt.Fprintf(b, "%s", v)
		}
		fmt.Fprintf(b, "</%s>", n.Token.Data)
		return b.String()
	case html.TextToken:
		return n.Token.Data
	}
	panic(fmt.Errorf("unexpected %T", n.Token))
}

// NextTextFilter returns a node containing all text in the next tag of the
// given names. It trims spaces from the text and ignores it if it is empty.
//
// It expects well-formed HTML: mismatching closing tags will result in error.
func NextTextFilter(t *html.Tokenizer, tagNames ...string) (Node, error) {
	tt, err := NextStartTag(t, tagNames...)
	if err != nil {
		return Node{}, err
	}
	list, err := TextInTag(t, tt.Data, tagNames...)
	if err != nil {
		return Node{}, err
	}
	return Node{Token: tt, List: list}, nil
}

// NextStartTag skips everything until we find a start tag of the given names.
func NextStartTag(t *html.Tokenizer, tagNames ...string) (html.Token, error) {
	for {
		switch t.Next() {
		case html.ErrorToken:
			return html.Token{}, t.Err()
		case html.SelfClosingTagToken, html.StartTagToken:
			tt := t.Token()
			for _, v := range tagNames {
				if v == tt.Data {
					return tt, nil
				}
			}
		}
	}
	panic("unreachable")
}

// TextInTag returns nodes containing all text until the given startTag is
// closed, including the tags of the given names.
//
// It expects well-formed HTML: mismatching closing tags will result in error.
func TextInTag(t *html.Tokenizer, startTag string, tagNames ...string) ([]Node, error) {
	stack := []string{startTag}
	var c []Node
	for len(stack) > 0 {
		// A token can be of the following types:
		//
		//   html.ErrorToken
		//   html.TextToken
		//   html.StartTagToken
		//   html.EndTagToken
		//   html.SelfClosingTagToken
		//   html.CommentToken
		//   html.DoctypeToken
		switch t.Next() {
		case html.ErrorToken:
			return nil, fmt.Errorf("unclosed tags: %v", stack)
		case html.SelfClosingTagToken:
			tt := t.Token()
			tag := tt.Data
			SelfClosingTagLoop:
			for _, v := range tagNames {
				if v == tag {
					c = append(c, Node{Token: tt})
					break SelfClosingTagLoop
				}
			}
		case html.StartTagToken:
			tt := t.Token()
			tag := tt.Data
			found := false
			StartTagLoop:
			for _, v := range tagNames {
				if v == tag {
					found = true
					list, err := TextInTag(t, tag, tagNames...)
					if err != nil {
						return nil, err
					}
					c = append(c, Node{Token: tt, List: list})
					break StartTagLoop
				}
			}
			if !found {
				stack = append(stack, tag)
			}
		case html.EndTagToken:
			var err error
			tag, _ := t.TagName()
			if stack, err = popTag(stack, string(tag)); err != nil {
				return nil, err
			}
		case html.TextToken:
			tt := t.Token()
			tt.Data = strings.TrimSpace(tt.Data)
			if len(tt.Data) > 0 {
				c = append(c, Node{Token: tt})
			}
		}
	}
	return c, nil
}

// popTag removes an expected tag name from a list, for balanced closing tags.
func popTag(tags []string, tag string) ([]string, error) {
	size := len(tags)
	if size == 0 {
		return nil, fmt.Errorf("unexpected closing tag %q", tag)
	}
	if last := tags[size-1]; last != tag {
		return nil, fmt.Errorf("expecting closing tag %q, got %q", last, tag)
	}
	return tags[:size-1], nil
}
