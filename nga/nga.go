package nga

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/buger/jsonparser"
	"github.com/panjf2000/ants/v2"
	"github.com/spf13/cast"
	"gopkg.in/ini.v1"
)

var (
	VERSION         = "NEO1"
	THREAD_COUNT    = 2
	DELAY_MS        = 330
	GET_IP_LOCATION = false
	mutex           sync.Mutex
)

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
	Tid           int
	Title         string
	Catelogy      string
	Username      string
	UserId        int
	WebMaxPage    int
	LocalMaxPage  int
	LocalMaxFloor int
	FloorCount    int    //包含主楼
	Floors        Floors //主楼为[0]
	Timestamp     int64  //page() fixFloorContent()  中会更新
	Version       string
	Assets        map[string]string
}

var responseChannel = make(chan string, 15)

func (it *Floors) ana(resp []byte) {

	jsonparser.ArrayEach(resp, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		value_int, _ := jsonparser.GetInt(value, "lou")
		lou := cast.ToInt(value_int)

		//根据楼数补充Floors
		for len(*it) < lou+1 {
			(*it) = append((*it), Floor{Lou: -1})
		}

		curFloor := &(*it)[lou]

		//楼层
		curFloor.Lou = lou

		//PID
		value_int, _ = jsonparser.GetInt(value, "pid")
		curFloor.Pid = cast.ToInt(value_int)

		//时间戳
		value_int, _ = jsonparser.GetInt(value, "postdatetimestamp")
		curFloor.Timestamp = value_int

		//用户名
		value_str, _ := jsonparser.GetString(value, "author", "username")
		curFloor.Username = value_str

		//用户id
		value_int, _ = jsonparser.GetInt(value, "author", "uid")
		curFloor.UserId = cast.ToInt(value_int)

		//内容
		value_str, _ = jsonparser.GetString(value, "content")
		curFloor.Content = value_str

		//点赞数
		value_int, _ = jsonparser.GetInt(value, "vote_good")
		curFloor.LikeNum = cast.ToInt(value_int)

		//下挂comments
		value_byte, dataType, _, _ := jsonparser.Get(value, "comments")
		if dataType != jsonparser.NotExist {
			curFloor.Comments.ana(value_byte)
		}

	})
}

func (it *Tiezi) page(page int) {
	resp, err := Client.R().SetFormData(map[string]string{
		"page": cast.ToString(page),
		"tid":  cast.ToString(it.Tid),
	}).Post("app_api.php?__lib=post&__act=list")
	if err != nil {
		log.Println(err.Error())
	}
	code, _ := jsonparser.GetInt(resp.Bytes(), "code")
	if code != 0 {
		log.Panic("返回code不为0")
	} else {
		it.Timestamp = ts()

		//标题
		value_str, _ := jsonparser.GetString(resp.Bytes(), "tsubject")
		it.Title = value_str

		//分区名
		value_str, _ = jsonparser.GetString(resp.Bytes(), "forum_name")
		it.Catelogy = value_str

		//作者
		value_str, _ = jsonparser.GetString(resp.Bytes(), "tauthor")
		it.Username = value_str

		//作者id
		value_int, _ := jsonparser.GetInt(resp.Bytes(), "tauthorid")
		it.UserId = cast.ToInt(value_int)

		//总页数
		value_int, _ = jsonparser.GetInt(resp.Bytes(), "totalPage")
		it.WebMaxPage = cast.ToInt(value_int)

		//楼层数，楼主也算一层
		value_int, _ = jsonparser.GetInt(resp.Bytes(), "vrows")
		it.FloorCount = cast.ToInt(value_int - 1)

		//初始化floors个数
		if it.Floors == nil || len(it.Floors) == 0 {
			it.Floors = make([]Floor, it.FloorCount)
			for i := range it.Floors {
				it.Floors[i].Lou = -1
			}
		}

		value_byte, _, _, _ := jsonparser.Get(resp.Bytes(), "result")
		it.Floors.ana(value_byte)
	}
}

// 初始化主楼和第一页
func (it *Tiezi) InitFromWeb(tid int) {
	it.init(tid, true)
	it.Version = VERSION
	it.Assets = map[string]string{}
	it.LocalMaxPage = 1
	it.LocalMaxFloor = -1
	fmt.Printf("Start page %d\n", it.LocalMaxPage)
	it.page(it.LocalMaxPage)
}

