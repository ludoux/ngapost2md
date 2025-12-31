package nga

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/buger/jsonparser"
	"github.com/imroc/req/v3"
	"github.com/panjf2000/ants/v2"
	"github.com/spf13/cast"
	"gopkg.in/ini.v1"
)

// 这里是配置文件可以改的
var (
	CFGFILE_THREAD_COUNT              = 2
	CFGFILE_GET_IP_LOCATION           = false //	获取ip地址
	CFGFILE_ENHANCE_ORI_REPLY         = false //	功能见 #35
	CFGFILE_ENHANCE_ORI_REPLY_ONLINE  = false //	功能见 #122
	CFGFILE_PAGE_DOWNLOAD_LIMIT       = 100   //	限制单次下载的页数 #56
	CFGFILE_USE_TITLE_AS_FOLDER_NAME  = false
	CFGFILE_USE_TITLE_AS_MD_FILE_NAME = false
	CFGFILE_USE_LOCAL_SMILE_PIC       = false       // 使用本地表情 #58
	CFGFILE_LOCAL_SMILE_PIC_PATH      = "../smile/" // 本地表情路径 #58
	CFGFILE_USE_NETWORK_MEDIA_URL     = false       // 媒体文件只引用在线链接 #109
	CFGFILE_ASSETS_PATH               = "./assets/" // 帖子资源路径 #124
	CFGFILE_SPLIT_MD_FILE             = -1          // 是否切分生成的md文件，以及单文件的页数 #105
)

// 这里传参可以改
// var ()

// 这里配置文件和传参都没法改
var (
	VERSION  = "1.10.0"     //需要手动改
	BUILD_TS = "1691664141" //无需，GitHub actions会自动填写
	GIT_REF  = ""           //无需，GitHub actions会自动填写
	GIT_HASH = ""           //无需，GitHub actions会自动填写
	DELAY_MS = 330
	mutex    sync.Mutex
)

// Flag 相关
var (
	page_download_limit_triggered = false
)

// ldflags 区域。GitHub Actions 编译时会使用 ldflags 来修改如下值：
var (
	DEBUG_MODE = "1" //GitHub Actions 打包的时候会修改为"0"。本地打包可以 go build -ldflags "-X 'github.com/ludoux/ngapost2md/nga.DEBUG_MODE=0'" main.go
	/**
	 * DEBUG_MODE 为true时会:
	 * 启动时禁用自动版本检查
	 */
)

// 通用媒体文件处理函数
func processMedia(content string, pattern string, urlPattern string, mediaType string, assets *map[string]string, floor *Floor, tiezi *Tiezi, isImage bool) string {
	re := regexp.MustCompile(pattern)
	for _, match := range re.FindAllStringSubmatch(content, -1) {
		var url string
		var fullMatch string

		if urlPattern != "" {
			// 如果需要从标签中提取URL
			srcRe := regexp.MustCompile(urlPattern)
			srcMatches := srcRe.FindStringSubmatch(match[0])
			if len(srcMatches) < 2 {
				continue
			}
			url = srcMatches[1]
			fullMatch = match[0]
		} else {
			// 如果match[1]就是URL。目前仅是img模式
			url = match[1]
			fullMatch = `[img]` + match[1] + `[/img]`
		}

		// 如果是图片，需要特殊处理URL格式
		if isImage {
			if url[0:2] == "./" {
				url = "https://img.nga.178.com/attachments/" + url[2:]
			}
			url = strings.ReplaceAll(url, ".medium.jpg", "")
		}

		var replacement string
		if CFGFILE_USE_NETWORK_MEDIA_URL {
			if isImage {
				replacement = fmt.Sprintf(`![img](%s)`, url)
			} else {
				replacement = fmt.Sprintf("【%s：%s】", mediaType, url)
			}
		} else {
			sha := sha256.Sum256([]byte(url))
			shaStr := hex.EncodeToString(sha[:])
			shorted := shaStr[2:8] + url[len(url)-6:]
			var fileName string

			mutex.Lock()
			var ok bool
			v, ok := (*assets)[shorted]
			if ok {
				//存在，直接复用
				fileName = v
			} else {
				fileName = cast.ToString(floor.Lou) + "_" + shorted
				(*assets)[shorted] = fileName
			}

			if !ok {
				mutex.Unlock()
				time.Sleep(time.Millisecond * time.Duration(DELAY_MS))
				log.Println("下载", mediaType, ":", fileName)
				// 确保目录存在
				assetDir := filepath.Join(".", tiezi.GetNeededFolderName(), CFGFILE_ASSETS_PATH)
				os.MkdirAll(assetDir, os.ModePerm)
				downloadAssets(url, filepath.Join(assetDir, fileName))
				//log.Println("下载",mediaType,"成功:", fileName)
			} else {
				mutex.Unlock()
			}
			// 更新引用路径
			relativePath := filepath.Join(".", CFGFILE_ASSETS_PATH, fileName)
			if isImage {
				replacement = fmt.Sprintf(`![img](%s)`, relativePath)
			} else {
				replacement = fmt.Sprintf("【%s：%s】", mediaType, relativePath)
			}
		}
		content = strings.ReplaceAll(content, fullMatch, replacement)
	}
	return content
}

type Floor struct {
	Lou        int
	Pid        int
	Timestamp  int64
	Username   string
	IpLocation string
	UserId     int
	Content    string
	LikeNum    int
	AppendPid  []int
	Comments   Floors
}
type Floors []Floor
type Tiezi struct {
	Tid             int
	AuthorId        int // 这个是用户传入的希望仅下载某用户id的发言贴参数
	Title           string
	TitleFolderSafe string
	Catelogy        string
	Username        string
	UserId          int
	WebMaxPage      int
	LocalMaxPage    int
	LocalMaxFloor   int
	FloorCount      int    // 包含主楼
	Floors          Floors // 主楼为[0]
	HotPosts        Floors
	Timestamp       int64  // page() fixFloorContent()  中会更新
	Version         string // 这个是软件的version
	Assets          map[string]string
}

