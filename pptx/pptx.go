package pptx

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/beevik/etree"
)

const (
	// PPTX 相关的命名空间
	NsRelationships  = "http://schemas.openxmlformats.org/package/2006/relationships"
	NsPresentationML = "http://schemas.openxmlformats.org/presentationml/2006/main"
	NsDrawingML      = "http://schemas.openxmlformats.org/drawingml/2006/main"
)

// Presentation 表示一个PPTX文件
type Presentation struct {
	zipReader *zip.ReadCloser
	files     map[string][]byte
	slides    []*Slide
	masters   []*Master
	rels      map[string]string // 关系ID到目标的映射
}

// Layout 表示幻灯片布局
type Layout struct {
	name     string
	xml      *etree.Document
	path     string
	relsPath string
	rels     map[string]string
}

// Master 表示母版
type Master struct {
	name     string
	xml      *etree.Document
	path     string
	relsPath string
	rels     map[string]string
	layouts  []*Layout
}

// Open 打开PPTX文件
func Open(filename string) (*Presentation, error) {
	reader, err := zip.OpenReader(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open pptx file: %w", err)
	}

	pptx := &Presentation{
		zipReader: reader,
		files:     make(map[string][]byte),
		rels:      make(map[string]string),
	}

	// 读取zip文件中的所有内容
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open zip file entry %s: %w", file.Name, err)
		}

		content, err := io.ReadAll(rc)
		if err != nil {
			rc.Close()
			return nil, fmt.Errorf("failed to read zip file entry %s: %w", file.Name, err)
		}
		rc.Close()
		// fmt.Println("file.Name=====>", file.Name)
		pptx.files[file.Name] = content
	}

	// 初始化presentation
	err = pptx.initialize()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize presentation: %w", err)
	}

	return pptx, nil
}

// initialize 初始化presentation
func (p *Presentation) initialize() error {
	// 首先解析presentation.xml.rels
	relsContent, ok := p.files["ppt/_rels/presentation.xml.rels"]
	if !ok {
		return errors.New("presentation.xml.rels not found")
	}

	relsDoc := etree.NewDocument()
	if err := relsDoc.ReadFromBytes(relsContent); err != nil {
		return fmt.Errorf("failed to parse presentation.xml.rels: %w", err)
	}

	// 解析关系
	for _, rel := range relsDoc.FindElements("//Relationship") {
		id := rel.SelectAttr("Id").Value
		target := rel.SelectAttr("Target").Value
		p.rels[id] = target
	}

	// 解析presentation.xml
	presContent, ok := p.files["ppt/presentation.xml"]
	if !ok {
		return errors.New("presentation.xml not found")
	}

	presDoc := etree.NewDocument()
	if err := presDoc.ReadFromBytes(presContent); err != nil {
		return fmt.Errorf("failed to parse presentation.xml: %w", err)
	}

	// 初始化masters
	if err := p.initMasters(presDoc); err != nil {
		return fmt.Errorf("failed to initialize masters: %w", err)
	}

	// 初始化slides
	if err := p.initSlides(presDoc); err != nil {
		return fmt.Errorf("failed to initialize slides: %w", err)
	}

	return nil
}

// initMasters 初始化母版
func (p *Presentation) initMasters(presDoc *etree.Document) error {
	p.masters = make([]*Master, 0)

	// 查找所有sldMasterId元素
	for _, masterEl := range presDoc.FindElements("//p:sldMasterId") {
		rId := masterEl.SelectAttr("r:id").Value
		masterPath := "ppt/" + p.rels[rId]

		masterContent, ok := p.files[masterPath]
		if !ok {
			return fmt.Errorf("master file not found: %s", masterPath)
		}

		master := &Master{
			path:    masterPath,
			rels:    make(map[string]string),
			layouts: make([]*Layout, 0),
		}

		// 解析master XML
		masterDoc := etree.NewDocument()
		if err := masterDoc.ReadFromBytes(masterContent); err != nil {
			return fmt.Errorf("failed to parse master file: %w", err)
		}
		master.xml = masterDoc

		// 解析master关系文件
		masterRelsPath := filepath.Join("ppt/slideMasters/_rels", filepath.Base(masterPath)+".rels")
		if relsContent, ok := p.files[masterRelsPath]; ok {
			master.relsPath = masterRelsPath
			relsDoc := etree.NewDocument()
			if err := relsDoc.ReadFromBytes(relsContent); err != nil {
				return fmt.Errorf("failed to parse master rels: %w", err)
			}

			// 解析布局关系
			for _, rel := range relsDoc.FindElements("//Relationship") {
				id := rel.SelectAttr("Id").Value
				target := rel.SelectAttr("Target").Value
				master.rels[id] = target

				// 如果是布局，则加载布局
				if strings.Contains(target, "slideLayout") {
					target = strings.Replace(target, "../", "", 1)
					layout, err := p.loadLayout(filepath.Join("ppt", target))
					if err != nil {
						return fmt.Errorf("failed to load layout: %w", err)
					}
					master.layouts = append(master.layouts, layout)
				}
			}
		}

		p.masters = append(p.masters, master)
	}

	return nil
}

