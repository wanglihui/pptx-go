package pptx

import (
	"fmt"
	"strconv"

	"github.com/beevik/etree"
)

// Slide 表示一个幻灯片
type Slide struct {
	xml      *etree.Document
	path     string
	relsPath string
	layout   *Layout
	master   *Master
	pres     *Presentation
	rels     map[string]*Relationship
}

// SlideSize 表示幻灯片大小
type SlideSize struct {
	Width  int
	Height int
	Type   string
}

// copyPlaceholdersFromLayout 从布局复制占位符到幻灯片
func (s *Slide) copyPlaceholdersFromLayout() error {
	if s.layout == nil || s.layout.xml == nil {
		return fmt.Errorf("layout not available")
	}

	layoutSpTree := s.layout.xml.FindElement("//p:spTree")
	if layoutSpTree == nil {
		return fmt.Errorf("layout spTree not found")
	}

	slideSpTree := s.xml.FindElement("//p:spTree")
	if slideSpTree == nil {
		return fmt.Errorf("slide spTree not found")
	}

	// 复制所有形状
	for _, sp := range layoutSpTree.SelectElements("p:sp") {
		// 创建新的形状元素
		newSp := slideSpTree.CreateElement("p:sp")

		// 复制属性和子元素
		for _, attr := range sp.Attr {
			newSp.CreateAttr(attr.Key, attr.Value)
		}

		// 深度复制内容
		copyXMLElement(sp, newSp)
	}

	return nil
}

// copyXMLElement 递归复制XML元素
func copyXMLElement(src, dst *etree.Element) {
	for _, child := range src.Child {
		switch v := child.(type) {
		case *etree.Element:
			newChild := dst.CreateElement(v.Tag)
			// 复制属性
			for _, attr := range v.Attr {
				newChild.CreateAttr(attr.Key, attr.Value)
			}
			// 递归复制子元素
			copyXMLElement(v, newChild)
		case *etree.CharData:
			dst.CreateText(v.Data)
		}
	}
}

// atoi 将字符串转换为整数，忽略错误
func atoi(s string) int {
	var n int
	fmt.Sscanf(s, "%d", &n)
	return n
}

