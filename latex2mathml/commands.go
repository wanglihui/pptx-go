package latex2mathml

type Style struct {
	Tag       string
	Modifiers map[string]string
}

const (
	OPENING_BRACE = "{"
	CLOSING_BRACE = "}"
	BRACES        = "{}"

	OPENING_BRACKET = "["
	CLOSING_BRACKET = "]"
	BRACKETS        = "[]"

	OPENING_PARENTHESIS = "("
	CLOSING_PARENTHESIS = ")"
	PARENTHESES         = "()"

	SUBSUP      = "_^"
	SUBSCRIPT   = "_"
	SUPERSCRIPT = "^"
	APOSTROPHE  = "'"

	PRIME  = `\prime`
	DPRIME = `\dprime`

	LEFT   = `\left`
	MIDDLE = `\middle`
	RIGHT  = `\right`

	ABOVE           = `\above`
	ABOVEWITHDELIMS = `\abovewithdelims`
	ATOP            = `\atop`
	ATOPWITHDELIMS  = `\atopwithdelims`
	BINOM           = `\binom`
	BRACE           = `\brace`
	BRACK           = `\brack`
	CFRAC           = `\cfrac`
	CHOOSE          = `\choose`
	DBINOM          = `\dbinom`
	DFRAC           = `\dfrac`
	FRAC            = `\frac`
	GENFRAC         = `\genfrac`
	OVER            = `\over`
	TBINOM          = `\tbinom`
	TFRAC           = `\tfrac`

	ROOT = `\root`
	SQRT = `\sqrt`

	OVERSET  = `\overset`
	UNDERSET = `\underset`

	ACUTE               = `\acute`
	BAR                 = `\bar`
	BREVE               = `\breve`
	CHECK               = `\check`
	DOT                 = `\dot`
	DDOT                = `\ddot`
	DDDOT               = `\dddot`
	DDDDOT              = `\ddddot`
	GRAVE               = `\grave`
	HAT                 = `\hat`
	MATHRING            = `\mathring`
	OVERBRACE           = `\overbrace`
	OVERLEFTARROW       = `\overleftarrow`
	OVERLEFTRIGHTARROW  = `\overleftrightarrow`
	OVERLINE            = `\overline`
	OVERPAREN           = `\overparen`
	OVERRIGHTARROW      = `\overrightarrow`
	TILDE               = `\tilde`
	UNDERBRACE          = `\underbrace`
	UNDERLEFTARROW      = `\underleftarrow`
	UNDERLINE           = `\underline`
	UNDERPAREN          = `\underparen`
	UNDERRIGHTARROW     = `\underrightarrow`
	UNDERLEFTRIGHTARROW = `\underleftrightarrow`
	VEC                 = `\vec`
	WIDEHAT             = `\widehat`
	WIDETILDE           = `\widetilde`
	XLEFTARROW          = `\xleftarrow`
	XRIGHTARROW         = `\xrightarrow`

	HREF   = `\href`
	TEXT   = `\text`
	TEXTBF = `\textbf`
	TEXTIT = `\textit`
	TEXTRM = `\textrm`
	TEXTSF = `\textsf`
	TEXTTT = `\texttt`

	BEGIN = `\begin`
	END   = `\end`

	LIMITS    = `\limits`
	INTEGRAL  = `\int`
	SUMMATION = `\sum`

	OPERATORNAME = `\operatorname`

	LBRACE = `\{`

	DETERMINANT = `\det`
	GCD         = `\gcd`
	INTOP       = `\intop`
	INJLIM      = `\injlim`
	LIMINF      = `\liminf`
	LIMSUP      = `\limsup`
	PR          = `\Pr`
	PROJLIM     = `\projlim`
	MOD         = `\mod`
	PMOD        = `\pmod`
	BMOD        = `\bmod`

	HDASHLINE = `\hdashline`
	HLINE     = `\hline`
	HFIL      = `\hfil`

	CASES        = `\cases`
	DISPLAYLINES = `\displaylines`
	SMALLMATRIX  = `\smallmatrix`
	SUBSTACK     = `\substack`
	SPLIT        = `\split`
	ALIGN        = `\align*`

	BACKSLASH       = `\`
	CARRIAGE_RETURN = `\cr`

	COLON           = `\:`
	COMMA           = `\,`
	DOUBLEBACKSLASH = `\\`
	ENSPACE         = `\enspace`
	EXCLAMATION     = `\!`
	GREATER_THAN    = `\>`
	HSKIP           = `\hskip`
	HSPACE          = `\hspace`
	KERN            = `\kern`
	MKERN           = `\mkern`
	MSKIP           = `\mskip`
	MSPACE          = `\mspace`
	NEGTHINSPACE    = `\negthinspace`
	NEGMEDSPACE     = `\negmedspace`
	NEGTHICKSPACE   = `\negthickspace`
	NOBREAKSPACE    = `\nobreakspace`
	SPACE           = `\space`
	THINSPACE       = `\thinspace`
	QQUAD           = `\qquad`
	QUAD            = `\quad`
	SEMICOLON       = `\;`

	BLACKBOARD_BOLD = `\Bbb`
	BOLD_SYMBOL     = `\boldsymbol`
	MIT             = `\mit`
	OLDSTYLE        = `\oldstyle`
	SCR             = `\scr`
	TT              = `\tt`

	MATH     = `\math`
	MATHBB   = `\mathbb`
	MATHBF   = `\mathbf`
	MATHCAL  = `\mathcal`
	MATHFRAK = `\mathfrak`
	MATHIT   = `\mathit`
	MATHRM   = `\mathrm`
	MATHSCR  = `\mathscr`
	MATHSF   = `\mathsf`
	MATHTT   = `\mathtt`

	BOXED = `\boxed`
	FBOX  = `\fbox`
	HBOX  = `\hbox`
	MBOX  = `\mbox`

	COLOR             = `\color`
	DISPLAYSTYLE      = `\displaystyle`
	TEXTSTYLE         = `\textstyle`
	SCRIPTSTYLE       = `\scriptstyle`
	SCRIPTSCRIPTSTYLE = `\scriptscriptstyle`
	STYLE             = `\style`

	HPHANTOM = `\hphantom`
	PHANTOM  = `\phantom`
	VPHANTOM = `\vphantom`

	IDOTSINT = `\idotsint`
	LATEX    = `\LaTeX`
	TEX      = `\TeX`

	SIDESET = `\sideset`

	SKEW = `\skew`
	NOT  = `\not`
)

var (
	COMMANDS_INITILIAZED = false

	LIMIT = []string{`\lim`, `\sup`, `\inf`, `\max`, `\min`}

	FUNCTIONS = []string{
		`\arccos`,
		`\arcsin`,
		`\arctan`,
		`\cos`,
		`\cosh`,
		`\cot`,
		`\coth`,
		`\csc`,
		`\deg`,
		`\dim`,
		`\exp`,
		`\hom`,
		`\ker`,
		`\ln`,
		`\lg`,
		`\log`,
		`\sec`,
		`\sin`,
		`\sinh`,
		`\tan`,
		`\tanh`,
	}

	MATRICES = []string{
		`\matrix`,
		`\matrix*`,
		`\pmatrix`,
		`\pmatrix*`,
		`\bmatrix`,
		`\bmatrix*`,
		`\Bmatrix`,
		`\Bmatrix*`,
		`\vmatrix`,
		`\vmatrix*`,
		`\Vmatrix`,
		`\Vmatrix*`,
		`\array`,
		SUBSTACK,
		CASES,
		DISPLAYLINES,
		SMALLMATRIX,
		SPLIT,
		ALIGN,
	}

	LOCAL_FONTS = map[string]map[string]string{
		BLACKBOARD_BOLD: {"default": "double-struck", "fence": "none"},
		BOLD_SYMBOL:     {"default": "bold", "mi": "bold-italic", "mtext": "none"},
		MATHBB:          {"default": "double-struck", "fence": "none"},
		MATHBF:          {"default": "bold", "fence": "none"},
		MATHCAL:         {"default": "script", "fence": "none"},
		MATHFRAK:        {"default": "fraktur", "fence": "none"},
		MATHIT:          {"default": "italic", "fence": "none"},
		MATHRM:          {"default": "none", "mi": "normal"},
		MATHSCR:         {"default": "script", "fence": "none"},
		MATHSF:          {"default": "none", "mi": "sans-serif"},
		MATHTT:          {"default": "monospace", "fence": "none"},
		MIT:             {"default": "italic", "fence": "none", "mi": "none"},
		OLDSTYLE:        {"default": "normal", "fence": "none"},
		SCR:             {"default": "script", "fence": "none"},
		TT:              {"default": "monospace", "fence": "none"},
	}

	OLD_STYLE_FONTS = map[string]map[string]string{
		`\rm`: {"default": "none", "mi": "normal"},
		`\bf`: {"default": "none", "mi": "bold"},
		`\it`: {"default": "none", "mi": "italic"},
		`\sf`: {"default": "none", "mi": "sans-serif"},
		`\tt`: {"default": "none", "mi": "monospace"},
	}

	GLOBAL_FONTS = map[string]map[string]string{
		`\rm`:   {"default": "none", "mi": "normal"},
		`\bf`:   {"default": "none", "mi": "bold"},
		`\it`:   {"default": "none", "mi": "italic"},
		`\sf`:   {"default": "none", "mi": "sans-serif"},
		`\tt`:   {"default": "none", "mi": "monospace"},
		`\cal`:  {"default": "script", "fence": "none"},
		`\frak`: {"default": "fraktur", "fence": "none"},
	}

	COMMANDS_WITH_ONE_PARAMETER = []string{
		ACUTE,
		BAR,
		BLACKBOARD_BOLD,
		BOLD_SYMBOL,
		BOXED,
		BREVE,
		CHECK,
		DOT,
		DDOT,
		DDDOT,
		DDDDOT,
		GRAVE,
		HAT,
		HPHANTOM,
		MATHRING,
		MIT,
		MOD,
		OLDSTYLE,
		OVERBRACE,
		OVERLEFTARROW,
		OVERLEFTRIGHTARROW,
		OVERLINE,
		OVERPAREN,
		OVERRIGHTARROW,
		PHANTOM,
		PMOD,
		SCR,
		TILDE,
		TT,
		UNDERBRACE,
		UNDERLEFTARROW,
		UNDERLINE,
		UNDERPAREN,
		UNDERRIGHTARROW,
		UNDERLEFTRIGHTARROW,
		VEC,
		VPHANTOM,
		WIDEHAT,
		WIDETILDE,
	}
	COMMANDS_WITH_TWO_PARAMETERS = []string{
		BINOM,
		CFRAC,
		DBINOM,
		DFRAC,
		FRAC,
		OVERSET,
		TBINOM,
		TFRAC,
		UNDERSET,
	}

	BIG_ORDER = []string{"minsize", "maxsize"}

	BIG = map[string]Style{
		`\Bigg`: {Tag: "mo", Modifiers: map[string]string{"minsize": "2.470em", "maxsize": "2.470em"}},
		`\bigg`: {Tag: "mo", Modifiers: map[string]string{"minsize": "2.047em", "maxsize": "2.047em"}},
		`\Big`:  {Tag: "mo", Modifiers: map[string]string{"minsize": "1.623em", "maxsize": "1.623em"}},
		`\big`:  {Tag: "mo", Modifiers: map[string]string{"minsize": "1.2em", "maxsize": "1.2em"}},
	}

	BIG_OPEN_CLOSE_ORDER = []string{"stretchy", "fence", "minsize", "maxsize"}

	BIG_OPEN_CLOSE = map[string]Style{}

	MSTYLE_SIZES = map[string]Style{
		`\Huge`:       {Tag: "mstyle", Modifiers: map[string]string{"mathsize": "2.49em"}},
		`\huge`:       {Tag: "mstyle", Modifiers: map[string]string{"mathsize": "2.07em"}},
		`\LARGE`:      {Tag: "mstyle", Modifiers: map[string]string{"mathsize": "1.73em"}},
		`\Large`:      {Tag: "mstyle", Modifiers: map[string]string{"mathsize": "1.44em"}},
		`\large`:      {Tag: "mstyle", Modifiers: map[string]string{"mathsize": "1.2em"}},
		`\normalsize`: {Tag: "mstyle", Modifiers: map[string]string{"mathsize": "1em"}},
		`\scriptsize`: {Tag: "mstyle", Modifiers: map[string]string{"mathsize": "0.7em"}},
		`\small`:      {Tag: "mstyle", Modifiers: map[string]string{"mathsize": "0.85em"}},
		`\tiny`:       {Tag: "mstyle", Modifiers: map[string]string{"mathsize": "0.5em"}},
		`\Tiny`:       {Tag: "mstyle", Modifiers: map[string]string{"mathsize": "0.6em"}},
	}

	STYLES = map[string]Style{
		DISPLAYSTYLE:      {Tag: "mstyle", Modifiers: map[string]string{"displaystyle": "true", "scriptlevel": "0"}},
		TEXTSTYLE:         {Tag: "mstyle", Modifiers: map[string]string{"displaystyle": "false", "scriptlevel": "0"}},
		SCRIPTSTYLE:       {Tag: "mstyle", Modifiers: map[string]string{"displaystyle": "false", "scriptlevel": "1"}},
		SCRIPTSCRIPTSTYLE: {Tag: "mstyle", Modifiers: map[string]string{"displaystyle": "false", "scriptlevel": "2"}},
	}

	DIACRITICS = map[string]Style{
		ACUTE:               {Tag: "&#x000B4;", Modifiers: map[string]string{}},
		BAR:                 {Tag: "&#x000AF;", Modifiers: map[string]string{"stretchy": "true"}},
		BREVE:               {Tag: "&#x002D8;", Modifiers: map[string]string{}},
		CHECK:               {Tag: "&#x002C7;", Modifiers: map[string]string{}},
		DOT:                 {Tag: "&#x002D9;", Modifiers: map[string]string{}},
		DDOT:                {Tag: "&#x000A8;", Modifiers: map[string]string{}},
		DDDOT:               {Tag: "&#x020DB;", Modifiers: map[string]string{}},
		DDDDOT:              {Tag: "&#x020DC;", Modifiers: map[string]string{}},
		GRAVE:               {Tag: "&#x00060;", Modifiers: map[string]string{}},
		HAT:                 {Tag: "&#x0005E;", Modifiers: map[string]string{"stretchy": "false"}},
		MATHRING:            {Tag: "&#x002DA;", Modifiers: map[string]string{}},
		OVERBRACE:           {Tag: "&#x23DE;", Modifiers: map[string]string{}},
		OVERLEFTARROW:       {Tag: "&#x02190;", Modifiers: map[string]string{}},
		OVERLEFTRIGHTARROW:  {Tag: "&#x02194;", Modifiers: map[string]string{}},
		OVERLINE:            {Tag: "&#x02015;", Modifiers: map[string]string{"accent": "true"}},
		OVERPAREN:           {Tag: "&#x023DC;", Modifiers: map[string]string{}},
		OVERRIGHTARROW:      {Tag: "&#x02192;", Modifiers: map[string]string{}},
		TILDE:               {Tag: "&#x0007E;", Modifiers: map[string]string{"stretchy": "false"}},
		UNDERBRACE:          {Tag: "&#x23DF;", Modifiers: map[string]string{}},
		UNDERLEFTARROW:      {Tag: "&#x02190;", Modifiers: map[string]string{}},
		UNDERLEFTRIGHTARROW: {Tag: "&#x02194;", Modifiers: map[string]string{}},
		UNDERLINE:           {Tag: "&#x02015;", Modifiers: map[string]string{"accent": "true"}},
		UNDERPAREN:          {Tag: "&#x023DD;", Modifiers: map[string]string{}},
		UNDERRIGHTARROW:     {Tag: "&#x02192;", Modifiers: map[string]string{}},
		VEC:                 {Tag: "&#x02192;", Modifiers: map[string]string{"stretchy": "true"}},
		WIDEHAT:             {Tag: "&#x0005E;", Modifiers: map[string]string{}},
		WIDETILDE:           {Tag: "&#x0007E;", Modifiers: map[string]string{}},
	}

	SIDE_ORDER   = []string{"stretchy", "fence", "form"}
	MIDDLE_ORDER = []string{"stretchy", "fence", "lspace", "rspace"}

	CONVERSION_MAP = map[string]Style{
		DISPLAYLINES:        {Tag: "mtable", Modifiers: map[string]string{"rowspacing": "0.5em", "columnspacing": "1em", "displaystyle": "true"}},
		SMALLMATRIX:         {Tag: "mtable", Modifiers: map[string]string{"rowspacing": "0.1em", "columnspacing": "0.2778em"}},
		SPLIT:               {Tag: "mtable", Modifiers: map[string]string{"displaystyle": "true", "columnspacing": "0em", "rowspacing": "3pt"}},
		ALIGN:               {Tag: "mtable", Modifiers: map[string]string{"displaystyle": "true", "rowspacing": "3pt"}},
		SUBSCRIPT:           {Tag: "msub", Modifiers: map[string]string{}},
		SUPERSCRIPT:         {Tag: "msup", Modifiers: map[string]string{}},
		SUBSUP:              {Tag: "msubsup", Modifiers: map[string]string{}},
		BINOM:               {Tag: "mfrac", Modifiers: map[string]string{"linethickness": "0"}},
		CFRAC:               {Tag: "mfrac", Modifiers: map[string]string{}},
		DBINOM:              {Tag: "mfrac", Modifiers: map[string]string{"linethickness": "0"}},
		DFRAC:               {Tag: "mfrac", Modifiers: map[string]string{}},
		FRAC:                {Tag: "mfrac", Modifiers: map[string]string{}},
		GENFRAC:             {Tag: "mfrac", Modifiers: map[string]string{}},
		TBINOM:              {Tag: "mfrac", Modifiers: map[string]string{"linethickness": "0"}},
		TFRAC:               {Tag: "mfrac", Modifiers: map[string]string{}},
		ACUTE:               {Tag: "mover", Modifiers: map[string]string{}},
		BAR:                 {Tag: "mover", Modifiers: map[string]string{}},
		BREVE:               {Tag: "mover", Modifiers: map[string]string{}},
		CHECK:               {Tag: "mover", Modifiers: map[string]string{}},
		DOT:                 {Tag: "mover", Modifiers: map[string]string{}},
		DDOT:                {Tag: "mover", Modifiers: map[string]string{}},
		DDDOT:               {Tag: "mover", Modifiers: map[string]string{}},
		DDDDOT:              {Tag: "mover", Modifiers: map[string]string{}},
		GRAVE:               {Tag: "mover", Modifiers: map[string]string{}},
		HAT:                 {Tag: "mover", Modifiers: map[string]string{}},
		LIMITS:              {Tag: "munderover", Modifiers: map[string]string{}},
		MATHRING:            {Tag: "mover", Modifiers: map[string]string{}},
		OVERBRACE:           {Tag: "mover", Modifiers: map[string]string{}},
		OVERLEFTARROW:       {Tag: "mover", Modifiers: map[string]string{}},
		OVERLEFTRIGHTARROW:  {Tag: "mover", Modifiers: map[string]string{}},
		OVERLINE:            {Tag: "mover", Modifiers: map[string]string{}},
		OVERPAREN:           {Tag: "mover", Modifiers: map[string]string{}},
		OVERRIGHTARROW:      {Tag: "mover", Modifiers: map[string]string{}},
		TILDE:               {Tag: "mover", Modifiers: map[string]string{}},
		OVERSET:             {Tag: "mover", Modifiers: map[string]string{}},
		UNDERBRACE:          {Tag: "munder", Modifiers: map[string]string{}},
		UNDERLEFTARROW:      {Tag: "munder", Modifiers: map[string]string{}},
		UNDERLINE:           {Tag: "munder", Modifiers: map[string]string{}},
		UNDERPAREN:          {Tag: "munder", Modifiers: map[string]string{}},
		UNDERRIGHTARROW:     {Tag: "munder", Modifiers: map[string]string{}},
		UNDERLEFTRIGHTARROW: {Tag: "munder", Modifiers: map[string]string{}},
		UNDERSET:            {Tag: "munder", Modifiers: map[string]string{}},
		VEC:                 {Tag: "mover", Modifiers: map[string]string{}},
		WIDEHAT:             {Tag: "mover", Modifiers: map[string]string{}},
		WIDETILDE:           {Tag: "mover", Modifiers: map[string]string{}},
		COLON:               {Tag: "mspace", Modifiers: map[string]string{"width": "0.222em"}},
		COMMA:               {Tag: "mspace", Modifiers: map[string]string{"width": "0.167em"}},
		DOUBLEBACKSLASH:     {Tag: "mspace", Modifiers: map[string]string{"linebreak": "newline"}},
		ENSPACE:             {Tag: "mspace", Modifiers: map[string]string{"width": "0.5em"}},
		EXCLAMATION:         {Tag: "mspace", Modifiers: map[string]string{"width": "negativethinmathspace"}},
		GREATER_THAN:        {Tag: "mspace", Modifiers: map[string]string{"width": "0.222em"}},
		HSKIP:               {Tag: "mspace", Modifiers: map[string]string{}},
		HSPACE:              {Tag: "mspace", Modifiers: map[string]string{}},
		KERN:                {Tag: "mspace", Modifiers: map[string]string{}},
		MKERN:               {Tag: "mspace", Modifiers: map[string]string{}},
		MSKIP:               {Tag: "mspace", Modifiers: map[string]string{}},
		MSPACE:              {Tag: "mspace", Modifiers: map[string]string{}},
		NEGTHINSPACE:        {Tag: "mspace", Modifiers: map[string]string{"width": "negativethinmathspace"}},
		NEGMEDSPACE:         {Tag: "mspace", Modifiers: map[string]string{"width": "negativemediummathspace"}},
		NEGTHICKSPACE:       {Tag: "mspace", Modifiers: map[string]string{"width": "negativethickmathspace"}},
		THINSPACE:           {Tag: "mspace", Modifiers: map[string]string{"width": "thinmathspace"}},
		QQUAD:               {Tag: "mspace", Modifiers: map[string]string{"width": "2em"}},
		QUAD:                {Tag: "mspace", Modifiers: map[string]string{"width": "1em"}},
		SEMICOLON:           {Tag: "mspace", Modifiers: map[string]string{"width": "0.278em"}},
		BOXED:               {Tag: "menclose", Modifiers: map[string]string{"notation": "box"}},
		FBOX:                {Tag: "menclose", Modifiers: map[string]string{"notation": "box"}},
		LEFT:                {Tag: "mo", Modifiers: map[string]string{"stretchy": "true", "fence": "true", "form": "prefix"}},
		MIDDLE:              {Tag: "mo", Modifiers: map[string]string{"stretchy": "true", "fence": "true", "lspace": "0.05em", "rspace": "0.05em"}},
		RIGHT:               {Tag: "mo", Modifiers: map[string]string{"stretchy": "true", "fence": "true", "form": "postfix"}},
		COLOR:               {Tag: "mstyle", Modifiers: map[string]string{}},
		SQRT:                {Tag: "msqrt", Modifiers: map[string]string{}},
		ROOT:                {Tag: "mroot", Modifiers: map[string]string{}},
		HREF:                {Tag: "mtext", Modifiers: map[string]string{}},
		TEXT:                {Tag: "mtext", Modifiers: map[string]string{}},
		TEXTBF:              {Tag: "mtext", Modifiers: map[string]string{"mathvariant": "bold"}},
		TEXTIT:              {Tag: "mtext", Modifiers: map[string]string{"mathvariant": "italic"}},
		TEXTRM:              {Tag: "mtext", Modifiers: map[string]string{}},
		TEXTSF:              {Tag: "mtext", Modifiers: map[string]string{"mathvariant": "sans-serif"}},
		TEXTTT:              {Tag: "mtext", Modifiers: map[string]string{"mathvariant": "monospace"}},
		HBOX:                {Tag: "mtext", Modifiers: map[string]string{}},
		MBOX:                {Tag: "mtext", Modifiers: map[string]string{}},
		HPHANTOM:            {Tag: "mphantom", Modifiers: map[string]string{}},
		PHANTOM:             {Tag: "mphantom", Modifiers: map[string]string{}},
		VPHANTOM:            {Tag: "mphantom", Modifiers: map[string]string{}},
		SIDESET:             {Tag: "mrow", Modifiers: map[string]string{}},
		SKEW:                {Tag: "mrow", Modifiers: map[string]string{}},
		MOD:                 {Tag: "mi", Modifiers: map[string]string{}},
		PMOD:                {Tag: "mi", Modifiers: map[string]string{}},
		BMOD:                {Tag: "mo", Modifiers: map[string]string{}},
		XLEFTARROW:          {Tag: "mover", Modifiers: map[string]string{}},
		XRIGHTARROW:         {Tag: "mover", Modifiers: map[string]string{}},
	}
)

func InitializeCommands() {
	if !COMMANDS_INITILIAZED {
		for _, postfix := range "lmr" {
			for command, style := range BIG {
				BIG_OPEN_CLOSE[command+string(postfix)] = Style{
					Tag: style.Tag,
					Modifiers: map[string]string{
						"stretchy": "true",
						"fence":    "true",
						"minsize":  style.Modifiers["minsize"],
						"maxsize":  style.Modifiers["maxsize"],
					},
				}

			}
		}

		for _, matrix := range MATRICES {
			CONVERSION_MAP[matrix] = Style{Tag: "mtable", Modifiers: map[string]string{}}
		}

		for _, limit := range LIMIT {
			CONVERSION_MAP[limit] = Style{Tag: "mo", Modifiers: map[string]string{}}
		}

		for _, overload := range []map[string]Style{BIG, BIG_OPEN_CLOSE, MSTYLE_SIZES, STYLES} {
			for command, style := range overload {
				CONVERSION_MAP[command] = style
			}
		}

		COMMANDS_INITILIAZED = true
	}
}
