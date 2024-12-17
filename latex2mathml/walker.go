package latex2mathml

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/wanglihui/pptx-go/latex2mathml/slices"
)

type Node struct {
	Token      string
	Children   []Node
	Delimiter  string
	Alignment  string
	Text       string
	Attributes map[string]string
	Modifier   string
}

var (
	SUBSUP_LIST = []string{SUBSUP, SUBSCRIPT, SUPERSCRIPT}
	POS_LIST    = []string{HSKIP, HSPACE, KERN, MKERN, MSKIP, MSPACE}
	BOX_LIST    = []string{FBOX, HBOX, MBOX, MIDDLE, TEXT, TEXTBF, TEXTIT, TEXTRM, TEXTSF, TEXTTT}
	OTHER_LIST  = []string{ABOVE, ATOP, ABOVEWITHDELIMS, ATOPWITHDELIMS, BRACE, BRACK, CHOOSE, OVER}
)

func Walk(data string) ([]Node, error) {
	tokens := Tokenize(data)
	iterator := makeIterator(tokens)
	return processToken(iterator, "", 0)
}

func makeIterator(tokens []string) func() string {
	var index = 0

	return func() string {
		var token string
		if index < len(tokens) {
			token = tokens[index]
		}
		index = index + 1
		return token
	}
}

func containsKey[M ~map[K]V, K comparable, V any](m M, k K) bool {
	_, ok := m[k]
	return ok
}