// 本地已经有生成过，现在来根据local信息来追加下载新楼层。
func (it *Tiezi) InitFromLocal(tid int) {
	it.init(tid, false)
	it.Version = VERSION

	processFileName := `./` + cast.ToString(it.Tid) + `/process.ini`
	//倘若丢失process文件，报错并退出
	if _, err := os.Stat(processFileName); os.IsNotExist(err) {
		fmt.Println("process.ini 文件丢失，软件将退出。")
		os.Exit(1)
	}

	assetsFileName := `./` + cast.ToString(it.Tid) + `/assets.json`
	//倘若丢失assets文件，报错并退出
	if _, err := os.Stat(assetsFileName); os.IsNotExist(err) {
		fmt.Println("assets.json 文件丢失，软件将退出。")
		os.Exit(1)
	}

	jsonBytes, _ := os.ReadFile(assetsFileName)
	json.Unmarshal(jsonBytes, &(it.Assets))

	cfg, _ := ini.Load(processFileName)
	it.LocalMaxPage = cfg.Section("local").Key("max_page").MustInt(1)
	it.LocalMaxFloor = cfg.Section("local").Key("max_floor").MustInt(-1)
	fmt.Printf("Start page %d\n", it.LocalMaxPage)
	it.page(it.LocalMaxPage)

}

func (it *Tiezi) init(tid int, crateDict bool) {
	if crateDict {
		os.MkdirAll(`./`+cast.ToString(tid), os.ModePerm)
	}

	it.Tid = tid
}

