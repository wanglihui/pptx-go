package pptx

import (
	"fmt"
	"io/ioutil"
	"net/http"
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

// SetText 设置占位符的文本内容，支持普通文本、LaTeX 公式和超链接
func (p *Placeholder) SetText(text string, options ...TextOption) error {
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

	// 应用选项
	opts := &TextOptions{}
	for _, option := range options {
		option(opts)
	}

	if opts.EnableLatex {
		// LaTeX 处理逻辑
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
			} else {
				p.addTextRun(para, segment.Text, opts)
			}
		}
	} else {
		// 普通文本处理
		p.addTextRun(para, text, opts)
	}

	return p.slide.SaveChanges()
}

// TextOptions 定义文本设置的选项
type TextOptions struct {
	EnableLatex bool
	Link        string   // 超链接URL
	LinkType    LinkType // 超链接类型
}

// LinkType 定义超链接类型
type LinkType int

const (
	LinkTypeExternal LinkType = iota // 外部链接
	LinkTypeSlide                    // 幻灯片内部链接
)

// TextOption 定义文本设置的选项函数
type TextOption func(*TextOptions)

// WithLatex 启用 LaTeX 支持
func WithLatex() TextOption {
	return func(o *TextOptions) {
		o.EnableLatex = true
	}
}

// WithLink 添加超链接
func WithLink(url string, linkType LinkType) TextOption {
	return func(o *TextOptions) {
		o.Link = url
		o.LinkType = linkType
	}
}

// addTextRun 添加文本运行
func (p *Placeholder) addTextRun(para *etree.Element, text string, opts *TextOptions) {
	run := para.CreateElement("a:r")

	// 添加运行属性
	rPr := run.CreateElement("a:rPr")
	rPr.CreateAttr("lang", "en-US")
	rPr.CreateAttr("sz", "2400")

	// 添加文本元素
	t := run.CreateElement("a:t")
	t.SetText(text)

	// 如果有超链接，添加超链接
	if opts != nil && opts.Link != "" {
		// 获取 slide 引用
		slide := p.slide
		if slide != nil {
			// 创建关系ID
			rId := fmt.Sprintf("rId%d", len(slide.rels)+1)

			// 添加超链接元素
			hlinkClick := rPr.CreateElement("a:hlinkClick")
			hlinkClick.CreateAttr("r:id", rId)

			// 更新关系
			if err := p.updateHyperlinkRelationship(rId, opts.Link, opts.LinkType); err != nil {
				// 在际应用中，你可能需要更好的错误处理
				fmt.Printf("Failed to update hyperlink relationship: %v\n", err)
			}
		}
	}
}

// updateHyperlinkRelationship 更新超链接关系
func (p *Placeholder) updateHyperlinkRelationship(rId, target string, linkType LinkType) error {
	// 获取 slide 引用
	slide := p.slide
	if slide == nil {
		return fmt.Errorf("could not find slide reference")
	}

	// 创建关系
	rel := &Relationship{
		Id:     rId,
		Target: target,
	}

	if linkType == LinkTypeExternal {
		rel.Type = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink"
		rel.TargetMode = "External"
	} else {
		rel.Type = "http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide"
	}

	// 添加到幻灯片关系中
	slide.rels[rId] = rel

	return nil
}

