# ngapost2md ver.[NEO_1.3.0]

ngapost2md 是一个将 NGA 论坛帖子转换为 Markdown 格式的工具。它支持快速爬楼并存储回复人、时间和内容，同时支持保存正文图片。

**此为 2023 年由 Go 语言重写的版本。倘若需要旧版 Python 版代码（不再维护），请切换分支至 LEGACY**

<img src="README.assets/gen_md_demo.png" width="900px" alt="gen_md_demo">

## 使用说明
1. 下载并解压发布版本的压缩包。
2. 修改 config.ini 文件中的配置项，根据需要进行相应的修改，确保 `config.ini`  文件存在且与可执行文件在同一目录下（平级关系）。
3. 打开终端或命令提示符。
4. 运行以下命令，并执行程序：

linux
```
./ngapost2md 5935947
```
windows
```
.\ngapost2md.exe 5935947
```
参数为帖子的 tid 。

5. 程序会开始爬取帖子内容并将其转换为 Markdown 格式，转换后的文件将保存在当前目录。
## 配置说明

请在 `config.ini` 文件中修改以下配置项：

```ini
[config]
version="1.2.0" ;请不要修改此处

[network]
base_url="https://bbs.nga.cn" ;软件访问的 NGA 域名。默认值 "https://bbs.nga.cn"
ua="""MODIFY_ME""" ;浏览器 User-Agent，请修改。通常来说填写你常用浏览器的 UA 即可。
ngaPassportUid="MODIFY_ME" ;NGA 网站个人 cookie，请修改。
ngaPassportCid="MODIFY_ME" ;NGA 网站个人 cookie，请修改。
thread=2 ;线程数，提高理论上可以增加下载速度。仅支持 1、2、3。若开启 enhance_ori_reply，请将此值设定为 1。默认值 2
page_download_limit=100; 每次下载限制新下载的大约页数。到上限后需要重新运行程序再追加下载，如此直至全部下载成功。允许范围 -1（含）至 100（含）。值为 0 或 -1 时则不限制。默认值 100（约 100 页）。

[post]
enable_post_title=False ;是否将 .md 文件以标题命名。默认值 False（不启用）
get_ip_location=False ;是否查询用户基于 IP 的地理位置？若启用则会导致至高 20 倍于未启用的网络请求。默认值 False（不启用）
enhance_ori_reply=False ;将被回复的楼层内容补充完整。见 issue#35 。开启此功能要求同步将 thread 线程数设置为 1，否则可能会补充到未 format 的文本。默认值 False（不启用）
```


## 注意事项


- 请确保您的网络连接正常，并且能够访问 NGA 论坛。
- 请遵守 NGA 论坛的相关规定和版权要求。
- 请使用合法、合规的方式进行爬取，遵守网站的爬虫规范和使用协议。
- 请尊重网站的服务器负载和带宽限制，避免对其造成过大的压力。
- 请避免频繁的请求和大量的并发连接，以免对网站的正常运行造成干扰。
- 转换过程可能需要一些时间，具体时间取决于帖子的页数和内容数量。

## 资瓷与不资瓷格式说明

资瓷的有：

- newline 换行
- pic 图片（会下载下来）
- smile 表情（只是引用在线资源）
- quote 回复与引用（阔以 jump 和 append 在最后 [#12](https://github.com/ludoux/ngapost2md/issues/12)）（多个 quote [#33](https://github.com/ludoux/ngapost2md/issues/33)）
- strikeout 删除线
- url 超链接
- anony 匿名 （[#11](https://github.com/ludoux/ngapost2md/issues/11)）
- 用户基于 IP 的位置 （[#45](https://github.com/ludoux/ngapost2md/pull/45)）

不资瓷并且常出现的有：
- ~~align 对齐~~ 目前 Go 版本不支持
- ~~collapse 折叠 （[#10](https://github.com/ludoux/ngapost2md/issues/10)）~~ 目前 Go 版本不支持
- ~~audio 音频 （[#15](https://github.com/ludoux/ngapost2md/issues/15)）~~ 目前 Go 版本不支持
- ~~video 音频~~ 目前 Go 版本不支持
- 字体颜色啊大小之类的格式
- 表格之类的复杂排版

## Special Thanks

- 特别感谢 [zsq @oarinv](https://github.com/oarinv) 的协助！
- 特别感谢 [crella6](https://github.com/crella6) 的捉虫以及意见！
- 感谢 [proItheus](https://github.com/proItheus) 对此项目的帮助！