var responseChannel = make(chan string, 15)

/**
 * @description: 分析floors原始数据并填充进floors里
 * @param {[]byte} resp 接口下来的原始数据
 * @param {bool} isComments 是否是下挂评论。假如是的话则自增楼层
 * @return {*}
 */
func (it *Floors) analyze(resp []byte, isComments bool) {
	lou_comment := 1
	jsonparser.ArrayEach(resp, func(value []byte, _ jsonparser.ValueType, _ int, _ error) {
		value_int, _ := jsonparser.GetInt(value, "lou")
		var lou int
		if !isComments {
			lou = cast.ToInt(value_int)
		} else {
			lou = lou_comment
		}
		// 根据楼数补充Floors
		for len(*it) < lou+1 {
			(*it) = append((*it), Floor{Lou: -1})
		}

		curFloor := &(*it)[lou]

		// 楼层
		curFloor.Lou = lou

		// PID
		value_int, _ = jsonparser.GetInt(value, "pid")
		curFloor.Pid = cast.ToInt(value_int)

		// 时间戳
		value_int, _ = jsonparser.GetInt(value, "postdatetimestamp")
		curFloor.Timestamp = value_int

		// 用户名
		value_str, _ := jsonparser.GetString(value, "author", "username")
		curFloor.Username = value_str

		// 用户id
		value_int, _ = jsonparser.GetInt(value, "author", "uid")
		curFloor.UserId = cast.ToInt(value_int)

		// 内容
		value_str, _ = jsonparser.GetString(value, "content")
		curFloor.Content = value_str

		// 点赞数
		value_int, _ = jsonparser.GetInt(value, "vote_good")
		curFloor.LikeNum = cast.ToInt(value_int)

		// 下挂comments
		value_byte, dataType, _, _ := jsonparser.Get(value, "comments")
		if dataType != jsonparser.NotExist {
			curFloor.Comments.analyze(value_byte, true)
		}
		lou_comment = lou_comment + 1
	})
}

/**
 * @description: 针对 tiezi 对象获取指定页的信息
 * @param {int} page 指定的页数
 * @return {*}
 */
func (tiezi *Tiezi) page(page int) {
	var resp *req.Response
	var err error
	if tiezi.AuthorId > 0 {
		resp, err = Client.R().SetFormData(map[string]string{
			"page":     cast.ToString(page),
			"tid":      cast.ToString(tiezi.Tid),
			"authorid": cast.ToString(tiezi.AuthorId),
		}).Post("app_api.php?__lib=post&__act=list")
	} else {
		resp, err = Client.R().SetFormData(map[string]string{
			"page": cast.ToString(page),
			"tid":  cast.ToString(tiezi.Tid),
		}).Post("app_api.php?__lib=post&__act=list")
	}
	if err != nil {
		log.Println(err.Error())
	}
	code, _ := jsonparser.GetInt(resp.Bytes(), "code")
	if code != 0 {
		msg, _ := jsonparser.GetString(resp.Bytes(), "msg")
		log.Fatalln("nga 返回代码不为0:", code, msg)
	} else {
		tiezi.Timestamp = ts()

		// 标题
		value_str, _ := jsonparser.GetString(resp.Bytes(), "tsubject")
		tiezi.Title = value_str
		tiezi.TitleFolderSafe = ToSaveFilename(tiezi.Title)

		// 分区名
		value_str, _ = jsonparser.GetString(resp.Bytes(), "forum_name")
		tiezi.Catelogy = value_str

		// 作者
		value_str, _ = jsonparser.GetString(resp.Bytes(), "tauthor")
		tiezi.Username = value_str

		// 作者id
		value_int, _ := jsonparser.GetInt(resp.Bytes(), "tauthorid")
		tiezi.UserId = cast.ToInt(value_int)

		// 总页数
		value_int, _ = jsonparser.GetInt(resp.Bytes(), "totalPage")
		tiezi.WebMaxPage = cast.ToInt(value_int)
		if CFGFILE_PAGE_DOWNLOAD_LIMIT > 0 && tiezi.WebMaxPage > (tiezi.LocalMaxPage+CFGFILE_PAGE_DOWNLOAD_LIMIT) {
			tiezi.WebMaxPage = tiezi.LocalMaxPage + CFGFILE_PAGE_DOWNLOAD_LIMIT
			page_download_limit_triggered = true
		}

		// 楼层数，楼主也算一层
		value_int, _ = jsonparser.GetInt(resp.Bytes(), "vrows")
		tiezi.FloorCount = cast.ToInt(value_int - 1)

		// 初始化floors个数
		if len(tiezi.Floors) == 0 {
			tiezi.Floors = make([]Floor, tiezi.FloorCount)
			for i := range tiezi.Floors {
				tiezi.Floors[i].Lou = -1
			}
		}
		value_byte, dataType, _, _ := jsonparser.Get(resp.Bytes(), "hot_post")
		if dataType != jsonparser.NotExist {
			tiezi.HotPosts.analyze(value_byte, false)
		}
		value_byte, _, _, _ = jsonparser.Get(resp.Bytes(), "result")
		tiezi.Floors.analyze(value_byte, false)
	}
}

/**
 * @description: 本地未生成过。初始化主楼和第一页
 * @param {int} tid 帖子tid
 * @return {*}
 */
