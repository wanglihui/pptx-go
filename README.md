# pptx-go 

参考python-pptx实现go版本的pptx操作库；
pptx-go 直接操作 open office xml 文件，不依赖unioffice库，避免版权问题;

## pptx-go 主要的功能为:
1. 读取pptx模版
2. 使用pptx母版中的layout添加 slide
3. 读取slide中的placeholder，并替换为对应的值;
    - 支持的placeholder类型有:
        - 文本框
        - 图片
        - 表格
        - 形状
        - 图表
        - 幻灯片
        - 页眉页脚
        - 页码
        - 页边距
    - 替换placeholder的时候，支持基于type或者name或者idx进行替换
4. 保存pptx文件

## 文件目录
- pptx/pptx.go  pptx 文件的读取和保存
    - 此文件主要封装操作pptx的函数，包括读取pptx，保存pptx等;
- pptx/slide.go  slide 的读取和保存
    - 此文件主要封装操作slide的函数，包括添加slide，删除slide，获取slide等;
    - 添加slide的时候，支持基于layout进行添加，支持基于master进行添加;
- pptx/placeholder.go placeholder 的读取和保存
    此文件主要封装操作placeholder的函数，包括替换placeholder的值，获取placeholder的值等;
    替换的方式有根据type, name, idx进行替换;
