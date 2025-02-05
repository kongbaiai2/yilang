# !/usr/bin/python
# -*- coding: utf-8 -*-

# import jieba #分词


# 词云,数据可视化pyecharts


from pyecharts import options as opts
from pyecharts.charts import Pie
from pyecharts.faker import Faker
from pyecharts.charts import Bar
from pyecharts.charts import WordCloud

# # 创建虚拟环境
# python -m venv /path/to/new/virtual/environment
 
# # 激活虚拟环境
# # 在Windows上
# /path/to/new/virtual/environment/Scripts/activate
# # 在Unix或MacOS上
# source /path/to/new/virtual/environment/bin/activate
# pip install virtualenv
# # 创建虚拟环境
# virtualenv /path/to/new/virtual/environment
 
# # 激活虚拟环境
# # 在Windows上
# /path/to/new/virtual/environment/Scripts/activate
# # 在Unix或MacOS上
# source /path/to/new/virtual/environment/bin/activate
# CSV文件及相关代码：
# 			链接：https://pan.baidu.com/s/1ugxtHNF7wpvZB2meKXhQpQ  
#                                        密码：r8qx

# 可视化图库「Pandas_Alive」，不仅包含动态条形图，还可以绘制动态曲线图、气泡图、饼状图、地图等。   
#       GitHub地址：https://github.com/JackMcKew/pandas_alive
#       使用文档：https://jackmckew.github.io/pandas_alive/
#       安装版本建议是0.2.3，matplotlib版本是3.2.1     
#       同时需自行安装tqdm(显示进度条)和descartes(绘制地图相关库)

# 可视化图库「Pandas_Alive」，不仅包含动态条形图，还可以绘制动态曲线图、气泡图、饼状图、地图等。   
#       GitHub地址：https://github.com/JackMcKew/pandas_alive
#       使用文档：https://jackmckew.github.io/pandas_alive/
#       安装版本建议是0.2.3，matplotlib版本是3.2.1     
#       同时需自行安装tqdm(显示进度条)和descartes(绘制地图相关库)


# Echarts 是一个由百度开源的数据可视化，凭借着良好的交互性，精巧的图表设计，得到了众多开发者的认可。而 Python 是一门富有表达力的语言，很适合用于数据处理。当数据分析遇上数据可视化时，pyecharts 诞生了。
#         官网：https://pyecharts.org/#/zh-cn/intro


# 使用词云进行文本分析的第三方库是wordcloud、matplotlib和scipy，其中wordcloud需要手动下载，其网址为http://www.lfd.uci.edu/~gohlke/Pythonlibs/#wordcloud
def Pie_graph():
    a = [("教育", 6800), ("吃穿", 8010), ("住行", 8040), ("医疗", 8100), ("旅行", 8020), ("培训", 8030), ("戒尺", 2800), ("其它", 1800), ]
    pie=Pie()
    # pie.add('',[list(z) for z in zip(Faker.choose(),Faker.values())])
    pie.add('消费',a)
    # .set_global_opts(title_opts=opts.TitleOpts(formater="{b}:{c}"))
    pie.set_global_opts(title_opts=opts.TitleOpts(title="饼图测试"))
    pie.page_title="sdflkj"
    pie.render("pie_base.html")
  
def Bar_graph():
    a = [("教育", 6800), ("吃穿", 8010), ("住行", 8040), ("医疗", 8100), ("旅行", 8020), ("培训", 8030), ("戒尺", 2800), ("其它", 1800), ]

    bar = Bar()
    bar.add_xaxis(["分类1", "分类2", "分类3"])
    bar.add_yaxis("系列名称", [10, 20, 30])
    # bar.add_dataset(a)
    bar.page_title="sdflkj"
    bar.render("bar_base.html")
def Word_cloud():
    data = [("住行", 8040), ("医疗", 8100), ("旅行", 6020), ("培训", 8030), ("戒尺", 5800), ("其它", 4800),]
    wrd = WordCloud()
    wrd.add(series_name="test",data_pair=data,word_size_range=[66,66])
    wrd.set_global_opts(title_opts=opts.TitleOpts(title="testest",title_textstyle_opts="font_style=23"),
                        tooltip_opts=opts.TooltipOpts(is_show=True)),
    wrd.page_title="sdflkj"
    wrd.render("wrd_base.html")
if __name__ == '__main__':
    Pie_graph()
    Bar_graph()
    Word_cloud()


