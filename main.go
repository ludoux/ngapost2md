package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/imroc/req/v3"
	"github.com/ludoux/ngapost2md/nga"
	"github.com/spf13/cast"
	"gopkg.in/ini.v1"
)

func main() {
	fmt.Printf("ngapost2md (c) ludoux [ GitHub: https://github.com/ludoux/ngapost2md/tree/neo ]\nVersion: %s\n", nga.VERSION)
	if nga.DEBUG_MODE == "1" {
		fmt.Println("***DEBUG MODE ON***")
	}
	if len(os.Args) != 2 && len(os.Args) != 3 {
		log.Fatalln("传参数目错误！请使用 ngapost2md -h 命令查看 ngapost2md 的使用参数说明。")
	}
	if len(os.Args) == 2 && (cast.ToString(os.Args[1]) == "help" || cast.ToString(os.Args[1]) == "-help" || cast.ToString(os.Args[1]) == "--help" || cast.ToString(os.Args[1]) == "-h") {
		fmt.Println("使用: ngapost2md tid [force_max_page]")
		fmt.Println("选项与参数说明: ")
		fmt.Println("tid: 待下载的帖子 tid 号")
		fmt.Println("force_max_page: 强制下载的最大页数，需要注意此页数需要小于帖子的实际页数。调试用。")
		os.Exit(0)
	}

	//DEBUG_MODE不为1时（不是DEBUG_MODE）时检测更新
	if nga.DEBUG_MODE != "1" {
		resp, _ := req.C().R().Get("https://gitee.com/ludoux/check-update/raw/master/ngapost2md/version_neo.txt")
		//版本更新配置文件改为 DO_NOT_CHECK ，软件则不会强制使用最新版本
		if resp.String() != nga.VERSION && resp.String() != "DO_NOT_CHECK" {
			log.Printf("目前版本: %s 最新版本: %s", nga.VERSION, resp.String())
			log.Fatalln("请去 GitHub Releases 页面下载最新版本。软件即将退出……")
		}
	}
	cfg, err := ini.Load("config.ini")
	if err != nil {
		log.Fatalln("无法读取 config.ini 文件，请检查文件是否存在。错误信息:", err.Error())
	}

	//Cookie
	var ngaPassportUid = cfg.Section("network").Key("ngaPassportUid").String()
	var ngaPassportCid = cfg.Section("network").Key("ngaPassportCid").String()
	var cookie strings.Builder
	cookie.WriteString("ngaPassportUid=" + ngaPassportUid + ";" + "ngaPassportCid=" + ngaPassportCid)
	nga.COOKIE = cookie.String()

	nga.BASE_URL = cfg.Section("network").Key("base_url").String()
	nga.UA = cfg.Section("network").Key("ua").String()
	//默认线程数为2,仅支持1~3
	nga.THREAD_COUNT = cfg.Section("network").Key("thread").InInt(2, []int{1, 2, 3})
	nga.GET_IP_LOCATION = cfg.Section("post").Key("get_ip_location").MustBool()
	nga.ENHANCE_ORI_REPLY = cfg.Section("post").Key("enhance_ori_reply").MustBool()
	nga.ENABLE_POST_TITLE = cfg.Section("post").Key("enable_post_title").MustBool()
	nga.Client = nga.NewNgaClient()

	tie := nga.Tiezi{}

	force_max_page := -1
	if len(os.Args) == 3 {
		force_max_page, err = cast.ToIntE(os.Args[2])
		if err != nil {
			log.Fatalln("force_max_page 无法转为数字:", err.Error())
		}
	}
	tid, err := cast.ToIntE(os.Args[1])
	if err != nil {
		log.Fatalln("tid 无法转为数字:", err.Error())
	}
	if _, err := os.Stat(cast.ToString(tid)); os.IsNotExist(err) {
		tie.InitFromWeb(tid, force_max_page)
	} else {
		log.Println("本地存在此 tid 文件夹，追加最新更改。")
		tie.InitFromLocal(tid, force_max_page)
	}

	tie.Download()
}
