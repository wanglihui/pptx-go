package pptx

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/beevik/etree"
)

// PlaceholderType 定义占位符类型
type PlaceholderType int

const (
	PlaceholderTitle PlaceholderType = iota
	PlaceholderSubTile
	PlaceholderCrtTitle
	PlaceholderBody
	PlaceholderImage
	PlaceholderChart
	PlaceholderTable
	PlaceholderShape
	PlaceholderFooter
	PlaceholderHeader
	PlaceholderSlideNumber
	PlaceholderDate
)

// Placeholder 表示幻灯片中的占位符
type Placeholder struct {
	Type  PlaceholderType
	Name  string
	Index int
	Shape *etree.Element
	slide *Slide
}

// SetText 设置占位符的文本内容
func (p *Placeholder) SetText(text string) error {
	if p.Shape == nil {
		return fmt.Errorf("shape element is nil")
	}

	// 查找或创建 txBody
	txBody := p.Shape.FindElement("txBody")
	if txBody == nil {
		txBody = p.Shape.CreateElement("txBody")
	}

	// 清除现有文本
	for _, a := range txBody.SelectElements("a:p") {
		txBody.RemoveChild(a)
	}

	// 创建新的段落
	para := txBody.CreateElement("a:p")
	run := para.CreateElement("a:r")
	textElement := run.CreateElement("a:t")
	textElement.SetText(text)

	// 确保 slide 引用存在
	if p.slide == nil {
		return fmt.Errorf("slide reference is nil")
	}

	// 保存更改
	if err := p.slide.SaveChanges(); err != nil {
		return fmt.Errorf("failed to save changes: %w", err)
	}
	// fmt.Println(p.slide.xml.WriteToString())
	return nil
}

// SetText 设置占位符的文本内容，支持 LaTeX 公式
func (p *Placeholder) SetTextWithLatex(text string) error {
	if p.Shape == nil {
		return fmt.Errorf("shape element is nil")
	}

	// 查找或创建 txBody
	txBody := p.Shape.FindElement("p:txBody")
	if txBody == nil {
		txBody = p.Shape.CreateElement("p:txBody")
	}

	// 清除现有文本
	for _, a := range txBody.SelectElements("a:p") {
		txBody.RemoveChild(a)
	}

	// 创建新的段落
	para := txBody.CreateElement("a:p")

	// 解析文本中的 LaTeX 公式
	segments := parseLatexFormula(text)
	for _, segment := range segments {
		if segment.IsLatex {
			// 转换 LaTeX 为 OMML
			ommlElements, err := convertLatexToOMML(segment.Text)
			if err != nil {
				return fmt.Errorf("failed to convert LaTeX to OMML: %w", err)
			}
			for _, elem := range ommlElements {
				// 添加 a14:m 容器
				mathContainer := etree.NewElement("a14:m")
				mathContainer.CreateAttr("xmlns:a14", "http://schemas.microsoft.com/office/drawing/2010/main")
				mathContainer.AddChild(elem)
				para.AddChild(mathContainer)
			}
			// 使用 wrapOMMLInRun 包装 OMML 元素并添加到段落中
			// run := wrapOMMLInRun(ommlElements)
			// para.AddChild(run)
		} else {
			// 普通文本处理
			run := para.CreateElement("a:r")
			textElement := run.CreateElement("a:t")
			textElement.SetText(segment.Text)
		}
	}

	// 保存更改
	return p.slide.SaveChanges()
}

// SetImage 设置占位符的图片
func (p *Placeholder) SetImage(imagePath string) error {
	// 读取图片文件
	imageData, err := ioutil.ReadFile(imagePath)
	if err != nil {
		return fmt.Errorf("failed to read image file: %w", err)
	}

	// 创建关系ID
	rId := fmt.Sprintf("rId%d", len(p.slide.rels)+1)

	// 更新关系文件
	if err := p.updateImageRelationship(rId, imagePath); err != nil {
		return err
	}

	// 更新形状内容
	blipFill := p.Shape.FindElement("blipFill")
	if blipFill == nil {
		blipFill = p.Shape.CreateElement("blipFill")
	}

	blip := blipFill.CreateElement("a:blip")
	blip.CreateAttr("r:embed", rId)

	// 保存图片到pptx文件
	imgExt := strings.ToLower(filepath.Ext(imagePath))
	imgPath := fmt.Sprintf("ppt/media/image%d%s", len(p.slide.rels)+1, imgExt)
	p.slide.pres.files[imgPath] = imageData

	// 保存幻灯片更改
	return p.slide.SaveChanges()
}

