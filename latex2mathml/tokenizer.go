package latex2mathml

import (
	"regexp"
	"strings"
)

var UNITS = []string{"in", "mm", "cm", "pt", "em", "ex", "pc", "bp", "dd", "cc", "sp", "mu"}
var PATTERN = []string{
	`(%[^\n]+)`,
	`(a-zA-Z)`,
	`([_^])(\d)`,
	`(-?\d+(?:\.\d+)?\s*(?:in|mm|cm|pt|em|ex|pc|bp|dd|cc|sp|mu))`,
	`(\d+(?:\.\d+)?)`,
	`(\.\d*)`,
	`(\\[\\\[\]{{}}\s!,:>;|_%#$&])`,
	`(\\(?:begin|end)\s*{[a-zA-Z]+\*?})`,
	`(\\operatorname\s*{[a-zA-Z\s*]+\*?\s*})`,
	`(\\(?:color|fbox|hbox|href|mbox|style|text|textbf|textit|textrm|textsf|texttt))\s*{([^}}]*)}`,
	`(\\[cdt]?frac)\s*([.\d])\s*([.\d])?`,
	`(\\math[a-z]+)({)([a-zA-Z])(})`,
	`(\\[a-zA-Z]+)`,
	`(\S)`,
}

var RE = regexp.MustCompile(strings.Join(PATTERN, "|"))

func Filter(vs []string, f func(string) bool) []string {
	filtered := make([]string, 0)
	for _, v := range vs {
		if f(v) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

func Tokenize(latex string) []string {
	var tokens = []string{}

	for _, latex_line := range strings.Split(latex, "\n") {
		for index, matchs := range RE.FindAllStringSubmatch(latex_line, -1) {
			for _, match := range matchs[1:] {
				if len(match) > 0 && !strings.HasPrefix(match, "%") {
					if index == 0 && strings.HasPrefix(match, MATH) {
						symbol, exists := Symbols[match]
						if exists {
							tokens = append(tokens, "&#x"+symbol+";")
							continue
						}
					} else {
						var add_token = false

						if match[0] == '_' || match[0] == '^' {
							tokens = append(tokens, match[0:1])
							tokens = append(tokens, match[1:])
							add_token = true
						}

						if !add_token {
							for _, unit := range UNITS {
								if strings.HasSuffix(match, unit) {
									tokens = append(tokens, strings.ReplaceAll(match, " ", ""))
									add_token = true
									break
								}
							}
						}

						if !add_token {
							for _, command := range []string{BEGIN, END, OPERATORNAME} {
								if strings.HasPrefix(match, command) {
									tokens = append(tokens, strings.ReplaceAll(match, " ", ""))
									add_token = true
									break
								}
							}
						}

						if !add_token {
							tokens = append(tokens, match)
						}
					}
				}
			}
		}
	}

	return Filter(tokens, func(s string) bool { return s != "" })
}