func (tiezi *Tiezi) fixContent(floor_i int) {
	/*此接口(app_api)与旧接口不太相同，有些源码格式和网页端看到的不一样！
	 *1. 疑似匿名直接显示
	 *2. 删除线有变
	 *3. quote reply等，[b]变化；假如是匿名用户，就不会有 uid框框
	 */
	//tid int, assets *(map[string]string)
	tid := tiezi.Tid
	assets := &tiezi.Assets
	oriFloor := &tiezi.Floors[floor_i]
	floor := &tiezi.Floors[floor_i]
	curCommentI := -1

	//循环尾部有判断是否有comments且是否进去的操作，请注意
	for {
		//假如要获取IP位置则在此处获取
		if GET_IP_LOCATION {
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
		//获取IP位置结束

		cont := floor.Content
		cont = strings.ReplaceAll(cont, `\u0026`, "&")
		cont = strings.ReplaceAll(cont, `\u003c`, "<")
		cont = strings.ReplaceAll(cont, `\u003e`, ">")
		cont = strings.ReplaceAll(cont, `&amp;#160;`, " ")

		//换行
		cont = strings.ReplaceAll(cont, `<br/>`, "\n")
		cont = strings.ReplaceAll(cont, `<br>`, "\n")
		cont = strings.ReplaceAll(cont, `&lt;br/&gt;`, "\n")
		cont = strings.ReplaceAll(cont, `&lt;br&gt;`, "\n")

		//匿名
		if len(floor.Username) > 7 && floor.Username[:7] == `#anony_` {
			floor.Username = anony(floor.Username)
		}
		re := regexp.MustCompile(`#anony_.{32}`)
		for _, it := range re.FindAllString(cont, -1) {

			cont = strings.ReplaceAll(cont, it, anony(it))
		}

		//图片
		re = regexp.MustCompile(`\[img\](.+?)\[/img\]`)
		for _, it := range re.FindAllStringSubmatch(cont, -1) {
			url := it[1]
			if url[0:2] == "./" {
				url = "https://img.nga.178.com/attachments/" + url[2:]
			}
			url = strings.ReplaceAll(url, ".medium.jpg", "")
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
				fmt.Printf("down pic: %v\n", fileName)
				downloadAssets(url, `./`+cast.ToString(tid)+`/`+fileName)
				fmt.Printf("down pic: %v------ok\n", fileName)
			} else {
				mutex.Unlock()
			}
			cont = strings.ReplaceAll(cont, `[img]`+it[1]+`[/img]`, `![img](./`+fileName+`)`)
		}

		//表情
		re = regexp.MustCompile(`\[s\:.+?\:.+?\]`)
		for _, it := range re.FindAllString(cont, -1) {
			cont = strings.ReplaceAll(cont, it, `![`+strings.Split(it, `:`)[2]+`(https://img4.nga.178.com/ngabbs/post/smile/`+strings.ReplaceAll(getSmile(it), `"`, ``)+`)`)
		}

		//删除线
		//这是一个不一样的 原本是[del] [/del]（而且原来无空格）
		cont = strings.ReplaceAll(cont, `<del class='gray'> `, `~~`)
		cont = strings.ReplaceAll(cont, ` </del>`, `~~`)

		//超链接
		re = regexp.MustCompile(`\[url\](.+?)\[/url\]`)
		for _, it := range re.FindAllStringSubmatch(cont, -1) {
			cont = strings.ReplaceAll(cont, `[url]`+it[1]+`[/url]`, `[url](`+it[1]+`)`)
		}
		re = regexp.MustCompile(`\[url=(.+?)\](.+?)\[/url\]`)
		for _, it := range re.FindAllStringSubmatch(cont, -1) {
			cont = strings.ReplaceAll(cont, `[url=`+it[1]+`]`+it[2]+`[/url]`, `[`+it[2]+`](`+it[1]+`)`)
		}

		//引用
		//下列的[b] 和[/b] 在这个接口下好像都变成了 <b> 和 </b>
		//圈主贴
		//(?s) 意思单行模式
		reg_str := `(?s)\[quote\]\[tid=.+?Post by \[uid.*?\](.+)\[\/uid\].*?\((\d{4}.+?)\):</b>(.+?)\[/quote\]((?:\n){0,2})`
		if !strings.Contains(cont, "uid=") {
			//匿名回复，没有uid
			reg_str = `(?s)\[quote\]\[tid=.+?Post by (.+)<span .*?\((\d{4}.+?)\):</b>(.+?)\[/quote\]((?:\n){0,2})`
		}
		re = regexp.MustCompile(reg_str)
		//[1]人名 [2]时间 [3]圈的内容
		for _, it := range re.FindAllStringSubmatch(cont, -1) {
			quoteText := strings.ReplaceAll(it[3], "\n", "\n>")
			quoteAuthor := it[1]
			quoteTime := it[2]
			if len(quoteAuthor) > 7 && quoteAuthor[:7] == `#anony_` {
				quoteAuthor = anony(quoteAuthor)
			}
			cont = strings.ReplaceAll(cont, it[0], `>[jump](#pid0) `+quoteAuthor+`(`+quoteTime+`)`+` 说:`+quoteText+"\n\n")
			floor.AppendPid = append(floor.AppendPid, 0)
		}

		//圈其他楼
		//[quote][pid=684818389,36006627,6]Reply[/pid] <b>Post by [uid=42264644]Lian君并没有名字[/uid] (2023-04-18 15:46):</b><br/><br/>有发表暴论的机会？那我也来整几个[/quote]1.后宫和
		//"[quote][pid=684833266,36008480,1]Reply[/pid] <b>Post by [uid=64575408]虹色的棉花糖[/uid] (2023-04-18 16:54):</b>\n\n上海呢[/quote]"
		//[quote][pid=684810015,36006627,5]Reply[/pid] <b>Post by 庚雷尤甲项季<span class=\"gray\">(83楼)</span> (2023-04-18 15:08):</b><br/><br/>阿巴[/quote]
		quoteCount := strings.Count(cont, "[quote]")
		for i := 0; i < quoteCount; i++ {
			//最内层的quote下标
			quoteStartIndex := strings.LastIndex(cont, "[quote]")
			quoteEndIndex := quoteStartIndex + strings.Index(cont[quoteStartIndex:], "[/quote]")
			clip := cont[quoteStartIndex : quoteEndIndex+8]

			reg_str := `(?s)\[quote\]\[pid=(\d+?),.+?Post by \[uid.*?\](.+)\[\/uid\].*?\((\d{4}.+?)\):</b>(.+?)\[/quote\]((?:\n){0,2})`
			if !strings.Contains(clip, "uid=") {
				//匿名回复，没有uid
				reg_str = `(?s)\[quote\]\[pid=(\d+?),.+?Post by (.+)<span .*?\((\d{4}.+?)\):</b>(.+?)\[/quote\]((?:\n){0,2})`
			}
			re = regexp.MustCompile(reg_str)
			//[1]pid [2]原作者 [3]时间 [4]说的东西
			for _, it := range re.FindAllStringSubmatch(clip, -1) {
				cont = strings.ReplaceAll(cont, `[url=`+it[1]+`]`+it[2]+`[/url]`, `[`+it[2]+`](`+it[1]+`)`)
				quoteText := strings.ReplaceAll(it[4], "\n", "\n>")
				quoteAuthor := it[2]
				quotePid := it[1]
				quoteTime := it[3]
				if len(quoteAuthor) > 7 && quoteAuthor[:7] == `#anony_` {
					quoteAuthor = anony(quoteAuthor)
				}
				cont = strings.ReplaceAll(cont, it[0], `>[jump](#pid`+quotePid+`) `+quoteAuthor+`(`+quoteTime+`)`+` 说:`+quoteText+"\n\n")
				//这里会有原文的，就不append了
			}
		}

		//回复
		reg_str = `(?s)<b>Reply to \[tid=(\d+?).+? Post by \[uid.*?\](.+)\[\/uid\].+?\((.+?)\)</b>((?:\n){0,2})`
		if !strings.Contains(cont, "uid=") {
			//匿名回复，没有uid
			reg_str = `(?s)<b>Reply to \[tid=(\d+?).+? Post by (.+)<span .+?\((.+?)\)</b>((?:\n){0,2})`
		}
		re = regexp.MustCompile(reg_str)
		//评论主楼[1]pid [2]原作者 [3]时间
		for _, it := range re.FindAllStringSubmatch(cont, -1) {
			quoteAuthor := it[2]
			quoteTime := it[3]
			if len(quoteAuthor) > 7 && quoteAuthor[:7] == `#anony_` {
				quoteAuthor = anony(quoteAuthor)
			}
			cont = strings.ReplaceAll(cont, it[0], `>[jump](#pid0) `+quoteAuthor+`(`+quoteTime+"):\n\n")
		}
		reg_str = `(?s)<b>Reply to \[pid=(\d+?),.+? Post by \[uid.*?\](.+)\[\/uid\].+?\((.+?)\)</b>((?:\n){0,2})`
		if !strings.Contains(cont, "uid=") {
			//匿名回复，没有uid
			reg_str = `(?s)<b>Reply to \[pid=(\d+?),.+? Post by (.+)<span .+?\((.+?)\)</b>((?:\n){0,2})`
		}
		re = regexp.MustCompile(reg_str)
		//[1]pid [2]原作者 [3]时间
		for _, it := range re.FindAllStringSubmatch(cont, -1) {
			quotePid := it[1]
			quoteAuthor := it[2]
			quoteTime := it[3]
			if len(quoteAuthor) > 7 && quoteAuthor[:7] == `#anony_` {
				quoteAuthor = anony(quoteAuthor)
			}
			cont = strings.ReplaceAll(cont, it[0], `>[jump](#pid`+quotePid+`) `+quoteAuthor+`(`+quoteTime+"):\n\n")
			floor.AppendPid = append(floor.AppendPid, cast.ToInt(quotePid))
		}
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

func (it *Tiezi) fixFloorContent(startFloor_i int) {

	var wg sync.WaitGroup
	p, _ := ants.NewPoolWithFunc(THREAD_COUNT, func(floor_i interface{}) {
		if it.Floors[cast.ToInt(floor_i)].Lou != -1 {
			responseChannel <- fmt.Sprintf("Start fix floor %d", cast.ToInt(floor_i))
			it.fixContent(cast.ToInt(floor_i))
		}
		wg.Done()
	})
	defer p.Release()
	//go responseController()

	startTime := time.Now()
	for i := startFloor_i; i < len(it.Floors); i++ {
		wg.Add(1)
		_ = p.Invoke(i)
		it.Timestamp = ts()
	}
	wg.Wait()
	elapsedTime := time.Since(startTime) / time.Millisecond
	fmt.Printf("fix floor总耗时 in %dms\n", elapsedTime)

}

func (tiezi *Tiezi) genMarkdown(localMaxFloor int) {
	fileName := `./` + cast.ToString(tiezi.Tid) + `/post.md`
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		_, _ = os.Create(fileName)
	}
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0666)
	defer f.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	for i := localMaxFloor; i < len(tiezi.Floors); i++ {
		floor := &tiezi.Floors[i]
		if floor.Lou == -1 {
			//被抽楼了
			continue
		}
		if floor.Pid == 0 {
			_, _ = f.WriteString(fmt.Sprintf("### %s\n\nMade by ngapost2md (c) ludoux [GitHub Repo](https://github.com/ludoux/ngapost2md)\n\n", tiezi.Title))
		}
		IpLocation_str := ""
		if floor.IpLocation != "" {
			IpLocation_str = "\\(" + floor.IpLocation + "\\)"
		}
		_, _ = f.WriteString(fmt.Sprintf("----\n##### <span id=\"pid%d\">%d.[%d] \\<pid:%d\\> %s by %s%s</span>\n%s", floor.Pid, floor.Lou, floor.LikeNum, floor.Pid, ts2t(floor.Timestamp), floor.Username, IpLocation_str, floor.Content))
		if floor.Comments != nil {
			_, _ = f.WriteString("\n*---下挂评论---*")
			for _, itt := range floor.Comments {
				_, _ = f.WriteString(fmt.Sprintf("\n\n%d.[%d] \\<pid:%d\\>%s by %s:\n%s", itt.Lou, itt.LikeNum, itt.Pid, ts2t(itt.Timestamp), itt.Username, itt.Content))
			}
		}
		// for _, v := range it.AppendPid {
		// 	_, _ = f.WriteString("\n*---AppendPid: " + cast.ToString(v) + "---*\n")
		// }
		_, ex := f.WriteString("\n\n")
		if ex != nil {
			fmt.Println(ex.Error())
		}
	}
}