// updateImageRelationship 更新图片关系
func (p *Placeholder) updateImageRelationship(rId, imagePath string) error {
	// 创���或更新关系文件
	relsDoc := etree.NewDocument()
	relationships := relsDoc.CreateElement("Relationships")
	relationships.CreateAttr("xmlns", NsRelationships)

	rel := relationships.CreateElement("Relationship")
	rel.CreateAttr("Id", rId)
	rel.CreateAttr("Type", "http://schemas.openxmlformats.org/officeDocument/2006/relationships/image")
	rel.CreateAttr("Target", "../media/"+filepath.Base(imagePath))

	// 保存关系文件
	data, err := relsDoc.WriteToBytes()
	if err != nil {
		return fmt.Errorf("failed to serialize relationships: %w", err)
	}

	p.slide.pres.files[p.slide.relsPath] = data // 使用presentation的files
	p.slide.rels[rId] = filepath.Base(imagePath)

	return nil
}

// SetTable 设置占位符的表格内容
func (p *Placeholder) SetTable(data [][]string) error {
	graphicFrame := p.Shape.FindElement("graphicFrame")
	if graphicFrame == nil {
		return fmt.Errorf("graphic frame not found")
	}

	// 创建表格
	tbl := graphicFrame.CreateElement("a:tbl")

	// 设置表格网格
	tblGrid := tbl.CreateElement("a:tblGrid")
	for range data[0] {
		tblGrid.CreateElement("a:gridCol")
	}

	// 添加行和单元格
	for _, row := range data {
		tr := tbl.CreateElement("a:tr")
		for _, cell := range row {
			tc := tr.CreateElement("a:tc")
			txBody := tc.CreateElement("a:txBody")
			p := txBody.CreateElement("a:p")
			r := p.CreateElement("a:r")
			t := r.CreateElement("a:t")
			t.SetText(cell)
		}
	}

	// 保存更改
	return p.slide.SaveChanges()
}

// parsePlaceholderType 解析占位符类型
func parsePlaceholderType(typeStr string) PlaceholderType {
	switch typeStr {
	case "title":
		return PlaceholderTitle
	case "body":
		return PlaceholderBody
	case "pic":
		return PlaceholderImage
	case "chart":
		return PlaceholderChart
	case "tbl":
		return PlaceholderTable
	case "ftr":
		return PlaceholderFooter
	case "hdr":
		return PlaceholderHeader
	case "sldNum":
		return PlaceholderSlideNumber
	case "dt":
		return PlaceholderDate
	default:
		return PlaceholderShape
	}
}

// GetPlaceholders 获取幻灯片中的所有占位符
func (s *Slide) GetPlaceholders() ([]*Placeholder, error) {
	var placeholders []*Placeholder

	spTree := s.xml.FindElement("//p:spTree")
	if spTree == nil {
		return nil, fmt.Errorf("shape tree not found in slide")
	}

	for _, sp := range spTree.SelectElements("p:sp") {
		nvSpPr := sp.FindElement("p:nvSpPr")
		if nvSpPr == nil {
			continue
		}

		ph := nvSpPr.FindElement(".//p:ph")
		if ph == nil {
			continue
		}

		placeholder := &Placeholder{
			Shape: sp,
			slide: s,
		}

		// 设置占位符属性
		if typeAttr := ph.SelectAttr("type"); typeAttr != nil {
			placeholder.Type = parsePlaceholderType(typeAttr.Value)
		}

		if nvPr := nvSpPr.FindElement("p:nvPr"); nvPr != nil {
			if name := nvPr.SelectAttrValue("name", ""); name != "" {
				placeholder.Name = name
			}
		}

		if idxAttr := ph.SelectAttr("idx"); idxAttr != nil {
			placeholder.Index, _ = strconv.Atoi(idxAttr.Value)
		}

		placeholders = append(placeholders, placeholder)
	}

	return placeholders, nil
}

// SaveChanges 保存对幻灯片的修改
func (s *Slide) SaveChanges() error {
	if s.xml != nil {
		data, err := s.xml.WriteToBytes()
		if err != nil {
			return fmt.Errorf("failed to serialize slide XML: %w", err)
		}
		s.pres.files[s.path] = data
		// fmt.Println("data===>", string(data))
	}
	return nil
}

// Debug 打印调试信息
func (p *Placeholder) Debug() {
	fmt.Printf("Placeholder Debug Info:\n")
	fmt.Printf("Type: %v\n", p.Type)
	fmt.Printf("Name: %s\n", p.Name)
	fmt.Printf("Index: %d\n", p.Index)
	fmt.Printf("Shape is nil: %v\n", p.Shape == nil)
	fmt.Printf("Slide is nil: %v\n", p.slide == nil)
	if p.slide != nil {
		fmt.Printf("Slide path: %s\n", p.slide.path)
		fmt.Printf("Presentation is nil: %v\n", p.slide.pres == nil)
	}
}