// loadLayout 加载布局文件
func (p *Presentation) loadLayout(layoutPath string) (*Layout, error) {
	layoutContent, ok := p.files[layoutPath]
	if !ok {
		return nil, fmt.Errorf("layout file not found: %s", layoutPath)
	}

	layout := &Layout{
		path: layoutPath,
		rels: make(map[string]string),
	}

	// 解析layout XML
	layoutDoc := etree.NewDocument()
	if err := layoutDoc.ReadFromBytes(layoutContent); err != nil {
		return nil, fmt.Errorf("failed to parse layout file: %w", err)
	}
	layout.xml = layoutDoc

	// 获取布局名称
	if cSld := layoutDoc.FindElement("//p:cSld"); cSld != nil {
		if nameAttr := cSld.SelectAttr("name"); nameAttr != nil {
			layout.name = nameAttr.Value
		}
	}

	// 解析layout关系文件
	layoutRelsPath := filepath.Join("ppt/slideLayouts/_rels", filepath.Base(layoutPath)+".rels")
	// fmt.Println("layoutRelsPath1=====>", layoutRelsPath)
	// fmt.Println(p.files[layoutRelsPath])
	if relsContent, ok := p.files[layoutRelsPath]; ok {
		layout.relsPath = layoutRelsPath
		// fmt.Println("layoutRelsPath=====>", layoutRelsPath)
		relsDoc := etree.NewDocument()
		if err := relsDoc.ReadFromBytes(relsContent); err != nil {
			return nil, fmt.Errorf("failed to parse layout rels: %w", err)
		}

		for _, rel := range relsDoc.FindElements("//Relationship") {
			id := rel.SelectAttr("Id").Value
			target := rel.SelectAttr("Target").Value
			layout.rels[id] = target
		}
	}

	return layout, nil
}

// initSlides 初始化幻灯片
func (p *Presentation) initSlides(presDoc *etree.Document) error {
	p.slides = make([]*Slide, 0)

	// 查找所有sldId元素
	for _, slideEl := range presDoc.FindElements("//p:sldId") {
		rId := slideEl.SelectAttr("r:id").Value
		slidePath := "ppt/" + p.rels[rId]

		slideContent, ok := p.files[slidePath]
		if !ok {
			return fmt.Errorf("slide file not found: %s", slidePath)
		}

		slide := &Slide{
			path: slidePath,
			rels: make(map[string]*Relationship),
		}

		// 解析slide XML
		slideDoc := etree.NewDocument()
		if err := slideDoc.ReadFromBytes(slideContent); err != nil {
			return fmt.Errorf("failed to parse slide file: %w", err)
		}
		slide.xml = slideDoc

		// 解析slide关系文件
		slideRelsPath := filepath.Join("ppt/_rels", filepath.Base(slidePath)+".rels")
		if relsContent, ok := p.files[slideRelsPath]; ok {
			slide.relsPath = slideRelsPath
			relsDoc := etree.NewDocument()
			if err := relsDoc.ReadFromBytes(relsContent); err != nil {
				return fmt.Errorf("failed to parse slide rels: %w", err)
			}

			for _, rel := range relsDoc.FindElements("//Relationship") {
				id := rel.SelectAttr("Id").Value
				target := rel.SelectAttr("Target").Value
				rel := &Relationship{
					Id:         id,
					Target:     target,
					Type:       rel.SelectAttr("Type").Value,
					TargetMode: rel.SelectAttr("TargetMode").Value,
				}
				slide.rels[id] = rel
			}
		}

		p.slides = append(p.slides, slide)
	}

	return nil
}