func responseController() {
	for rc := range responseChannel {
		fmt.Println("response: ", rc)
	}
}

// 默认清空内容
func (it *Tiezi) SaveAsFile() {
	//为节省大小和导入导出压力，清空具体回复内容
	copy := it
	for i := range copy.Floors {
		copy.Floors[i].Content = ""
		for ii := range copy.Floors[i].Comments {
			copy.Floors[i].Comments[ii].Content = ""
		}
	}
	result, err := json.Marshal(copy)
	if err != nil {
		fmt.Println(err)
	}
	fileName := `./` + cast.ToString(copy.Tid) + `/tiezi.json`
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		_, _ = os.Create(fileName)
	}
	f, _ := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0666)
	_, _ = f.Write(result)
	defer f.Close()
}

func (it *Tiezi) SaveProcessInfo() {
	fileName := `./` + cast.ToString(it.Tid) + `/process.ini`
	cfg := ini.Empty()
	cfg.NewSection("local")
	cfg.Section("local").NewKey("max_floor", cast.ToString(it.LocalMaxFloor))
	cfg.Section("local").NewKey("max_page", cast.ToString(it.LocalMaxPage))
	cfg.SaveTo(fileName)
}

func (it *Tiezi) SaveAssetsMap() {
	fileName := `./` + cast.ToString(it.Tid) + `/assets.json`
	result, err := json.Marshal(it.Assets)
	if err != nil {
		fmt.Println(err)
	}
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		_, _ = os.Create(fileName)
	}
	f, _ := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0666)
	_, _ = f.Write(result)
	defer f.Close()
}