func (tiezi *Tiezi) InitFromWeb(tid int, authorId int) {
	tiezi.init(tid, authorId)
	tiezi.Version = VERSION
	tiezi.Assets = map[string]string{}
	tiezi.LocalMaxPage = 1
	tiezi.LocalMaxFloor = -1
	log.Printf("下载第 %02d 页\n", tiezi.LocalMaxPage)
	tiezi.page(tiezi.LocalMaxPage)
}

/**
 * @description: 本地已经有生成过，现在来根据local信息来追加下载新楼层。
 * @param {int} tid 帖子tid
 * @return {*}
 */
func (tiezi *Tiezi) InitFromLocal(tid int, authorId int) {
	tiezi.init(tid, authorId)
	tiezi.Version = VERSION

	checkFileExistence := func(fileName string) {
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			log.Fatalln(fileName, "文件丢失，软件将退出。")
		}
	}
	folderName := FindFolderNameByTid(tid, authorId)
	if folderName == "" {
		log.Fatalln("找不到本地 tid 文件夹，软件将退出。")
	}
	processFileName := filepath.Join(".", folderName, "process.ini")
	checkFileExistence(processFileName)

	assetsFileName := filepath.Join(".", folderName, "assets.json")
	checkFileExistence(assetsFileName)

	jsonBytes, _ := os.ReadFile(assetsFileName)
	err := json.Unmarshal(jsonBytes, &(tiezi.Assets))
	if err != nil {
		log.Fatalln("解析 assets.json 失败:", err.Error())
	}
	cfg, _ := ini.Load(processFileName)
	tiezi.LocalMaxPage = cfg.Section("local").Key("max_page").MustInt(1)
	tiezi.LocalMaxFloor = cfg.Section("local").Key("max_floor").MustInt(-1)
	log.Printf("下载第 %02d 页\n", tiezi.LocalMaxPage)
	tiezi.page(tiezi.LocalMaxPage)

}

/**
 * @description: 初始化 Tiezi。
 * @param {int} tid 帖子tid
 * @return {*}
 */
func (tiezi *Tiezi) init(tid int, authorId int) {
	tiezi.Tid = tid
	tiezi.AuthorId = authorId
}

/**
 * @description: 由传入 Tiezi 对象里根据 pid 查找一 Floor 对象。若没有查到则返回空
 * @param {int} pid 楼层 pid
 * @return {*}
 */
func (tiezi *Tiezi) findFloorByPid(pid int) *Floor {
	for _, v := range tiezi.Floors {
		if v.Pid == pid {
			return &v
		}
	}
	if CFGFILE_ENHANCE_ORI_REPLY_ONLINE {
		// 可以开新的网络请求，只填充content
		resp, err := Client.R().SetFormData(map[string]string{
			"tid": cast.ToString(tiezi.Tid),
			"pid": cast.ToString(pid),
		}).Post("app_api.php?__lib=post&__act=list")
		if err != nil {
			log.Fatalln(err.Error())
		}
		code, _ := jsonparser.GetInt(resp.Bytes(), "code")
		if code != 0 {
			msg, _ := jsonparser.GetString(resp.Bytes(), "msg")
			log.Println("获取回复内容失败 nga返回代码不为0:", code, msg)
			return &Floor{Content: fmt.Sprint(pid, "获取回复内容失败 nga返回代码不为0:", code, msg)}
		}
		// 解析返回的楼层数据
		value_byte, dataType, _, _ := jsonparser.Get(resp.Bytes(), "result")
		if dataType == jsonparser.NotExist {
			log.Println("获取回复内容失败，result不存在")
			return &Floor{Content: fmt.Sprint(pid, "获取回复内容失败 result不存在")}
		}
		content, err := jsonparser.GetString(value_byte, "[0]", "content")
		if err != nil {
			log.Fatalln(err.Error())
		}
		// 尽量修大部分文本内容
		return &Floor{Content: fixMost(content, nil, nil)}
	}
	return nil
}