// GetPlaceholder 通过类型、名称、索引或文本内容获取占位符
func (s *Slide) GetPlaceholder(params ...interface{}) (*Placeholder, error) {
	if len(params) == 0 {
		return nil, fmt.Errorf("no parameters provided")
	}
	// fmt.Println(s.xml.WriteToString())
	// 查找 spTree，不使用命名空间前缀
	spTree := s.xml.FindElement("//sld/cSld/spTree")
	if spTree == nil {
		// 尝试带命名空间的路径
		spTree = s.xml.FindElement("//p:sld/p:cSld/p:spTree")
		if spTree == nil {
			return nil, fmt.Errorf("shape tree not found in slide")
		}
	}

	// 查找所有 sp 元素，不使用命名空间前缀
	for _, sp := range spTree.SelectElements("sp") {
		nvSpPr := sp.FindElement("nvSpPr")
		if nvSpPr == nil {
			continue
		}

		// 获取占位符信息，尝试不带命名空间的路径
		ph := nvSpPr.FindElement("nvPr/ph")
		if ph == nil {
			// 尝试带命名空间的路径
			ph = nvSpPr.FindElement("p:nvPr/p:ph")
			if ph == nil {
				continue
			}
		}

		placeholder := &Placeholder{
			Shape: sp,
			slide: s,
		}

		// 获取占位符类型
		if typeAttr := ph.SelectAttr("type"); typeAttr != nil {
			placeholder.Type = parsePlaceholderType(typeAttr.Value)
		}

		// 获取占位符名称
		nvPr := nvSpPr.FindElement("nvPr")
		if nvPr == nil {
			nvPr = nvSpPr.FindElement("p:nvPr")
		}
		if nvPr != nil {
			if userDrawn := nvPr.SelectAttr("userDrawn"); userDrawn != nil && userDrawn.Value == "1" {
				continue // 跳过用户绘制的形状
			}
			if name := nvPr.SelectAttrValue("name", ""); name != "" {
				placeholder.Name = name
			}
		}

		// 获取占位符索引
		if idxAttr := ph.SelectAttr("idx"); idxAttr != nil {
			placeholder.Index, _ = strconv.Atoi(idxAttr.Value)
		}

		// 获取占位符文本内容
		var placeholderText string
		if txBody := sp.FindElement("txBody"); txBody != nil {
			for _, p := range txBody.SelectElements("a:p") {
				for _, r := range p.SelectElements("a:r") {
					if t := r.SelectElement("a:t"); t != nil {
						placeholderText += t.Text()
					}
				}
			}
		}

		// 根据参数匹配占位符
		for _, param := range params {
			switch v := param.(type) {
			case PlaceholderType:
				if placeholder.Type == v {
					return placeholder, nil
				}
			case string:
				// 匹配名称或文本内容
				if placeholder.Name == v || placeholderText == v {
					return placeholder, nil
				}
			case int:
				if placeholder.Index == v {
					return placeholder, nil
				}
			case struct {
				Type PlaceholderType
				Text string
			}:
				// 同时匹配类型和文本
				if placeholder.Type == v.Type && placeholderText == v.Text {
					return placeholder, nil
				}
			case struct {
				Name string
				Text string
			}:
				// 同时匹配名称和文本
				if placeholder.Name == v.Name && placeholderText == v.Text {
					return placeholder, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("placeholder not found")
}

// copyLayoutRelationships 复制布局中的关系到新幻灯片
func (s *Slide) copyLayoutRelationships(layout *Layout) error {
	// 读取布局的关系文件
	layoutRelsData, exists := s.pres.files[layout.relsPath]
	if !exists {
		return nil // 如果布局没有关系文件，直接返回
	}

	// 解析布局关系文件
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(layoutRelsData); err != nil {
		return fmt.Errorf("failed to parse layout relationships: %w", err)
	}

	// 遍历所有关系
	for _, rel := range doc.FindElements("//Relationship") {
		rId := rel.SelectAttrValue("Id", "")
		target := rel.SelectAttrValue("Target", "")
		relType := rel.SelectAttrValue("Type", "")
		targetMode := rel.SelectAttrValue("TargetMode", "")

		// 创建新的关系ID
		newRId := fmt.Sprintf("rId%d", len(s.rels)+1)

		// 添加到幻灯片的关系中
		s.rels[newRId] = &Relationship{
			Id:         newRId,
			Type:       relType,
			Target:     target,
			TargetMode: targetMode,
		}

		// 更新幻灯片XML中的rId引用
		s.updateReferenceIds(rId, newRId)
	}

	return nil
}

// updateReferenceIds 更新幻灯片XML中的rId引用
func (s *Slide) updateReferenceIds(oldRId, newRId string) {
	// 查找所有包含 r:id 属性的元素
	for _, elem := range s.xml.FindElements("//*[@r:id]") {
		if id := elem.SelectAttrValue("r:id", ""); id == oldRId {
			elem.RemoveAttr("r:id")
			elem.CreateAttr("r:id", newRId)
		}
	}

	// 查找所有包含 r:embed 属性的元素
	for _, elem := range s.xml.FindElements("//*[@r:embed]") {
		if id := elem.SelectAttrValue("r:embed", ""); id == oldRId {
			elem.RemoveAttr("r:embed")
			elem.CreateAttr("r:embed", newRId)
		}
	}

	// 查找所有包含 r:link 属性的元素
	for _, elem := range s.xml.FindElements("//*[@r:link]") {
		if id := elem.SelectAttrValue("r:link", ""); id == oldRId {
			elem.RemoveAttr("r:link")
			elem.CreateAttr("r:link", newRId)
		}
	}
}
