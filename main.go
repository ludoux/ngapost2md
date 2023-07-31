package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/imroc/req/v3"
	"github.com/ludoux/ngapost2md/nga"
	"github.com/spf13/cast"
	"gopkg.in/ini.v1"
)

var (
	//修改了 config.ini 文件后，需要同步修改此处和 assets/config.ini 文件里的 version
	CONFIG_VERSION = "1.2.0"
)

func main() {
	fmt.Printf("ngapost2md (c) ludoux [ GitHub: https://github.com/ludoux/ngapost2md/tree/neo ]\nVersion: %s\n", nga.VERSION)
	if nga.DEBUG_MODE == "1" {
		fmt.Println("***DEBUG MODE ON***")
	}
	if len(os.Args) != 2 {
		log.Fatalln("传参数目错误！请使用 ngapost2md -h 命令查看 ngapost2md 的使用参数说明。")
	}
	if len(os.Args) == 2 && (cast.ToString(os.Args[1]) == "help" || cast.ToString(os.Args[1]) == "-help" || cast.ToString(os.Args[1]) == "--help" || cast.ToString(os.Args[1]) == "-h") {
		fmt.Println("使用: ngapost2md tid")
		fmt.Println("选项与参数说明: ")
		fmt.Println("tid: 待下载的帖子 tid 号")
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
	if cfg.Section("config").Key("version").String() != CONFIG_VERSION {
		log.Fatalf("config.ini 版本号(%s)与本软件所需版本(%s)不符！\n软件升级后，请使用最新的 config.ini 配置文件，并修改其内的 UA、uid、cid 和其它个性化设置\n", cfg.Section("config").Key("version").String(), CONFIG_VERSION)
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
	nga.PAGE_DOWNLOAD_LIMIT = cfg.Section("network").Key("page_download_limit").RangeInt(100, -1, 100)
	nga.GET_IP_LOCATION = cfg.Section("post").Key("get_ip_location").MustBool()
	nga.ENHANCE_ORI_REPLY = cfg.Section("post").Key("enhance_ori_reply").MustBool()
	nga.USE_LOCAL_SMILE_PIC = cfg.Section("post").Key("use_local_smile_pic").MustBool()
	nga.LOCAL_SMILE_PIC_PATH = cfg.Section("post").Key("local_smile_pic_path").String()
	nga.Client = nga.NewNgaClient()

	tie := nga.Tiezi{}

	tid, err := cast.ToIntE(os.Args[1])
	if err != nil {
		log.Fatalln("tid 无法转为数字:", err.Error())
	}

	// 检查是否存在包含tid的文件夹
	dirname := fmt.Sprintf(("%v%v%v"), "./", tid, "*")

	matches, err := filepath.Glob(dirname)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if len(matches) > 0 {
		for _, match := range matches {
			nga.FILENAME = match
			log.Println("本地存在此 tid 文件夹，追加最新更改。")
			tie.InitFromLocal(tid)
		}
	} else {
		tie.InitFromWeb(tid)
	}

	tie.Download()
	// 第一次下载完成后修改文件夹名，追加修改时跳过
	tie.ChangeDirName(tid)
}