func fixMost(cont string, tiezi *Tiezi, floor *Floor) string {
	fixReplace := func(cont string) string {
		replacements := map[string]string{
			`\u0026`:             "&",
			`\u003c`:             "<",
			`\u003e`:             ">",
			`&amp;#160;`:         " ",
			`<br/>`:              "\n",
			`<br>`:               "\n",
			`&lt;br/&gt;`:        "\n",
			`&lt;br&gt;`:         "\n",
			`<del class='gray'>`: `~~`,
			`</del>`:             `~~`,
		}
		for old, new := range replacements {
			cont = strings.ReplaceAll(cont, old, new)
		}
		return cont
	}
	fixAnony := func(cont string, floor *Floor) string {
		// 匿名
		if floor != nil {
			if len(floor.Username) > 7 && floor.Username[:7] == `#anony_` {
				floor.Username = anony(floor.Username)
			}
		}
		re := regexp.MustCompile(`#anony_.{32}`)
		for _, it := range re.FindAllString(cont, -1) {

			cont = strings.ReplaceAll(cont, it, anony(it))
		}
		return cont
	}
	fixDice := func(cont string) string {
		// ROLL DICE
		// <div class='dice'><b>ROLL : 1d100</b>=d100(32)=<b>32</b></div>
		re := regexp.MustCompile(`<div class='dice'><b>ROLL : (.+?)</b>=(.+?)=<b>(.+?)</b></div>`)
		for _, it := range re.FindAllStringSubmatch(cont, -1) {
			rollSrc := it[1]
			rollRt := it[3]
			cont = strings.ReplaceAll(cont, it[0], fmt.Sprintf(" **【ROLL** : %s= **%s】** ", rollSrc, rollRt))
		}
		return cont
	}
	fixCollapse := func(cont string) string {
		// collapse 折叠
		// <div class="foldBox no"><div class="collapse_btn"><a href="javascript:;" onclick="collapse(this);">+</a>外层看到的 ...</div><span class="collapse_content" id="foldCnt">里面</span></div>
		re := regexp.MustCompile(`<div class="foldBox no"><div class="collapse_btn"><a href="javascript:;" onclick="collapse\(this\);">\+</a>(.+?) ...</div><span class="collapse_content" id="foldCnt">(.+?)</span></div>`)
		for _, it := range re.FindAllStringSubmatch(cont, -1) {
			outTxt := it[1]
			inTxt := it[2]
			rt := fmt.Sprintf("<details>\n  <summary>%s</summary>\n  <pre>%s</pre>\n</details>", outTxt, strings.ReplaceAll(inTxt, "\n", "<br>"))

			cont = strings.ReplaceAll(cont, it[0], rt)
		}
		return cont
	}
	fixSmile := func(cont string) string {
		re := regexp.MustCompile(`\[s\:.+?\:.+?\]`)
		for _, it := range re.FindAllString(cont, -1) {
			if !CFGFILE_USE_LOCAL_SMILE_PIC {
				cont = strings.ReplaceAll(cont, it, `![`+strings.Split(it, `:`)[2]+`(https://img4.nga.178.com/ngabbs/post/smile/`+strings.ReplaceAll(getSmile(it), `"`, ``)+`)`)
			} else {
				smile_name := strings.Split(it, `:`)[1] + strings.TrimRight(strings.Split(it, `:`)[2], "]")
				if strings.Contains(smile_name, "web") {
					smile_name = smile_name + ".gif"
				} else {
					smile_name = smile_name + ".png"
				}
				final := filepath.Join(CFGFILE_LOCAL_SMILE_PIC_PATH, smile_name)
				cont = strings.ReplaceAll(cont, it, `![`+strings.Split(it, `:`)[2]+`(`+final+`)`)
			}
		}
		return cont
	}
	fixUrl := func(cont string) string {
		// 超链接
		re := regexp.MustCompile(`\[url\](.+?)\[/url\]`)
		for _, it := range re.FindAllStringSubmatch(cont, -1) {
			cont = strings.ReplaceAll(cont, `[url]`+it[1]+`[/url]`, `[url](`+it[1]+`)`)
		}
		re = regexp.MustCompile(`\[url=(.+?)\](.+?)\[/url\]`)
		for _, it := range re.FindAllStringSubmatch(cont, -1) {
			cont = strings.ReplaceAll(cont, `[url=`+it[1]+`]`+it[2]+`[/url]`, `[`+it[2]+`](`+it[1]+`)`)
		}
		return cont
	}
	fixQuote := func(cont string, floor *Floor) string {
		// 引用
		// 下列的[b] 和[/b] 在这个接口下好像都变成了 <b> 和 </b>
		// 圈主贴
		// (?s) 意思单行模式
		reg_str := `(?s)\[quote\]\[tid=.+?Post by \[uid.*?\](.+)\[\/uid\].*?\((\d{4}.+?)\):</b>(.+?)\[/quote\]((?:\n){0,2})`
		if !strings.Contains(cont, "uid=") {
			// 匿名回复，没有uid
			reg_str = `(?s)\[quote\]\[tid=.+?Post by (.+)<span .*?\((\d{4}.+?)\):</b>(.+?)\[/quote\]((?:\n){0,2})`
		}
		re := regexp.MustCompile(reg_str)
		// [1]人名 [2]时间 [3]圈的内容
		for _, it := range re.FindAllStringSubmatch(cont, -1) {
			quoteText := strings.ReplaceAll(it[3], "\n", "\n>")
			quoteAuthor := it[1]
			quoteTime := it[2]
			if len(quoteAuthor) > 7 && quoteAuthor[:7] == `#anony_` {
				quoteAuthor = anony(quoteAuthor)
			} else {
				// 拼一下，以拿到uid
				reg_str = `\[uid=(\d+?)\]` + regexp.QuoteMeta(quoteAuthor) + `\[\/uid\]`
				re = regexp.MustCompile(reg_str)
				it := re.FindStringSubmatch(cont)
				if len(it) >= 2 {
					quoteAuthor = fmt.Sprintf("%s(%s)", quoteAuthor, it[1])
				}
			}
			cont = strings.ReplaceAll(cont, it[0], `>[jump](#pid0) `+quoteAuthor+`(`+quoteTime+`)`+` 说: `+quoteText+"\n\n")
			if floor != nil {
				floor.AppendPid = append(floor.AppendPid, 0)
			}
		}

		// 圈其他楼
		// [quote][pid=684818389,36006627,6]Reply[/pid] <b>Post by [uid=42264644]Lian君并没有名字[/uid] (2023-04-18 15:46):</b><br/><br/>有发表暴论的机会？那我也来整几个[/quote]1.后宫和
		// "[quote][pid=684833266,36008480,1]Reply[/pid] <b>Post by [uid=64575408]虹色的棉花糖[/uid] (2023-04-18 16:54):</b>\n\n上海呢[/quote]"
		// [quote][pid=684810015,36006627,5]Reply[/pid] <b>Post by 庚雷尤甲项季<span class=\"gray\">(83楼)</span> (2023-04-18 15:08):</b><br/><br/>阿巴[/quote]
		quoteCount := strings.Count(cont, "[quote]")
		for range quoteCount {
			//最内层的quote下标
			quoteStartIndex := strings.LastIndex(cont, "[quote]")
			if quoteStartIndex < 0 {
				break
			}
			quoteEndIndex := quoteStartIndex + strings.Index(cont[quoteStartIndex:], "[/quote]")
			if quoteEndIndex < 0 || quoteStartIndex >= quoteEndIndex+8 {
				break
			}
			clip := cont[quoteStartIndex : quoteEndIndex+8]

			reg_str := `(?s)\[quote\]\[pid=(\d+?),.+?Post by \[uid.*?\](.+)\[\/uid\].*?\((\d{4}.+?)\):</b>(.+?)\[/quote\]((?:\n){0,2})`
			if !strings.Contains(clip, "uid=") {
				// 匿名回复，没有uid
				reg_str = `(?s)\[quote\]\[pid=(\d+?),.+?Post by (.+)<span .*?\((\d{4}.+?)\):</b>(.+?)\[/quote\]((?:\n){0,2})`
			}
			re = regexp.MustCompile(reg_str)
			// [1]pid [2]原作者 [3]时间 [4]说的东西
			for _, it := range re.FindAllStringSubmatch(clip, -1) {
				cont = strings.ReplaceAll(cont, `[url=`+it[1]+`]`+it[2]+`[/url]`, `[`+it[2]+`](`+it[1]+`)`)
				quoteText := strings.ReplaceAll(it[4], "\n", "\n>")
				quoteAuthor := it[2]
				quotePid := it[1]
				quoteTime := it[3]
				if len(quoteAuthor) > 7 && quoteAuthor[:7] == `#anony_` {
					quoteAuthor = anony(quoteAuthor)
				} else {
					// 拼一下，以拿到uid
					reg_str = `\[uid=(\d+?)\]` + regexp.QuoteMeta(quoteAuthor) + `\[\/uid\]`
					re = regexp.MustCompile(reg_str)
					it := re.FindStringSubmatch(cont)
					if len(it) >= 2 {
						quoteAuthor = fmt.Sprintf("%s(%s)", quoteAuthor, it[1])
					}
				}
				cont = strings.ReplaceAll(cont, it[0], `>[jump](#pid`+quotePid+`) `+quoteAuthor+`(`+quoteTime+`)`+` 说: `+quoteText+"\n\n")
				// 这里会有原文的，就不append了
			}
		}
		return cont
	}
	fixReply := func(cont string, tiezi *Tiezi, floor *Floor) string {
		// 回复
		reg_str := `(?s)<b>Reply to \[tid=(\d+?).+? Post by \[uid.*?\](.+)\[\/uid\].+?\((.+?)\)</b>((?:\n){0,2})`
		if !strings.Contains(cont, "uid=") {
			// 匿名回复，没有uid
			reg_str = `(?s)<b>Reply to \[tid=(\d+?).+? Post by (.+)<span .+?\((.+?)\)</b>((?:\n){0,2})`
		}
		re := regexp.MustCompile(reg_str)
		// 评论主楼[1]pid [2]原作者 [3]时间
		for _, it := range re.FindAllStringSubmatch(cont, -1) {
			quoteAuthor := it[2]
			quoteTime := it[3]
			if len(quoteAuthor) > 7 && quoteAuthor[:7] == `#anony_` {
				quoteAuthor = anony(quoteAuthor)
			} else {
				// 拼一下，以拿到uid
				reg_str = `\[uid=(\d+?)\]` + regexp.QuoteMeta(quoteAuthor) + `\[\/uid\]`
				re = regexp.MustCompile(reg_str)
				it := re.FindStringSubmatch(cont)
				if len(it) >= 2 {
					quoteAuthor = fmt.Sprintf("%s(%s)", quoteAuthor, it[1])
				}
			}
			cont = strings.ReplaceAll(cont, it[0], `>[jump](#pid0) `+quoteAuthor+`(`+quoteTime+"):\n\n")
		}
		reg_str = `(?s)<b>Reply to \[pid=(\d+?),.+? Post by \[uid.*?\](.+)\[\/uid\].+?\((.+?)\)</b>((?:\n){0,2})`
		if !strings.Contains(cont, "uid=") {
			// 匿名回复，没有uid
			reg_str = `(?s)<b>Reply to \[pid=(\d+?),.+? Post by (.+)<span .+?\((.+?)\)</b>((?:\n){0,2})`
		}
		re = regexp.MustCompile(reg_str)
		// [1]pid [2]原作者 [3]时间
		for _, it := range re.FindAllStringSubmatch(cont, -1) {
			quotePid := it[1]
			quoteAuthor := it[2]
			quoteTime := it[3]
			if len(quoteAuthor) > 7 && quoteAuthor[:7] == `#anony_` {
				quoteAuthor = anony(quoteAuthor)
			} else {
				// 拼一下，以拿到uid
				reg_str = `\[uid=(\d+?)\]` + regexp.QuoteMeta(quoteAuthor) + `\[\/uid\]`
				re = regexp.MustCompile(reg_str)
				it := re.FindStringSubmatch(cont)
				if len(it) >= 2 {
					quoteAuthor = fmt.Sprintf("%s(%s)", quoteAuthor, it[1])
				}
			}
			replyedText := ":"
			if tiezi != nil && CFGFILE_ENHANCE_ORI_REPLY {
				// 正常的从fixContent来调用
				replyedFloor := tiezi.findFloorByPid(cast.ToInt(quotePid))
				if replyedFloor != nil {
					replyedText = "说:\n>" + strings.ReplaceAll(replyedFloor.Content, "\n", "\n>")
				}
			} else if tiezi == nil {
				// 可能是ENHANCE_ONLINE调用过来的
			}

			cont = strings.ReplaceAll(cont, it[0], `>[jump](#pid`+quotePid+`) `+quoteAuthor+`(`+quoteTime+")"+replyedText+"\n\n")
			if floor != nil {
				floor.AppendPid = append(floor.AppendPid, cast.ToInt(quotePid))
			}
		}
		return cont
	}
	cont = fixReplace(cont)
	cont = fixAnony(cont, floor)
	cont = fixDice(cont)
	cont = fixCollapse(cont)
	cont = fixSmile(cont)
	cont = fixUrl(cont)
	cont = fixQuote(cont, floor)
	cont = fixReply(cont, tiezi, floor)
	return cont
}

