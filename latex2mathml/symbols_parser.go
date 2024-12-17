package latex2mathml

import (
	"bufio"
	"embed"
	"errors"
	"regexp"
	"strings"
)

//go:embed `unimathsymbols.txt`
var SymbolsFile embed.FS

var Symbols = map[string]string{}

var SYMBOLS_INITILIZED = false

func ConvertSymbol(symbol string) (string, error) {
	code, exist := Symbols[symbol]
	if exist {
		return code, nil
	} else {
		return code, errors.New("Symbol not found")
	}
}

func ParseSymbol() {
	if !SYMBOLS_INITILIZED {
		file, _ := SymbolsFile.Open("unimathsymbols.txt")

		re := regexp.MustCompile(`[=#]\s*(\\[^,^ ]+),?`)
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			line := scanner.Text()

			if !strings.HasPrefix(line, "#") {
				columns := strings.Split(strings.Trim(line, " "), "^")

				for i := 1; i <= 3; i++ {
					_, exists := Symbols[columns[i]]

					if columns[i] != "" && !exists {
						Symbols[columns[i]] = columns[0]
					}
				}

				for _, equivalents := range re.FindAllStringSubmatch(columns[len(columns)-1], -1) {
					for _, equivalent := range equivalents[1:] {
						if len(equivalent) > 0 {
							_, exists := Symbols[equivalent]
							if !exists {
								Symbols[equivalent] = columns[0]
							}
						}
					}
				}

			}
		}

		var symbolsUpdate = map[string]string{
			`\And`:            Symbols[`\ampersand`],
			`\bigcirc`:        Symbols[`\lgwhtcircle`],
			`\Box`:            Symbols[`\square`],
			`\circledS`:       "024C8",
			`\diagdown`:       "02572",
			`\diagup`:         "02571",
			`\dots`:           "02026",
			`\dotsb`:          Symbols[`\cdots`],
			`\dotsc`:          "02026",
			`\dotsi`:          Symbols[`\cdots`],
			`\dotsm`:          Symbols[`\cdots`],
			`\dotso`:          "02026",
			`\emptyset`:       "02205",
			`\gggt`:           "022D9",
			`\gvertneqq`:      "02269",
			`\gt`:             Symbols[`\greate`],
			`\ldotp`:          Symbols[`\period`],
			`\llless`:         Symbols[`\lll`],
			`\lt`:             Symbols[`\less`],
			`\lvert`:          Symbols[`\vert`],
			`\lVert`:          Symbols[`\Vert`],
			`\lvertneqq`:      Symbols[`\lneqq`],
			`\ngeqq`:          Symbols[`\ngeq`],
			`\nshortmid`:      Symbols[`\nmid`],
			`\nshortparallel`: Symbols[`\nparallel`],
			`\nsubseteqq`:     Symbols[`\nsubseteq`],
			`\omicron`:        Symbols[`\upomicron`],
			`\rvert`:          Symbols[`\vert`],
			`\rVert`:          Symbols[`\Vert`],
			`\shortmid`:       Symbols[`\mid`],
			`\smallfrown`:     Symbols[`\frown`],
			`\smallint`:       "0222B",
			`\smallsmile`:     Symbols[`\smile`],
			`\surd`:           Symbols[`\sqrt`],
			`\thicksim`:       "0223C",
			`\thickapprox`:    Symbols[`\approx`],
			`\varsubsetneqq`:  Symbols[`\subsetneqq`],
			`\varsupsetneq`:   "0228B",
			`\varsupsetneqq`:  Symbols[`\supsetneqq`],
		}

		for key, value := range symbolsUpdate {
			Symbols[key] = value
		}

		delete(Symbols, `\mathring`)
		SYMBOLS_INITILIZED = true
	}
}