// SetImage 设置占位符的图片，支持本地文件路径和网络URL
func (p *Placeholder) SetImage(imagePath string) error {
	var imageData []byte
	var err error

	// 检查是否为网络URL
	if strings.HasPrefix(imagePath, "http://") || strings.HasPrefix(imagePath, "https://") {
		// 下载网络图片
		imageData, err = downloadImage(imagePath)
		if err != nil {
			return fmt.Errorf("failed to download image: %w", err)
		}
	} else {
		// 读取本地图片文件
		imageData, err = ioutil.ReadFile(imagePath)
		if err != nil {
			return fmt.Errorf("failed to read image file: %w", err)
		}
	}

	// 保存图片到pptx文件
	imgExt := strings.ToLower(filepath.Ext(imagePath))
	if imgExt == "" {
		// 如果是网络URL没有扩展名，默认使用.png
		imgExt = ".png"
	}
	fmt.Println("p.slide.rels===>", p.slide.rels)
	imgPath := fmt.Sprintf("ppt/media/image%d%s", len(p.slide.rels)+1, imgExt)

	// 创建关系ID
	rId := fmt.Sprintf("rId%d", len(p.slide.rels)+1)

	// 更新关系文件
	if err := p.updateImageRelationship(rId, filepath.Base(imgPath)); err != nil {
		return err
	}

	// 创建新的 p:pic 元素
	pic := etree.NewElement("p:pic")

	// 添加 nvPicPr (non-visual picture properties)
	nvPicPr := pic.CreateElement("p:nvPicPr")
	cNvPr := nvPicPr.CreateElement("p:cNvPr")
	// cNvPr.CreateAttr("id", "1")
	cNvPr.CreateAttr("name", "Picture")

	cNvPicPr := nvPicPr.CreateElement("p:cNvPicPr")
	cNvPicPr.CreateElement("a:picLocks").CreateAttr("noChangeAspect", "1")

	nvPr := nvPicPr.CreateElement("p:nvPr")
	ph := nvPr.CreateElement("p:ph")
	ph.CreateAttr("type", "pic")

	// 添加 blipFill
	blipFill := pic.CreateElement("p:blipFill")
	blip := blipFill.CreateElement("a:blip")
	blip.CreateAttr("r:embed", rId)

	stretch := blipFill.CreateElement("a:stretch")
	stretch.CreateElement("a:fillRect")

	// 复制原占位符的 spPr (shape properties)
	spPr := pic.CreateElement("p:spPr")
	if originalSpPr := p.Shape.FindElement("p:spPr"); originalSpPr != nil {
		// 复制变换信息
		if xfrm := originalSpPr.FindElement("a:xfrm"); xfrm != nil {
			newXfrm := spPr.CreateElement("a:xfrm")
			// 复制所有属性
			for _, attr := range xfrm.Attr {
				newXfrm.CreateAttr(attr.Key, attr.Value)
			}
			// 复制位置和大小元素
			for _, child := range xfrm.ChildElements() {
				newChild := newXfrm.CreateElement(child.Tag)
				for _, attr := range child.Attr {
					newChild.CreateAttr(attr.Key, attr.Value)
				}
			}
		}

		// 添加预设形状
		prstGeom := spPr.CreateElement("a:prstGeom")
		prstGeom.CreateAttr("prst", "rect")
		prstGeom.CreateElement("a:avLst")
	}

	// 替换原占位符内容
	parent := p.Shape.Parent()
	if parent == nil {
		return fmt.Errorf("placeholder parent element not found")
	}

	// 移除原占位符
	parent.RemoveChild(p.Shape)

	// 添加新的 pic 元素
	parent.AddChild(pic)

	// 更新 Shape 引用
	p.Shape = pic

	// 保存图片数据
	p.slide.pres.files[imgPath] = imageData

	// 保存幻灯片更改
	return p.slide.SaveChanges()
}

// downloadImage 下载网络图片
func downloadImage(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image, status code: %d", resp.StatusCode)
	}

	return ioutil.ReadAll(resp.Body)
}

// updateImageRelationship 更新图片关系
func (p *Placeholder) updateImageRelationship(rId, imagePath string) error {
	// 创建或更新关系文件
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
	p.slide.rels[rId] = &Relationship{
		Id:     rId,
		Type:   "http://schemas.openxmlformats.org/officeDocument/2006/relationships/image",
		Target: "../media/" + filepath.Base(imagePath),
	}

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

// SaveChanges 保存对幻灯片的更改
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

// Relationship 定义了 PPTX 中的关系
type Relationship struct {
	Id         string
	Type       string
	Target     string
	TargetMode string
}