func (it *Tiezi) Download() {
	if it.Tid != 0 {
		var wg sync.WaitGroup
		p, _ := ants.NewPoolWithFunc(THREAD_COUNT, func(page interface{}) {
			time.Sleep(time.Millisecond * time.Duration(DELAY_MS))
			responseChannel <- fmt.Sprintf("Start page %d", page)
			//1. 并行下载page
			it.page(cast.ToInt(page))
			wg.Done()
		})
		defer p.Release()
		go responseController()

		startTime := time.Now()
		//因为 it.LocalMaxPage 在InitFromxxx的时候已经page过了
		for page := it.LocalMaxPage + 1; page <= it.WebMaxPage; page++ {
			wg.Add(1)
			_ = p.Invoke(page)
		}
		wg.Wait()

		elapsedTime := time.Since(startTime) / time.Millisecond
		fmt.Printf("下载page总耗时 in %dms\n", elapsedTime)

		//2. 格式化content
		it.fixFloorContent(it.LocalMaxFloor + 1)

		//3. 制作文件
		it.genMarkdown(it.LocalMaxFloor + 1)

		it.LocalMaxPage = it.WebMaxPage

		//因为NGA会抽楼，floorcount不准，只能这样子
		for i := len(it.Floors) - 1; ; i-- {
			floor := &it.Floors[i]
			if floor.Lou > -1 {
				it.LocalMaxFloor = floor.Lou
				break
			}
		}
		// 存储tiezi---暂时注释掉，还是使用存储localmaxpage和maxfloor(SaveProcessInfo)的方法。
		//it.SaveAsFile()

		//存储localmaxpage和maxfloor
		it.SaveProcessInfo()

		//存储assets map
		it.SaveAssetsMap()
	}
}
