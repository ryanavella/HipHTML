package hiphtml

import "golang.org/x/net/html"

func isElem(n *html.Node) bool {
	return n.Type == html.ElementNode
}

func isText(n *html.Node) bool {
	return n.Type == html.TextNode
}