// Save 保存PPTX文件
func (p *Presentation) Save(filename string) error {
	// 更新所有已修改的XML文件到files map中
	if err := p.updateFiles(); err != nil {
		return fmt.Errorf("failed to update files: %w", err)
	}

	buf := new(bytes.Buffer)
	writer := zip.NewWriter(buf)

	// 写入所有文件
	for name, content := range p.files {
		w, err := writer.Create(name)
		if err != nil {
			return fmt.Errorf("failed to create zip entry %s: %w", name, err)
		}

		_, err = w.Write(content)
		if err != nil {
			return fmt.Errorf("failed to write zip entry %s: %w", name, err)
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close zip writer: %w", err)
	}

	if err := os.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// updateFiles 更新所有已修改的XML文件到files map中
func (p *Presentation) updateFiles() error {
	// 更新slides
	for _, slide := range p.slides {
		if slide.xml != nil {
			data, err := slide.xml.WriteToBytes()
			if err != nil {
				return fmt.Errorf("failed to serialize slide XML: %w", err)
			}
			p.files[slide.path] = data
		}

		// 保存关系文件
		if len(slide.rels) > 0 {
			relsDoc := etree.NewDocument()
			relationships := relsDoc.CreateElement("Relationships")
			relationships.CreateAttr("xmlns", NsRelationships)

			for id, r := range slide.rels {
				rel := relationships.CreateElement("Relationship")
				rel.CreateAttr("Id", id)
				rel.CreateAttr("Type", r.Type)
				rel.CreateAttr("Target", r.Target)
				if r.TargetMode == "External" {
					rel.CreateAttr("TargetMode", r.TargetMode)
				}
			}

			data, err := relsDoc.WriteToBytes()
			if err != nil {
				return fmt.Errorf("failed to serialize slide relationships: %w", err)
			}
			p.files[slide.relsPath] = data
		}
	}

	// 更新layouts
	for _, master := range p.masters {
		for _, layout := range master.layouts {
			if layout.xml != nil {
				data, err := layout.xml.WriteToBytes()
				if err != nil {
					return fmt.Errorf("failed to serialize layout XML: %w", err)
				}
				p.files[layout.path] = data
			}
		}
	}

	// 更新masters
	for _, master := range p.masters {
		if master.xml != nil {
			data, err := master.xml.WriteToBytes()
			if err != nil {
				return fmt.Errorf("failed to serialize master XML: %w", err)
			}
			p.files[master.path] = data
		}
	}

	return nil
}

// Close 关闭PPTX文件
func (p *Presentation) Close() error {
	if p.zipReader != nil {
		return p.zipReader.Close()
	}
	return nil
}

// GetMasterByName 通过名称获取母版
func (p *Presentation) GetMasterByName(name string) *Master {
	for _, master := range p.masters {
		if master.name == name {
			return master
		}
	}
	return nil
}

// GetLayoutByName 通过名称获取布局
func (p *Presentation) GetLayoutByName(name string) *Layout {
	// for _, master := range p.masters {
	// 	for idx, layout := range master.layouts {
	// 		fmt.Println(idx, layout.name)
	// 	}
	// }
	for _, master := range p.masters {
		for _, layout := range master.layouts {
			if layout.name == name {
				return layout
			}
		}
	}
	return nil
}

// AddSlide 添加新的幻灯片
func (p *Presentation) AddSlide(layoutName string) (*Slide, error) {
	// 查找布局
	layout := p.GetLayoutByName(layoutName)
	if layout == nil {
		return nil, fmt.Errorf("layout not found: %s", layoutName)
	}

	// 创建新的slide XML，从layout复制
	slideDoc := etree.NewDocument()
	// 添加完整的XML声明
	slideDoc.CreateProcInst("xml", `version="1.0" encoding="UTF-8" standalone="yes"`)

	if layout.xml != nil && layout.xml.Root() != nil {
		root := layout.xml.Root().Copy()
		root.Tag = "sld"
		slideDoc.SetRoot(root)
	} else {
		return nil, fmt.Errorf("invalid layout XML")
	}

	// 设置slide路径
	slideIndex := len(p.slides) + 1
	slidePath := fmt.Sprintf("ppt/slides/slide%d.xml", slideIndex)
	slideRelsPath := fmt.Sprintf("ppt/slides/_rels/slide%d.xml.rels", slideIndex)

	// 首先复制布局的关系
	layoutRels := make(map[string]*Relationship)
	// fmt.Println("layoutRels=====>", layout.relsPath)
	if layoutRelsData, exists := p.files[layout.relsPath]; exists {
		// 解析布局关系文件
		layoutRelsDoc := etree.NewDocument()
		if err := layoutRelsDoc.ReadFromBytes(layoutRelsData); err != nil {
			return nil, fmt.Errorf("failed to parse layout relationships: %w", err)
		}

		// 复制所有关系
		for _, rel := range layoutRelsDoc.FindElements("//Relationship") {
			layoutRels[rel.SelectAttrValue("Id", "")] = &Relationship{
				Id:         rel.SelectAttrValue("Id", ""),
				Type:       rel.SelectAttrValue("Type", ""),
				Target:     rel.SelectAttrValue("Target", ""),
				TargetMode: rel.SelectAttrValue("TargetMode", ""),
			}
		}
	}

	// 创建新的slide对象，使用复制的关系
	slide := &Slide{
		xml:      slideDoc,
		path:     slidePath,
		relsPath: slideRelsPath,
		rels:     layoutRels, // 使用复制的布局关系
		layout:   layout,
		master:   p.findMasterForLayout(layout),
		pres:     p,
	}

	// 生成新的rId
	newRid := fmt.Sprintf("rId%d", len(slide.rels)+1)

	// 添加对layout的基础引用关系
	layoutRel := &Relationship{
		Id:     newRid,
		Type:   "http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout",
		Target: "../slideLayouts/" + filepath.Base(layout.path),
	}
	slide.rels[newRid] = layoutRel

	// 更新presentation.xml中的幻灯片列表
	if err := p.updatePresentationSlideList(slide); err != nil {
		return nil, fmt.Errorf("failed to update presentation slide list: %w", err)
	}

	// 序列化并保存slide XML
	slideData, err := slideDoc.WriteToBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize slide XML: %w", err)
	}
	p.files[slidePath] = slideData

	// 创建slide关系文件
	relsDoc := etree.NewDocument()
	relationships := relsDoc.CreateElement("Relationships")
	relationships.CreateAttr("xmlns", NsRelationships)

	// 添加所有关系（包括从布局复制的和layout引用）
	for _, rel := range slide.rels {
		relationship := relationships.CreateElement("Relationship")
		relationship.CreateAttr("Id", rel.Id)
		relationship.CreateAttr("Type", rel.Type)
		relationship.CreateAttr("Target", rel.Target)
		if rel.TargetMode != "" {
			relationship.CreateAttr("TargetMode", rel.TargetMode)
		}
	}

	relsData, err := relsDoc.WriteToBytes()
	if err != nil {
		return nil, fmt.Errorf("failed to create slide relationships: %w", err)
	}
	p.files[slideRelsPath] = relsData

	// 添加到幻灯片集合
	p.slides = append(p.slides, slide)

	return slide, nil
}

// DeleteSlide 删除指定索引的幻灯片
func (p *Presentation) DeleteSlide(index int) error {
	if index < 0 || index >= len(p.slides) {
		return fmt.Errorf("invalid slide index: %d", index)
	}

	// 获取要删除的幻灯片
	slide := p.slides[index]

	// 从presentation.xml中删除引用
	if err := p.removeSlideReference(slide); err != nil {
		return fmt.Errorf("failed to remove slide reference: %w", err)
	}

	// 删除幻灯片文件和关系文件
	delete(p.files, slide.path)
	delete(p.files, slide.relsPath)

	// 从幻灯片集合中删除
	p.slides = append(p.slides[:index], p.slides[index+1:]...)

	return nil
}

// removeSlideReference 从presentation.xml中删除幻灯片引用
func (p *Presentation) removeSlideReference(slide *Slide) error {
	presContent, ok := p.files["ppt/presentation.xml"]
	if !ok {
		return fmt.Errorf("presentation.xml not found")
	}

	presDoc := etree.NewDocument()
	if err := presDoc.ReadFromBytes(presContent); err != nil {
		return fmt.Errorf("failed to parse presentation.xml: %w", err)
	}

	// 查找并删除sldId元素
	sldIdLst := presDoc.FindElement("//p:sldIdLst")
	if sldIdLst != nil {
		for _, sldId := range sldIdLst.SelectElements("p:sldId") {
			rId := sldId.SelectAttrValue("r:id", "")
			if target, ok := p.rels[rId]; ok && target == strings.TrimPrefix(slide.path, "ppt/") {
				sldIdLst.RemoveChild(sldId)
				break
			}
		}
	}

	// 更新presentation.xml
	data, err := presDoc.WriteToBytes()
	if err != nil {
		return fmt.Errorf("failed to serialize presentation.xml: %w", err)
	}
	p.files["ppt/presentation.xml"] = data

	return nil
}

func (p *Presentation) GetSlides() []*Slide {
	return p.slides
}

// GetSlide 获取指定索引的幻灯片
func (p *Presentation) GetSlide(index int) (*Slide, error) {
	if index < 0 || index >= len(p.slides) {
		return nil, fmt.Errorf("invalid slide index: %d", index)
	}
	return p.slides[index], nil
}

// findMasterForLayout 查找布局对应的母版
func (p *Presentation) findMasterForLayout(layout *Layout) *Master {
	for _, master := range p.masters {
		for _, l := range master.layouts {
			if l == layout {
				return master
			}
		}
	}
	return nil
}

// updatePresentationRels 更新presentation.xml.rels
func (p *Presentation) updatePresentationRels(slide *Slide, rId string) error {
	relsPath := "ppt/_rels/presentation.xml.rels"
	relsContent, ok := p.files[relsPath]
	if !ok {
		return fmt.Errorf("presentation.xml.rels not found")
	}

	relsDoc := etree.NewDocument()
	if err := relsDoc.ReadFromBytes(relsContent); err != nil {
		return fmt.Errorf("failed to parse presentation.xml.rels: %w", err)
	}

	// 添加新的关系
	relationships := relsDoc.FindElement("//Relationships")
	if relationships == nil {
		return fmt.Errorf("relationships element not found")
	}

	rel := relationships.CreateElement("Relationship")
	rel.CreateAttr("Id", rId)
	rel.CreateAttr("Type", "http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide")
	rel.CreateAttr("Target", strings.TrimPrefix(slide.path, "ppt/"))

	// 更新关系文件
	data, err := relsDoc.WriteToBytes()
	if err != nil {
		return fmt.Errorf("failed to serialize presentation.xml.rels: %w", err)
	}
	p.files[relsPath] = data

	return nil
}

// updatePresentationSlideList 更新presentation.xml中的幻灯片列表
func (p *Presentation) updatePresentationSlideList(slide *Slide) error {
	// 获取presentation.xml
	presContent, ok := p.files["ppt/presentation.xml"]
	if !ok {
		return fmt.Errorf("presentation.xml not found")
	}

	presDoc := etree.NewDocument()
	if err := presDoc.ReadFromBytes(presContent); err != nil {
		return fmt.Errorf("failed to parse presentation.xml: %w", err)
	}

	// 查找或创建sldIdLst元素
	sldIdLst := presDoc.FindElement("//p:sldIdLst")
	if sldIdLst == nil {
		presentation := presDoc.FindElement("//p:presentation")
		if presentation == nil {
			return fmt.Errorf("presentation element not found")
		}
		sldIdLst = presentation.CreateElement("p:sldIdLst")
	}

	// 创建新的slide ID
	maxId := 255 // 起始ID
	for _, sldId := range sldIdLst.SelectElements("p:sldId") {
		if id := sldId.SelectAttrValue("id", ""); id != "" {
			if idNum := atoi(id); idNum > maxId {
				maxId = idNum
			}
		}
	}
	newId := maxId + 1
	// 从presentation.xml.rels获取最大的rId
	relsPath := "ppt/_rels/presentation.xml.rels"
	relsContent, ok := p.files[relsPath]
	if !ok {
		return fmt.Errorf("presentation.xml.rels not found")
	}

	relsDoc := etree.NewDocument()
	if err := relsDoc.ReadFromBytes(relsContent); err != nil {
		return fmt.Errorf("failed to parse presentation.xml.rels: %w", err)
	}

	// 找到最大的rId
	maxRid := 0
	for _, rel := range relsDoc.FindElements("//Relationship") {
		if rid := rel.SelectAttr("Id"); rid != nil {
			if num := getRidNumber(rid.Value); num > maxRid {
				maxRid = num
			}
		}
	}
	newRid := fmt.Sprintf("rId%d", maxRid+1)
	// 创建新的sldId元素
	sldId := sldIdLst.CreateElement("p:sldId")
	sldId.CreateAttr("id", fmt.Sprintf("%d", newId))
	sldId.CreateAttr("r:id", newRid)

	// 更新关系文件
	if err := p.updatePresentationRels(slide, newRid); err != nil {
		return fmt.Errorf("failed to update presentation rels: %w", err)
	}

	// 更新presentation.xml
	data, err := presDoc.WriteToBytes()
	if err != nil {
		return fmt.Errorf("failed to serialize presentation.xml: %w", err)
	}
	p.files["ppt/presentation.xml"] = data

	return nil
}

// getRidNumber 从rId字符串中提取数字
func getRidNumber(rid string) int {
	var num int
	fmt.Sscanf(rid, "rId%d", &num)
	return num
}