func processToken(tokens func() string, terminator string, limit int) ([]Node, error) {
	var nodes = []Node{}
	var hasAvailableTokens = false
	var node Node

	for token := tokens(); token != ""; token = tokens() {
		hasAvailableTokens = true
		var delimiter = ""

		if token == terminator {
			if terminator == RIGHT {
				delimiter = tokens()
			}
			nodes = append(nodes, Node{Token: token, Delimiter: delimiter})
			break
		} else if token == RIGHT && terminator != RIGHT || token == MIDDLE && terminator != RIGHT {
			return nodes, errors.New("Extra left or missing right")
		} else if token == LEFT {
			delimiter = tokens()
			children, err := processToken(tokens, RIGHT, -1)
			if err != nil || len(children) == 0 || children[len(children)-1].Token != RIGHT {
				return nodes, errors.New("Extra left or missing right")
			}
			if len(children) > 0 {
				node = Node{Token: token, Children: children, Delimiter: delimiter}
			} else {
				node = Node{Token: token, Delimiter: delimiter}
			}
		} else if token == OPENING_BRACE {
			children, _ := processToken(tokens, CLOSING_BRACE, -1)
			if len(children) > 0 && children[len(children)-1].Token == CLOSING_BRACE {
				children = children[:len(children)-1]
			}
			node = Node{Token: BRACES, Children: children}
		} else if token == SUBSCRIPT || token == SUPERSCRIPT {
			var previous Node
			if len(nodes) > 0 {
				previous = nodes[len(nodes)-1]
				nodes = nodes[:len(nodes)-1]
			} else {
				previous = Node{}
			}

			if token == previous.Token && token == SUBSCRIPT {
				return nodes, errors.New("Double subscript")
			}
			if token == previous.Token && token == SUPERSCRIPT && len(previous.Children) >= 2 && previous.Children[1].Token != PRIME {
				return nodes, errors.New("Double superscript")
			}

			var modifier string

			if previous.Token == LIMITS {
				modifier = LIMITS
				if len(nodes) > 0 {
					previous = nodes[len(nodes)-1]
					nodes = nodes[:len(nodes)-1]

					if !strings.HasSuffix(previous.Token, `\\`) {
						return nodes, errors.New("Limits must flow math operator")
					}
				} else {
					return nodes, errors.New("Limits must flow math operator")
				}
			}

			if token == SUBSCRIPT && previous.Token == SUPERSCRIPT && len(previous.Children) > 0 {
				children, _ := processToken(tokens, terminator, 1)
				children = append([]Node{previous.Children[0]}, children...)
				children = append(children, previous.Children[1])
				node = Node{Token: SUBSUP, Children: children, Delimiter: previous.Delimiter}
			} else if token == SUPERSCRIPT && previous.Token == SUBSCRIPT && len(previous.Children) > 0 {
				children, _ := processToken(tokens, terminator, 1)
				children = append(previous.Children, children...)
				node = Node{Token: SUBSUP, Children: children, Delimiter: previous.Delimiter}
			} else if token == SUPERSCRIPT && previous.Token == SUPERSCRIPT && len(previous.Children) > 0 && previous.Children[1].Token == PRIME {
				children, _ := processToken(tokens, terminator, 1)
				children = append([]Node{previous.Children[1]}, children...)
				node = Node{
					Token:     SUBSUP,
					Children:  []Node{previous.Children[0], {Token: BRACES, Children: children}},
					Delimiter: previous.Delimiter,
				}
			} else {
				children, err := processToken(tokens, terminator, 1)

				if err != nil {
					return nodes, errors.New("Missing superscript or subscript")
				}

				if previous.Token == OVERBRACE || previous.Token == UNDERBRACE {
					modifier = previous.Token
				}
				children = append([]Node{previous}, children...)
				node = Node{Token: token, Children: children, Modifier: modifier}
			}
		} else if token == APOSTROPHE {
			var previous Node
			if len(nodes) > 0 {
				previous = nodes[len(nodes)-1]
				nodes = nodes[:len(nodes)-1]
			} else {
				previous = Node{}
			}

			if previous.Token == SUPERSCRIPT && len(previous.Children) >= 2 && previous.Children[1].Token != PRIME {
				return nodes, errors.New("Double superscripts")
			}

			if previous.Token == SUPERSCRIPT && len(previous.Children) >= 2 && previous.Children[1].Token == PRIME {
				node = Node{Token: SUPERSCRIPT, Children: []Node{previous.Children[0], {Token: DPRIME}}}
			} else if previous.Token == SUPERSCRIPT && len(previous.Children) > 0 {
				node = Node{
					Token:    SUBSUP,
					Children: append(previous.Children, Node{Token: PRIME}),
					Modifier: previous.Modifier,
				}
			} else {
				node = Node{Token: SUPERSCRIPT, Children: []Node{previous, {Token: PRIME}}}
			}

		} else if slices.Contains(COMMANDS_WITH_TWO_PARAMETERS, token) {
			children, _ := processToken(tokens, terminator, 2)
			if token == OVERSET || token == UNDERSET {
				slices.Reverse(children)
			}
			node = Node{Token: token, Children: children}
		} else if slices.Contains(COMMANDS_WITH_ONE_PARAMETER, token) || strings.HasPrefix(token, MATH) {
			children, _ := processToken(tokens, terminator, 1)
			node = Node{Token: token, Children: children}
		} else if token == NOT {
			next, err := processToken(tokens, terminator, 1)
			if err != nil {
				nextNode := next[0]
				if strings.HasPrefix(nextNode.Token, `\\`) {
					negatedSymbol := `\n` + nextNode.Token[1:]
					_, err := ConvertSymbol(negatedSymbol)
					if err != nil {
						node = Node{Token: negatedSymbol}
						nodes = append(nodes, node)
						continue
					}
					node = Node{Token: token}
					nodes = append(nodes, node, nextNode)
				}
			} else {
				node = Node{Token: token}
			}
		} else if token == XLEFTARROW || token == XRIGHTARROW {
			children, _ := processToken(tokens, terminator, 1)
			if children[0].Token == OPENING_BRACKET {
				childrenChildren, _ := processToken(tokens, CLOSING_BRACKET, -1)
				childrenChildren = childrenChildren[:len(childrenChildren)-1]
				childrenNext, _ := processToken(tokens, terminator, 1)
				children = append([]Node{{Token: BRACES, Children: childrenChildren}}, childrenNext...)
			}
			node = Node{Token: token, Children: children}
		} else if slices.Contains(POS_LIST, token) {
			children, _ := processToken(tokens, terminator, 1)
			if len(children) > 0 && children[0].Token == BRACES && len(children[0].Children) > 0 {
				children = children[0].Children
			}
			node = Node{Token: token, Attributes: map[string]string{"width": children[0].Token}}
		} else if token == COLOR {
			attributes := map[string]string{"mathcolor": tokens()}
			children, _ := processToken(tokens, terminator, -1)
			if len(children) > 0 && children[len(children)-1].Token == terminator {
				sibling := children[len(children)-1]
				children = children[:len(children)-1]
				nodes = append(nodes, Node{Token: token, Children: children, Attributes: attributes})
				nodes = append(nodes, sibling)
			} else {
				nodes = append(nodes, Node{Token: token, Children: children, Attributes: attributes})
			}
			break
		} else if token == STYLE {
			attributes := map[string]string{"style": tokens()}
			next, _ := processToken(tokens, terminator, 1)
			node = next[0]
			node.Attributes = attributes
		} else if slices.Contains(BOX_LIST, token) || containsKey(BIG, token) || containsKey(BIG_OPEN_CLOSE, token) {
			node = Node{Token: token, Text: tokens()}
		} else if token == HREF {
			attributes := map[string]string{"href": tokens()}
			children, _ := processToken(tokens, terminator, 1)
			node = Node{Token: token, Children: children, Attributes: attributes}
		} else if slices.Contains(OTHER_LIST, token) {
			var delimiter = ""
			var attributes = map[string]string{}

			if token == ABOVEWITHDELIMS {
				delimiter = strings.TrimLeft(tokens(), `\\`) + strings.TrimLeft(tokens(), `\\`)
			} else if token == ATOPWITHDELIMS {
				attributes = map[string]string{"linethickness": "0"}
				delimiter = strings.TrimLeft(tokens(), `\\`) + strings.TrimLeft(tokens(), `\\`)
			} else if token == BRACE {
				delimiter = "{}"
			} else if token == BRACK {
				delimiter = "[]"
			} else if token == CHOOSE {
				delimiter = "()"
			}

			if token == ABOVE || token == ABOVEWITHDELIMS {
				dimensionNodes, _ := processToken(tokens, terminator, 1)
				dimension := getDimension(dimensionNodes[0])
				attributes = map[string]string{"linethickness": dimension}
			} else if token == ATOP || token == BRACE || token == BRACK || token == CHOOSE {
				attributes = map[string]string{"linethickness": "0"}
			}

			denominator, _ := processToken(tokens, terminator, -1)

			var sibling Node
			var children []Node

			if len(denominator) > 0 && denominator[len(denominator)-1].Token == terminator {
				sibling = denominator[len(denominator)-1]
				denominator = denominator[:len(denominator)-1]
			}

			if len(denominator) == 0 {
				if token == BRACE || token == BRACK {
					denominator = []Node{{Token: BRACES}}
				} else {
					return nodes, errors.New("Denominator not found")
				}
			}

			if len(nodes) == 0 {
				if token == BRACE || token == BRACK {
					nodes = []Node{{Token: BRACES}}
				} else {
					return nodes, errors.New("Numerator not found")
				}

			}

			if len(denominator) > 1 {
				denominator = []Node{{Token: BRACES, Children: denominator}}
			}

			if len(nodes) == 1 {
				children = append(nodes, denominator...)
			} else {
				children = []Node{{Token: BRACES, Children: nodes}}
				children = append(children, denominator...)
			}

			nodes = []Node{{Token: FRAC, Children: children, Attributes: attributes, Delimiter: delimiter}}
			if sibling.Token != "" {
				nodes = append(nodes, sibling)
			}
			break

		} else if token == SQRT {
			next, _ := processToken(tokens, "", 1)
			nextNode := next[0]
			var rootNodes = []Node{}

			if nextNode.Token == OPENING_BRACKET {
				rootNodes, _ = processToken(tokens, CLOSING_BRACKET, -1)
				rootNodes = rootNodes[:len(rootNodes)-1]

				if len(rootNodes) > 1 {
					rootNodes = []Node{
						{
							Token:    BRACES,
							Children: rootNodes,
						},
					}
				}
			}

			if len(rootNodes) > 0 {
				node = Node{Token: ROOT, Children: append([]Node{nextNode}, rootNodes...)}
			} else {
				node = Node{Token: token, Children: []Node{nextNode}}
			}

		} else if token == ROOT {
			rootNodes, _ := processToken(tokens, CLOSING_BRACKET, -1)
			rootNodes = rootNodes[:len(rootNodes)-1]
			next, _ := processToken(tokens, "", 1)
			nextNode := next[0]

			if len(rootNodes) > 1 {
				rootNodes = []Node{{Token: BRACES, Children: rootNodes}}
			}

			if len(rootNodes) > 0 {
				node = Node{Token: token, Children: append([]Node{nextNode}, rootNodes...)}
			} else {
				node = Node{Token: token, Children: []Node{nextNode, {Token: BRACES}}}
			}

		} else if slices.Contains(MATRICES, token) {
			children, _ := processToken(tokens, terminator, -1)
			var sibling Node

			if len(children) > 0 && children[len(children)-1].Token == terminator {
				sibling = children[len(children)-1]
				children = children[:len(children)-1]
			}

			if len(children) == 1 && children[0].Token == BRACES && len(children[0].Children) > 0 {
				children = children[0].Children
			}
			if sibling.Token != "" {
				nodes = append(nodes, Node{Token: token, Children: children}, sibling)
				break
			} else {
				node = Node{Token: token, Children: children}
			}

		} else if token == GENFRAC {
			delimiter := strings.TrimLeft(tokens(), `\\`) + strings.TrimLeft(tokens(), `\\`)
			next, _ := processToken(tokens, terminator, 2)
			dimension := getDimension(next[0])
			style, _ := getStyle(next[1])
			attributes := map[string]string{"linethickness": dimension}
			children, _ := processToken(tokens, terminator, 2)
			nodes = append(nodes, Node{Token: style}, Node{Token: token, Children: children, Delimiter: delimiter, Attributes: attributes})
		} else if token == SIDESET {
			next, _ := processToken(tokens, terminator, 3)
			operator := next[2]

			leftToken, leftChildren, _ := makeSupSub(next[0])
			rightToken, rightChildren, _ := makeSupSub(next[1])

			attributes := map[string]string{"movablelimits": "false"}

			node = Node{
				Token: token,
				Children: []Node{
					{
						Token: leftToken,
						Children: append(
							[]Node{
								{
									Token: VPHANTOM,
									Children: []Node{
										{
											Token:      operator.Token,
											Children:   operator.Children,
											Attributes: attributes,
										},
									},
								},
							},
							leftChildren...),
					},
					{
						Token: rightToken,
						Children: append(
							[]Node{
								{
									Token:      operator.Token,
									Children:   operator.Children,
									Attributes: attributes,
								},
							},
							rightChildren...,
						),
					},
				},
			}

		} else if token == SKEW {
			next, _ := processToken(tokens, terminator, 2)
			width := next[0].Token

			if width == BRACES {
				if len(next[0].Children) == 0 {
					return nodes, errors.New("Invalid width")
				}
				width = next[0].Children[0].Token
			}

			value, err := strconv.ParseInt(width, 10, 64)

			if err != nil {
				return nodes, errors.New("Invalid width")
			}

			node = Node{
				Token:      token,
				Children:   []Node{next[1]},
				Attributes: map[string]string{"width": fmt.Sprintf("%gem", 0.0555*float64(value))},
			}

		} else if strings.HasPrefix(token, BEGIN) {
			node, _ = getEnvinmentNode(token, tokens)
		} else {
			node = Node{Token: token}
		}

		nodes = append(nodes, node)

		if limit > 0 && len(nodes) >= limit {
			break
		}
	}

	if !hasAvailableTokens {
		return nodes, errors.New("No available tokens")
	}

	return nodes, nil
}