/**
 * @description: 由bbcode转md，以及下载图片、转化表情等
 * @param {int} floor_i floor下标
 * @return {*}
 */
func (tiezi *Tiezi) fixContent(floor_i int) {
	/*此接口(app_api)与旧接口不太相同，有些源码格式和网页端看到的不一样！
	 *1. 疑似匿名直接显示
	 *2. 删除线有变
	 *3. quote reply等，[b]变化；假如是匿名用户，就不会有 uid框框
	 */
	// tid int, assets *(map[string]string)
	assets := &tiezi.Assets
	oriFloor := &tiezi.Floors[floor_i]
	floor := &tiezi.Floors[floor_i]
	curCommentI := -1

	// 循环尾部有判断是否有comments且是否进去的操作，请注意
	for {
		// 假如要获取IP位置则在此处获取
		if CFGFILE_GET_IP_LOCATION {
			resp, err := Client.R().SetFormData(map[string]string{
				"uid": cast.ToString(floor.UserId),
			}).Post("nuke.php?__lib=ucp&__act=get&__output=8")
			if err != nil {
				log.Println(err.Error())
			} else {
				value_str, err := jsonparser.GetString(resp.Bytes(), "data", "0", "ipLoc")
				if err != nil {
					log.Println("获取用户IP位置失败: " + err.Error())
				} else {
					floor.IpLocation = value_str
				}
			}
		}
		// 获取IP位置结束
		cont := floor.Content
		cont = fixMost(cont, tiezi, floor)
		// 视频
		cont = processMedia(cont, `<span class="video">(<video[^>]*>.*?</video>)</span>`, `src="([^"]+)"`, "视频", assets, floor, tiezi, false)
		// 音频
		cont = processMedia(cont, `<span class="audio" onclick="audioClick\(event\)"> <audio src="([^"]+)"[^>]*></audio></span>`, `src="([^"]+)"`, "音频", assets, floor, tiezi, false)
		// 图片
		cont = processMedia(cont, `\[img\](.+?)\[/img\]`, "", "", assets, floor, tiezi, true)

		floor.Content = cont
		//到这里，fix已经结束了

		//判断是否有、是否有剩余下挂comment需要处理
		if curCommentI+1 < len(oriFloor.Comments) {
			floor = &oriFloor.Comments[curCommentI+1]
			curCommentI++
		} else {
			break
		}
	}

}

