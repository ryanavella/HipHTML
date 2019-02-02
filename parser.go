// Package hiphtml provides an interface for parsing arbitrary html.
//
// HipHTML is essentially a wrapper around html.Parse from golang.org/x/net/html that abstracts away some of the nitty gritty details.
// The /x/net/html package recommends writing a recursive descent parser for simple tasks like locating the <body> element, but as your
// requirements scale this very quickly leads to massive code duplication that is difficult to both navigate and maintain. Rather than
// reinvent the wheel for each new task, you can use HipHTML to navigate an html tree with just a few intuitive method calls.
package hiphtml

import (
	"errors"
	"io"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Parser errors
var (
	ErrBegOfDoc       = errors.New("reached beginning of document")
	ErrEndOfDoc       = errors.New("reached end of document")
	ErrNoSuchRelative = errors.New("node does not have requested rel")
)

// Parser is a HTML parser.
type Parser struct {
	doc   *html.Node
	node  *html.Node
	level int
}

// NewParser returns a html parser from an io.Reader.
func NewParser(r io.Reader) (p *Parser, err error) {
	doc, err := html.Parse(r)
	if err != nil {
		return p, err
	}
	p = &Parser{
		doc:  doc,
		node: doc,
	}
	return p, nil
}

// Node returns the current node of the html parser
func (p *Parser) Node() *html.Node {
	return p.node
}

// Level returns the current level of the html parser
//
// The <html> tag is fixed to level 1, and nested tags are one level higher than their parent.
func (p *Parser) Level() int {
	return p.level
}

// Next advances the parser to the next node if it exists.
func (p *Parser) Next() (*html.Node, error) {
	var err error
	_, err = p.FirstChild()
	if err == nil {
		return p.node, nil
	}
	_, err = p.nextSiblingAscending()
	if err == nil {
		return p.node, nil
	}
	return nil, ErrEndOfDoc
}

// Prev retreats the parser to the previous node if it exists.
func (p *Parser) Prev() (*html.Node, error) {
	var err error
	_, err = p.prevSiblingDescending()
	if err == nil {
		return p.node, nil
	}
	_, err = p.Parent()
	if err == nil {
		return p.node, nil
	}
	return nil, ErrBegOfDoc
}

// NextElement advances the parser to the next element node if it exists.
func (p *Parser) NextElement() (*html.Node, error) {
	var err error
	for !isElem(p.node) {
		_, err = p.Next()
		if err != nil {
			return nil, ErrEndOfDoc
		}
	}
	return p.node, nil
}

// PrevElement retreats the parser to the previous element node if it exists.
func (p *Parser) PrevElement() (*html.Node, error) {
	var err error
	for !isElem(p.node) {
		_, err = p.Prev()
		if err != nil {
			return nil, ErrBegOfDoc
		}
	}
	return p.node, nil
}

// nextSiblingAscending advances the parser to the next sibling, uncle, great uncle, etc. if it exists.
func (p *Parser) nextSiblingAscending() (*html.Node, error) {
	var err error
	node := p.node
	level := p.level
	for err == nil {
		_, err = p.NextSibling()
		if err == nil {
			return p.node, nil
		}
		_, err = p.Parent()
	}
	p.node = node
	p.level = level
	return nil, ErrNoSuchRelative
}

// prevSiblingDescending retreats the parser to the previous sibling, newphew, great nephew, etc. if it exists.
func (p *Parser) prevSiblingDescending() (*html.Node, error) {
	var err error
	_, err = p.PrevSibling()
	if err != nil {
		return nil, ErrNoSuchRelative
	}
	for err == nil {
		_, err = p.LastChild()
	}
	return p.node, nil
}

// NextSibling advances the parser to the next sibling if it exists.
func (p *Parser) NextSibling() (*html.Node, error) {
	if p.node.NextSibling == nil {
		return nil, ErrNoSuchRelative
	}
	p.node = p.node.NextSibling
	return p.node, nil
}

// PrevSibling retreats the parser to the previous sibling if it exists.
func (p *Parser) PrevSibling() (*html.Node, error) {
	if p.node.PrevSibling == nil {
		return nil, ErrNoSuchRelative
	}
	p.node = p.node.PrevSibling
	return p.node, nil
}

// Parent retreats the parser to the parent if it exists.
func (p *Parser) Parent() (*html.Node, error) {
	if p.node.Parent == nil {
		return nil, ErrNoSuchRelative
	}
	p.node = p.node.Parent
	p.level--
	return p.node, nil
}

// FirstChild advances the parser to the first child if it exists.
func (p *Parser) FirstChild() (*html.Node, error) {
	if p.node.FirstChild == nil {
		return nil, ErrNoSuchRelative
	}
	p.node = p.node.FirstChild
	p.level++
	return p.node, nil
}

// LastChild advances the parser to the last child if it exists.
func (p *Parser) LastChild() (*html.Node, error) {
	if p.node.LastChild == nil {
		return nil, ErrNoSuchRelative
	}
	p.node = p.node.LastChild
	p.level++
	return p.node, nil
}

// Body advances the parser to the body element in the document.
func (p *Parser) Body() (*html.Node, error) {
	return p.FirstElementByAtom(atom.Body)
}

// Head advances the parser to the head element in the document.
func (p *Parser) Head() (*html.Node, error) {
	return p.FirstElementByAtom(atom.Head)
}

// FirstMeta advances the parser to the first meta tag in the document.
func (p *Parser) FirstMeta() (*html.Node, error) {
	return p.FirstElementByAtom(atom.Meta)
}

// NextMeta advances the parser to the next meta tag in the document.
func (p *Parser) NextMeta() (*html.Node, error) {
	return p.NextElementByAtom(atom.Meta)
}

// FirstElementByAtom advances the parser to the first element in a document with the given atom.
func (p *Parser) FirstElementByAtom(a atom.Atom) (*html.Node, error) {
	var err error
	p.Reset()
	_, err = p.Next()
	if err != nil {
		return nil, err
	}
	for !(isElem(p.node) && p.node.DataAtom == a) {
		_, err = p.Next()
		if err != nil {
			return nil, err
		}
	}
	return p.node, nil
}

// Reset retreats the parser to the beginning of the document.
func (p *Parser) Reset() *html.Node {
	p.node = p.doc
	p.level = 0
	return p.node
}

// NextElementByAtom advances to the next element with the given atom.
func (p *Parser) NextElementByAtom(a atom.Atom) (*html.Node, error) {
	var err error
	_, err = p.Next()
	if err != nil {
		return nil, err
	}
	for !(isElem(p.node) && p.node.DataAtom == a) {
		_, err = p.Next()
		if err != nil {
			return nil, err
		}
	}
	return p.node, nil
}
