package main

import (
	"encoding/json"
	"fmt"
	"github.com/imroc/req/v3"
	"github.com/jessevdk/go-flags"
	"github.com/ludoux/ngapost2md/nga"
	"github.com/spf13/cast"
	"log"
	"math"
	"os"
	"strings"
	"time"
)

type Option struct {
	Version bool `short:"v" long:"version" description:"显示版本信息并退出"`
	Help    bool `short:"h" long:"help" description:"显示此帮助信息并退出"`
	Update  bool `short:"u" long:"update" description:"检查最新版本"`
}

// 检查更新，解析json数据
type Repo struct {
	Tag_name string `json:"tag_name"` // 最新版本号
	Body     string `json:"body"`     // 更新信息为markdown格式
}

func checkUpdate() {
	resp, _ := req.C().R().Get("https://api.github.com/repos/ludoux/ngapost2md/releases/latest")

	// 读取最新版本号
	var repo Repo
	err := json.Unmarshal([]byte(resp.String()), &repo)
	if err != nil {
		fmt.Println("解析json数据失败:", err)
	}

	// 输出信息
	log.Printf("目前版本: %s 最新版本: %s", nga.VERSION, repo.Tag_name)
	log.Fatalln("请去 GitHub Releases 页面下载最新版本。软件即将退出……")

}

func main() {
	var opts Option
	parser := flags.NewParser(&opts, flags.Default & ^flags.HelpFlag)
	//args为剩余未解析的（比如tid）
	args, err := parser.Parse()
	if err != nil {
		log.Fatalln("参数解析出现问题:", err.Error())
	}

	if opts.Version {
		fmt.Println("ngapost2md", nga.VERSION)
		fmt.Println("Build_Time:", nga.BUILD_TS, time.Unix(cast.ToInt64(nga.BUILD_TS), 0).Local().Format("2006-01-02T15:04:05Z07:00"))
		fmt.Println("Git_Ref:", nga.GIT_REF)
		fmt.Println("Git_Hash:", nga.GIT_HASH)
		os.Exit(0)
	} else if opts.Help {
		fmt.Println("使用: ngapost2md tid")
		fmt.Println("选项与参数说明: ")
		fmt.Println("tid: 待下载的帖子 tid 号")
		fmt.Println("")
		fmt.Println("ngapost2md -v, --version    ", parser.FindOptionByLongName("version").Description)
		fmt.Println("ngapost2md -h, --help       ", parser.FindOptionByLongName("help").Description)
		fmt.Println("ngapost2md -u, --update     ", parser.FindOptionByLongName("update").Description)
		os.Exit(0)
	} else if opts.Update {
		checkUpdate()
	}

	var tid int
	if len(args) != 1 {
		log.Fatalln("未传入 tid 或格式错误")
	} else {
		tid, err = cast.ToIntE(args[0])
		if err != nil {
			log.Fatalln("tid", args[0], "无法转为数字:", err.Error())
		}
	}

	//args check all passed

	fmt.Printf("ngapost2md (c) ludoux [ GitHub: https://github.com/ludoux/ngapost2md/tree/neo ]\nVersion: %s     %s\n", nga.VERSION, time.Unix(cast.ToInt64(nga.BUILD_TS), 0).Local().Format("2006-01-02T15:04:05Z07:00"))
	if nga.DEBUG_MODE == "1" {
		fmt.Println("==debug mode===")
		fmt.Println("***DEBUG MODE ON***")
		fmt.Printf("opts: %+v ; args: %v\n", opts, args)
		fmt.Println("==debug mode===")
	}

	// 检查并按需更新配置文件
	cfg := nga.Read_cfg()

	//Cookie
	var ngaPassportUid = cfg.Network.Uid
	var ngaPassportCid = cfg.Network.Cid
	var cookie strings.Builder
	cookie.WriteString("ngaPassportUid=" + ngaPassportUid + ";" + "ngaPassportCid=" + ngaPassportCid)
	nga.COOKIE = cookie.String()

	nga.BASE_URL = cfg.Network.Base_url

	// 如果配置中的Ua留空就使用默认Ua
	if cfg.Network.Ua == "" {
		nga.UA = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36"
	} else {
		nga.UA = cfg.Network.Ua
	}
	//默认线程数为2,仅支持1~3
	nga.THREAD_COUNT = int(math.Max(1, math.Min(float64(cfg.Network.Thread), 3)))
	nga.PAGE_DOWNLOAD_LIMIT = int(math.Max(-1, math.Min(float64(cfg.Network.Page_limit), 100)))
	nga.GET_IP_LOCATION = cfg.Post.Ip
	nga.ENHANCE_ORI_REPLY = cfg.Post.Reply
	nga.USE_LOCAL_SMILE_PIC = cfg.Post.Local_smile
	nga.LOCAL_SMILE_PIC_PATH = cfg.Post.Local_smile_path
	nga.USE_TITLE_AS_FOLDER_NAME = cfg.Post.Title_dir_name
	nga.USE_TITLE_AS_MD_FILE_NAME = cfg.Post.Title_md_name
	nga.Client = nga.NewNgaClient()

	tie := nga.Tiezi{}

	path := nga.FindFolderNameByTid(tid)
	if path != "" {
		log.Printf("本地存在此 tid (%s) 文件夹，追加最新更改。", path)
		tie.InitFromLocal(tid)

	} else {
		tie.InitFromWeb(tid)
	}

	tie.Download()
}