/**
 * @description: 对fixContent的包裹。主要是为了并行……
 * @param {int} startFloor_i 从哪一下标开始修。主要是针对追加楼层更新时
 * @return {*}
 */
func (tiezi *Tiezi) fixFloorContent(startFloor_i int) {

	var wg sync.WaitGroup
	p, _ := ants.NewPoolWithFunc(CFGFILE_THREAD_COUNT, func(floor_i interface{}) {
		if tiezi.Floors[cast.ToInt(floor_i)].Lou != -1 {
			responseChannel <- fmt.Sprintf("开始修正第 %02d 楼层", cast.ToInt(floor_i))
			tiezi.fixContent(cast.ToInt(floor_i))
		}
		wg.Done()
	})
	defer p.Release()

	startTime := time.Now()
	for i := startFloor_i; i < len(tiezi.Floors); i++ {
		wg.Add(1)
		_ = p.Invoke(i)
		tiezi.Timestamp = ts()
	}
	wg.Wait()
	log.Println("修正楼层总耗时:", time.Since(startTime).Truncate(time.Second).String())
	// 如果为了调试取消并行的话，上述代码均注释，换成下面的
	// for i := startFloor_i; i < len(tiezi.Floors); i++ {
	// 	log.Printf("开始修正第 %02d 楼层", i)
	// 	tiezi.fixContent(cast.ToInt(i))
	// }

}

/**
 * @description: 写markdown文件
 * @param {int} localMaxFloor 本地已有的楼
 * @return {*}
 */
