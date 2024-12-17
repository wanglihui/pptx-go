package pptx

import (
	"fmt"
	"strings"

	_ "embed"

	"github.com/beevik/etree"
	xslt "github.com/wamuir/go-xslt"
	"github.com/wanglihui/pptx-go/latex2mathml"
)

// 初始化 XSLT 转换器
var mathmlToOmmlXslt *xslt.Stylesheet

//go:embed MATH2OMML.xsl
var xsltContent []byte

func init() {
	var err error
	// 读取本地 XSLT 文件
	mathmlToOmmlXslt, err = xslt.NewStylesheet(xsltContent)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse XSLT stylesheet: %v", err))
	}
	// defer mathmlToOmmlXslt.Close()
}

// convertLatexToMathML 使用 latex2mathml 将 LaTeX 转换为 MathML
func convertLatexToMathML(latex string) (string, error) {
	// 创建 latex2mathml 命令
	mathml := latex2mathml.Convert(latex, "http://www.w3.org/1998/Math/MathML", "inline", 0)
	return mathml, nil
}

// convertLatexToOMML 将 LaTeX 转换为 OMML
func convertLatexToOMML(latex string) ([]*etree.Element, error) {
	// fmt.Println("latex===>", latex)
	// 1. 使用 latex2mathml 将 LaTeX 转换为 MathML
	mathml, err := convertLatexToMathML(latex)
	if err != nil {
		return nil, fmt.Errorf("failed to convert LaTeX to MathML: %w", err)
	}
	// fmt.Println("mathml===>", mathml)
	// 2. 使用 XSLT 将 MathML 转换为 OMML
	omml, err := convertMathMLToOMMLUsingXSLT(mathml)
	if err != nil {
		return nil, fmt.Errorf("failed to convert MathML to OMML: %w", err)
	}
	// fmt.Println("omml===>", omml)
	// 3. 解析 OMML 为 etree.Element
	doc := etree.NewDocument()
	if err := doc.ReadFromString(omml); err != nil {
		return nil, fmt.Errorf("failed to parse OMML: %w", err)
	}

	// 4. 创建正确的 PPTX 数学公式结构
	if root := doc.Root(); root != nil {
		// 创建 m:oMathPara 元素包装
		oMathPara := etree.NewElement("m:oMathPara")
		oMathPara.CreateAttr("xmlns:m", "http://schemas.openxmlformats.org/officeDocument/2006/math")

		// 添加 m:oMathParaPr 元素
		oMathParaPr := etree.NewElement("m:oMathParaPr")
		oMathPara.AddChild(oMathParaPr)

		oMath := etree.NewElement("m:oMath")
		oMath.CreateAttr("xmlns:m", "http://schemas.openxmlformats.org/officeDocument/2006/math")

		// 将转换后的数学公式元素添加到 oMath 中
		for _, child := range root.Child {
			if elem, ok := child.(*etree.Element); ok {
				oMath.AddChild(elem.Copy())
			}
		}

		oMathPara.AddChild(oMath)
		return []*etree.Element{oMathPara}, nil
	}

	return nil, fmt.Errorf("no elements found in OMML")
}

// convertMathMLToOMMLUsingXSLT 使用 XSLT 将 MathML 转换为 OMML
func convertMathMLToOMMLUsingXSLT(mathml string) (string, error) {

	// 应用 XSLT 转换
	result, err := mathmlToOmmlXslt.Transform([]byte(mathml))
	if err != nil {
		return "", fmt.Errorf("failed to apply XSLT transformation: %w", err)
	}
	return string(result), nil
}

// TextSegment 表示文本片段，可以是普通文本或LaTeX公式
type TextSegment struct {
	Text    string
	IsLatex bool
}

// parseLatexFormula 解析文本中的LaTeX公式
func parseLatexFormula(text string) []TextSegment {
	var segments []TextSegment
	var currentText strings.Builder
	inFormula := false

	// 遍历文本的每个字符
	for i := 0; i < len(text); i++ {
		if text[i] == '$' {
			// 检查是否为转义的 $
			if i > 0 && text[i-1] == '\\' {
				currentText.WriteByte('$')
				continue
			}

			// 处理当前累积的文本
			if currentText.Len() > 0 {
				segments = append(segments, TextSegment{
					Text:    currentText.String(),
					IsLatex: inFormula,
				})
				currentText.Reset()
			}

			// 切换公式状态
			inFormula = !inFormula
			continue
		}

		// 累积当前字符
		currentText.WriteByte(text[i])
	}

	// 处理最后剩余的文本
	if currentText.Len() > 0 {
		segments = append(segments, TextSegment{
			Text:    currentText.String(),
			IsLatex: inFormula,
		})
	}

	// 如果最后还在公式状态，说明式没有正确闭合
	if inFormula {
		// 将最后一个片段标记为普通文本
		if len(segments) > 0 {
			lastSegment := segments[len(segments)-1]
			segments[len(segments)-1] = TextSegment{
				Text:    "$" + lastSegment.Text,
				IsLatex: false,
			}
		}
	}

	return segments
}
