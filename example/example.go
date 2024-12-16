package main

import (
	"fmt"
	"log"

	"github.com/wanglihui/pptx-go/pptx"
)

func main() {
	// 示例1：打开模板并添加新幻灯片
	if err := example1(); err != nil {
		log.Fatal("Example 1 failed:", err)
	}

	// 示例2：修改占位符内容
	if err := example2(); err != nil {
		log.Fatal("Example 2 failed:", err)
	}

	// 示例3：创建完整的演示文稿
	if err := example3(); err != nil {
		log.Fatal("Example 3 failed:", err)
	}
}

// example1 演示基本的打开和保存操作
func example1() error {
	// 打开模板文件
	pres, err := pptx.Open("templates/template4.pptx")
	if err != nil {
		return fmt.Errorf("failed to open template: %w", err)
	}
	defer pres.Close()
	for {
		if len(pres.GetSlides()) == 0 {
			break
		}
		if err := pres.DeleteSlide(0); err != nil {
			fmt.Printf("%v", err)
		}
	}

	// 添加一个新的幻灯片（使用"Title Slide"布局）
	slide, err := pres.AddSlide("仅标题")
	if err != nil {
		return fmt.Errorf("failed to add slide: %w", err)
	}

	// 获取标题占位符并设置文本
	titlePlaceholder, err := slide.GetPlaceholder(pptx.PlaceholderTitle)
	if err != nil {
		return fmt.Errorf("failed to get title placeholder: %w", err)
	}
	titlePlaceholder.Debug()
	err = titlePlaceholder.SetText("Hello, PPTX!")
	if err != nil {
		return fmt.Errorf("failed to set title placeholder: %w", err)
	}

	// 保存修改后的文件
	if err := pres.Save("output/output1.pptx"); err != nil {
		return fmt.Errorf("failed to save presentation: %w", err)
	}

	return nil
}

// example2 演示占位符操作
func example2() error {
	// 打开模板文件
	pres, err := pptx.Open("templates/template4.pptx")
	if err != nil {
		return fmt.Errorf("failed to open template: %w", err)
	}
	defer pres.Close()

	// 添加一个内容幻灯片
	slide, err := pres.AddSlide("内容与标题")
	if err != nil {
		return fmt.Errorf("failed to add slide: %w", err)
	}

	// 设置标题
	if title, err := slide.GetPlaceholder(pptx.PlaceholderTitle); err == nil {
		title.SetText("Placeholder Demo")
	}

	// 设置正文内容
	if body, err := slide.GetPlaceholder(pptx.PlaceholderBody); err == nil {
		body.SetText("This is the body text\n• Point 1\n• Point 2\n• Point 3")
	}

	// 添加图片
	if pic, err := slide.GetPlaceholder(pptx.PlaceholderImage); err == nil {
		pic.SetImage("example.jpg")
	}

	// 添加表格
	if tbl, err := slide.GetPlaceholder(pptx.PlaceholderTable); err == nil {
		data := [][]string{
			{"Header 1", "Header 2", "Header 3"},
			{"Row 1, Col 1", "Row 1, Col 2", "Row 1, Col 3"},
			{"Row 2, Col 1", "Row 2, Col 2", "Row 2, Col 3"},
		}
		tbl.SetTable(data)
	}

	// 保存修改后的文件
	if err := pres.Save("output/output2.pptx"); err != nil {
		return fmt.Errorf("failed to save presentation: %w", err)
	}

	return nil
}

// example3 演示创建完整演示文稿
func example3() error {
	// 打开模板文件
	pres, err := pptx.Open("templates/template4.pptx")
	if err != nil {
		return fmt.Errorf("failed to open template: %w", err)
	}
	defer pres.Close()

	// 1. 创建标题幻灯片
	titleSlide, err := pres.AddSlide("标题幻灯片")
	if err != nil {
		return fmt.Errorf("failed to add title slide: %w", err)
	}

	if title, err := titleSlide.GetPlaceholder(pptx.PlaceholderTitle); err == nil {
		title.SetText("Annual Report 2023")
	}
	if subtitle, err := titleSlide.GetPlaceholder("Subtitle"); err == nil {
		subtitle.SetText("Company Performance Overview")
	}

	// 2. 创建目录幻灯片
	tocSlide, err := pres.AddSlide("节标题")
	if err != nil {
		return fmt.Errorf("failed to add TOC slide: %w", err)
	}

	if title, err := tocSlide.GetPlaceholder(pptx.PlaceholderTitle); err == nil {
		title.SetText("Agenda")
	}
	if body, err := tocSlide.GetPlaceholder(pptx.PlaceholderBody); err == nil {
		body.SetText("1. Financial Highlights\n2. Market Analysis\n3. Future Strategy\n4. Q&A")
	}

	// 3. 创建图表幻灯片
	chartSlide, err := pres.AddSlide("标题和内容")
	if err != nil {
		return fmt.Errorf("failed to add chart slide: %w", err)
	}

	if title, err := chartSlide.GetPlaceholder(pptx.PlaceholderTitle); err == nil {
		title.SetText("Financial Highlights")
	}

	// 4. 创建图片幻灯片
	imageSlide, err := pres.AddSlide("图片与标题")
	if err != nil {
		return fmt.Errorf("failed to add image slide: %w", err)
	}

	if title, err := imageSlide.GetPlaceholder(pptx.PlaceholderTitle); err == nil {
		title.SetText("Market Analysis")
	}
	if pic, err := imageSlide.GetPlaceholder(pptx.PlaceholderImage); err == nil {
		pic.SetImage("market_analysis.jpg")
	}

	// 5. 创建总结幻灯片
	summarySlide, err := pres.AddSlide("内容与标题")
	if err != nil {
		return fmt.Errorf("failed to add summary slide: %w", err)
	}

	if title, err := summarySlide.GetPlaceholder(pptx.PlaceholderTitle); err == nil {
		title.SetText("Thank You!")
	}
	if body, err := summarySlide.GetPlaceholder(pptx.PlaceholderBody); err == nil {
		body.SetText("Questions?\nContact: example@company.com")
	}

	// 保存演示文稿
	if err := pres.Save("output/annual_report_2023.pptx"); err != nil {
		return fmt.Errorf("failed to save presentation: %w", err)
	}

	return nil
}