func (tiezi *Tiezi) genMarkdown(localMaxFloor int) {
	folder := filepath.Join(".", tiezi.GetNeededFolderName())
	os.MkdirAll(folder, os.ModePerm)

	splitInfoPath := filepath.Join(folder, "splitinfo.ini")

	// 判断一下 md 文件切分的缓存文件
	// 当前需要写入的文件尾标，如1则写 name-001.md
	curSplitFileSuffixNum := 1
	// 当前剩余的层数，>0。若为小于等于0则关闭此功能
	curSplitFloorLeft := -1
	// 为了处理曾经切但现在不切的情况（按切但是=1来处理）
	localCFGFILE_SPLIT_MD_FILE := CFGFILE_SPLIT_MD_FILE
	if _, err := os.Stat(splitInfoPath); os.IsNotExist(err) {
		// 文件不存在，若有已知的未切分md文件，意味着本md文件不用切割，写入原先的大文件夹即可
		// 若不存在已知的未切分文件，则说明是新任务
	} else {
		// 文件存在，需要切分
		cfg, err := ini.Load(splitInfoPath)
		if err != nil {
			log.Fatalln("无法加载 splitinfo.ini:", err)
		}

		// 检查配置是否发生了变化，如果变动了则提示，然后使用新的参数
		// 如果最新配置是不切分，但存在splitinfo.ini文件，说明之前是切分模式，继续使用之前的切分模式，切分参数临时调为1
		prevSplitSetting := cfg.Section("split").Key("setting_value").MustInt(1)
		if prevSplitSetting != CFGFILE_SPLIT_MD_FILE {
			// 配置发生了变化，提示后使用新的设置
			log.Println("此文件夹存在旧切分配置(", prevSplitSetting, ")和当前配置(", CFGFILE_SPLIT_MD_FILE, ")不符，将使用新切分配置")
			if CFGFILE_SPLIT_MD_FILE < 1 {
				log.Println("当前配置为关闭切分(-1)，但本任务会按照=1页来进行切分。即只要曾经切分过，则后续仍切分")
				localCFGFILE_SPLIT_MD_FILE = 1
			}
		}
		curSplitFileSuffixNum = cfg.Section("split").Key("file_suffix").MustInt()
		// 最小为1，即往旧文件写一楼然后开新文件
		curSplitFloorLeft = max(min(cfg.Section("split").Key("floor_left").MustInt(), localCFGFILE_SPLIT_MD_FILE*20), 1)
	}

	// 首先判断是否存在 post.md 或者 个性化md (无后缀)，若有则继续沿用
	// 最传统的
	fileExists := true
	mdName := "post.md"
	mdFilePath := filepath.Join(folder, mdName)
	if _, err := os.Stat(mdFilePath); os.IsNotExist(err) {
		fileExists = false
	} else {
		fileExists = true
		// 存在最传统的post.md，后续为post.md
	}

	if !fileExists && CFGFILE_USE_TITLE_AS_MD_FILE_NAME {
		// 最传统的不存在，且开启了个性化，判断是否存在个性化无后缀
		mdName = fmt.Sprintf("%s.md", tiezi.TitleFolderSafe)
		mdFilePath = filepath.Join(folder, mdName)
		if _, err := os.Stat(mdFilePath); os.IsNotExist(err) {
			fileExists = false
		} else {
			fileExists = true
			// 存在个性化无后缀，后续为此类型
		}
	}

	rawName := "post"
	if !fileExists && (curSplitFloorLeft > 0 || localCFGFILE_SPLIT_MD_FILE > 0) {
		// 以上两种没有后缀的都不存在，且开启了切分，则需要设定为带切分的了
		if curSplitFloorLeft < 1 {
			// 说明上面没有读到splitinfo.ini文件，需要设置为配置值
			// 一页=20楼
			curSplitFloorLeft = localCFGFILE_SPLIT_MD_FILE * 20
		}
		rawName = "post"
		if CFGFILE_USE_TITLE_AS_MD_FILE_NAME {
			rawName = tiezi.TitleFolderSafe
		}
		mdName = fmt.Sprintf("%s-%03d.md", rawName, curSplitFileSuffixNum)
		mdFilePath = filepath.Join(folder, mdName)
	}
	if _, err := os.Stat(mdFilePath); os.IsNotExist(err) {
		if _, err := os.Create(mdFilePath); err != nil {
			log.Fatalf("创建 .md 文件失败：%v", err)
		}
	}

	f, err := os.OpenFile(mdFilePath, os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("打开 .md 文件失败：%v", err)
	}
	defer f.Close()
	for i := localMaxFloor; i < len(tiezi.Floors); i++ {
		floor := &tiezi.Floors[i]
		if floor.Lou == -1 {
			// 被抽楼了
			continue
		}

		if curSplitFloorLeft == 0 && localCFGFILE_SPLIT_MD_FILE > 0 {
			// 当前文件已经写完了，需要开新文件了（只有在切分模式下才创建新文件）
			f.Close()
			curSplitFileSuffixNum = curSplitFileSuffixNum + 1
			curSplitFloorLeft = localCFGFILE_SPLIT_MD_FILE * 20
			mdName = fmt.Sprintf("%s-%03d.md", rawName, curSplitFileSuffixNum)
			mdFilePath = filepath.Join(folder, mdName)
			f, err = os.OpenFile(mdFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
			if err != nil {
				log.Fatalf("创建或打开 .md 文件失败：%v", err)
			}
			// 补一个这个，防止把顶的`----`识别成注释，导致无法渲染第一个楼层
			f.WriteString(" ")
		}

		if floor.Pid == 0 {
			var authorIdOptText = ""
			if tiezi.AuthorId > 0 {
				authorIdOptText = fmt.Sprintf("-只看 %d", tiezi.AuthorId)
			}
			_, _ = f.WriteString(fmt.Sprintf("### %s%s\n\nMade by ngapost2md (c) ludoux [GitHub Repo](https://github.com/ludoux/ngapost2md)\n\n", tiezi.Title, authorIdOptText))
		}

		if floor.Pid == 0 && len(tiezi.HotPosts) > 0 {
			_, _ = f.WriteString("##### 热门回复\n\n")
			for _, v := range tiezi.HotPosts {
				if v.Lou == -1 {
					continue
				}
				content := v.Content
				if len([]rune(content)) > 22 {
					content = string([]rune(content)[:20]) + "..."
				}
				_, _ = f.WriteString(fmt.Sprintf("- [%d楼](#pid%d): %s\n", v.Lou, v.Pid, content))
			}
			_, _ = f.WriteString("\n")
		}

		IpLocationStr := ""
		if floor.IpLocation != "" {
			IpLocationStr = "\\(" + floor.IpLocation + "\\)"
		}

		_, _ = f.WriteString(fmt.Sprintf("----\n\n##### <span id=\"pid%d\">%d.[%d] \\<pid:%d\\> %s by %s(%d)%s</span>\n%s", floor.Pid, floor.Lou, floor.LikeNum, floor.Pid, ts2t(floor.Timestamp), floor.Username, floor.UserId, IpLocationStr, floor.Content))

		if floor.Comments != nil {
			_, _ = f.WriteString("\n\n*---下挂评论---*")
			for _, comment := range floor.Comments {
				if comment.Lou <= 0 {
					// 为了评论从1楼开始，评论[0]恒为为空
					continue
				}
				_, _ = f.WriteString(fmt.Sprintf("\n\n%d.[%d] \\<pid:%d\\>%s by %s(%d):\n%s", comment.Lou, comment.LikeNum, comment.Pid, ts2t(comment.Timestamp), comment.Username, comment.UserId, comment.Content))
			}
		}

		_, _ = f.WriteString("\n\n")
		if curSplitFloorLeft > 0 {
			curSplitFloorLeft = curSplitFloorLeft - 1
		}
	}

	// 要考虑为0时的情况，为0时也是关闭了这个功能
	if localCFGFILE_SPLIT_MD_FILE > 0 {
		fileName := filepath.Join(folder, "splitinfo.ini")
		cfg := ini.Empty()
		cfg.NewSection("split")
		cfg.Section("split").NewKey("file_suffix", cast.ToString(curSplitFileSuffixNum))
		cfg.Section("split").NewKey("floor_left", cast.ToString(curSplitFloorLeft))
		cfg.Section("split").NewKey("setting_value", cast.ToString(localCFGFILE_SPLIT_MD_FILE)) // 记录当前配置值
		cfg.SaveTo(fileName)
	}
}

func responseController() {
	for rc := range responseChannel {
		log.Println(rc)
	}
}

// 会首先调用FindFolderNameByTid，确定本地没有相关文件夹再返回指定格式文件名。否则返回本地已有文件名
func (tiezi *Tiezi) GetNeededFolderName() string {
	already := FindFolderNameByTid(tiezi.Tid, tiezi.AuthorId)
	if already != "" {
		return already
	}
	if CFGFILE_USE_TITLE_AS_FOLDER_NAME {
		if tiezi.AuthorId > 0 {
			return fmt.Sprintf("%d(%d)-%s", tiezi.Tid, tiezi.AuthorId, tiezi.TitleFolderSafe)
		} else {
			return fmt.Sprintf("%d-%s", tiezi.Tid, tiezi.TitleFolderSafe)
		}
	} else {
		if tiezi.AuthorId > 0 {
			return fmt.Sprintf("%d(%d)", tiezi.Tid, tiezi.AuthorId)
		} else {
			return cast.ToString(tiezi.Tid)
		}
	}
}

func (tiezi *Tiezi) SaveProcessInfo() {
	folder := filepath.Join(".", tiezi.GetNeededFolderName())

	fileName := filepath.Join(folder, "process.ini")
	cfg := ini.Empty()
	cfg.NewSection("local")
	cfg.Section("local").NewKey("max_floor", cast.ToString(tiezi.LocalMaxFloor))
	cfg.Section("local").NewKey("max_page", cast.ToString(tiezi.LocalMaxPage))
	cfg.SaveTo(fileName)
}

func (tiezi *Tiezi) SaveAssetsMap() {
	folder := filepath.Join(".", tiezi.GetNeededFolderName())

	fileName := filepath.Join(folder, "assets.json")
	result, err := json.Marshal(tiezi.Assets)
	if err != nil {
		log.Fatalln("将附件转化为 Json 格式失败:", err.Error())
	}
	f, _ := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0666)
	_, err = f.Write(result)
	if err != nil {
		log.Fatalln("保存 assets.json 文件失败:", err.Error())
	}
	defer f.Close()
}

