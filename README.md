# ngapost2md ver.[3]

快速爬楼存回复人+时间+内容，支持保存正文图片、音视频等，支持常见 bbcode 格式（具体见 README 后半部分）。

**2022 年新版在每次运行前，会访问 gitee 以检查版本信息，若有新版会提示并自行退出，请下载最新版本代码（`nga.py` 与 `nga_format.py`）运行即可。若出现问题请自行移除相关检查更新代码。**

## 使用指引

1. 把nga.py **和 nga_format.py** 下下来，修改前者代码文件内的 cookies（cookies是自己账号登录后的具体内容）

2. ~~将[smile.zip](https://github.com/ludoux/ngapost2md/releases/tag/alpha)解压，确保smile文件夹（里面就是各种ac娘表情包）和nga.py在同一个目录下~~

3. 双击启动输入 tid 即可，之后会反显爬楼爬页的情况和下图片的情况

4. 最后会在 nga.py 所在的目录下出一个新的以 tid 命名的文件夹，里面有 post.md 直接查看就行。info.txt 可以看标题和每次增量的具体信息（以及错误信息）。

### 图片快速指引

![image-20200414232616854](README.assets/image-20200414232616854.png)

![image-20200414232733377](README.assets/image-20200414232733377.png)

![image-20200414232929882](README.assets/image-20200414232929882.png)

![postmd2](README.assets/postmd2.png)

![image-20200414233052905](README.assets/image-20200414233052905.png)

## 资瓷与不资瓷格式说明

资瓷的有：

- newline 换行
- pic 图片（会下载下来）
- smile 表情（只是引用在线资源）
- quote 回复与引用（阔以 jump 和 append 在最后 [#12](https://github.com/ludoux/ngapost2md/issues/12)）（多个 quote [#33](https://github.com/ludoux/ngapost2md/issues/33)）
- strikeout 删除线
- url 超链接
- align 对齐
- collapse 折叠 （[#10](https://github.com/ludoux/ngapost2md/issues/10)）
- anony 匿名 （[#11](https://github.com/ludoux/ngapost2md/issues/11)）
- audio 音频 （[#15](https://github.com/ludoux/ngapost2md/issues/15)）
- video 音频

不资瓷并且常出现的有：

- 字体颜色啊大小之类的格式
- 表格之类的复杂排版

## Special Thanks

特别感谢 [crella6](https://github.com/crella6) 的捉虫以及意见！