func makeSupSub(node Node) (string, []Node, error) {
	if node.Token != BRACES {
		return "", []Node{}, errors.New("Token is not BRACES `" + BRACES + "`")
	}

	if len(node.Children) > 0 && 2 <= len(node.Children[0].Children) && len(node.Children[0].Children) <= 3 && slices.Contains(SUBSUP_LIST, node.Children[0].Token) {
		return node.Children[0].Token, node.Children[0].Children[1:], nil
	}

	return "", []Node{}, errors.New("Index error in makeSupSub")
}

func getDimension(node Node) string {
	var dimension = node.Token
	if node.Token == BRACES && len(node.Children) > 0 {
		dimension = node.Children[0].Token
	}
	return dimension
}

func getStyle(node Node) (string, error) {
	var style = node.Token
	if node.Token == BRACES && len(node.Children) > 0 {
		style = node.Children[0].Token
	}

	switch style {
	case "0":
		return DISPLAYSTYLE, nil
	case "1":
		return TEXTSTYLE, nil
	case "2":
		return SCRIPTSTYLE, nil
	case "3":
		return SCRIPTSCRIPTSTYLE, nil
	}

	return "", errors.New("Invalid style for node")
}

func getEnvinmentNode(token string, tokens func() string) (Node, error) {
	startIndex := strings.Index(token, "{") + 1
	environment := token[startIndex : len(token)-1]
	terminator := END + "{" + environment + "}"
	children, _ := processToken(tokens, terminator, 0)

	if len(children) > 0 && children[len(children)-1].Token != terminator {
		return Node{}, errors.New("Missing end in tokens")
	}

	children = children[:len(children)-1]
	alignment := ""

	if len(children) > 0 && children[0].Token == OPENING_BRACKET {
		var index int
		var c Node
		for index, c = range children[1:] {
			if c.Token == CLOSING_BRACKET {
				break
			} else if strings.Contains("lcr|", c.Token) {
				return Node{}, errors.New("Invalid alignment error")
			}
			alignment = alignment + c.Token
		}
		children = children[index:]
	} else if len(children) > 0 && len(children[0].Children) > 0 && (children[0].Token == BRACES || strings.HasSuffix(environment, "*") && children[0].Token == BRACKETS) {
		var allAlignment = true
		var innerAlignment = ""
		for _, c := range children[0].Children {
			if !strings.Contains("lcr|", c.Token) {
				allAlignment = false
				innerAlignment = innerAlignment + c.Token
				break
			}
		}
		if allAlignment {
			alignment = alignment + innerAlignment
			children = children[1:]
		}
	}

	return Node{Token: `\` + environment, Children: children, Alignment: alignment}, nil
}