func (tiezi *Tiezi) Download() {
	if tiezi.Tid != 0 {
		var wg sync.WaitGroup
		p, _ := ants.NewPoolWithFunc(CFGFILE_THREAD_COUNT, func(page interface{}) {
			time.Sleep(time.Millisecond * time.Duration(DELAY_MS))
			responseChannel <- fmt.Sprintf("下载第 %02d 页", page)
			// 1. 并行下载page
			tiezi.page(cast.ToInt(page))
			wg.Done()
		})
		defer p.Release()
		go responseController()

		startTime := time.Now()
		// 因为 it.LocalMaxPage 在InitFromxxx的时候已经page过了
		for page := tiezi.LocalMaxPage + 1; page <= tiezi.WebMaxPage; page++ {
			wg.Add(1)
			_ = p.Invoke(page)
		}
		wg.Wait()

		log.Println("下载所有页面总耗时:", time.Since(startTime).Truncate(time.Second).String())
		if page_download_limit_triggered {
			log.Println("单次下载 Page 数已达上限！本次导出完毕后需要多次重新运行才可全部导出此帖内容。")
		}

		// 2. 格式化content
		tiezi.fixFloorContent(tiezi.LocalMaxFloor + 1)

		// 3. 制作文件
		tiezi.genMarkdown(tiezi.LocalMaxFloor + 1)

		tiezi.LocalMaxPage = tiezi.WebMaxPage

		// 因为NGA会抽楼，floorcount不准，只能这样子
		for i := len(tiezi.Floors) - 1; ; i-- {
			floor := &tiezi.Floors[i]
			if floor.Lou > -1 {
				tiezi.LocalMaxFloor = floor.Lou
				break
			}
		}
		// 存储tiezi---暂时注释掉，还是使用存储localmaxpage和maxfloor(SaveProcessInfo)的方法。
		// tiezi.SaveAsFile()

		// 存储localmaxpage和maxfloor
		tiezi.SaveProcessInfo()

		// 存储assets map
		tiezi.SaveAssetsMap()
		if page_download_limit_triggered {
			log.Println("单次下载 Page 数已达上限！本次导出完毕后需要多次重新运行才可全部导出此帖内容。")
		}
		log.Println("本次任务结束。")
	}
}
