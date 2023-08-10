package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/imroc/req/v3"
	"github.com/ludoux/ngapost2md/config"
	"github.com/ludoux/ngapost2md/nga"
	"github.com/spf13/cast"
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
		fmt.Println("")
		fmt.Println("ngapost2md --gen-config-file : 可生成默认配置文件于 config.ini")
		os.Exit(0)
	}
	if len(os.Args) == 2 && (cast.ToString(os.Args[1]) == "--gen-config-file") {
		err := config.SaveDefaultConfigFile()
		if err != nil {
			log.Fatalln(err.Error())
		}
		log.Println("导出默认配置文件 config.ini 成功。")
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

	// 检查并按需更新配置文件
	cfg, err := config.GetConfigAutoUpdate()
	if err != nil {
		log.Fatalln(err.Error())
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
	nga.USE_TITLE_AS_FOLDER_NAME = cfg.Section("post").Key("use_title_as_folder_name").MustBool()
	nga.USE_TITLE_AS_MD_FILE_NAME = cfg.Section("post").Key("use_title_as_md_file_name").MustBool()
	nga.Client = nga.NewNgaClient()

	tie := nga.Tiezi{}

	tid, err := cast.ToIntE(os.Args[1])
	if err != nil {
		log.Fatalln("tid 无法转为数字:", err.Error())
	}
	path := nga.FindFolderNameByTid(tid)
	if path != "" {
		log.Printf("本地存在此 tid (%s) 文件夹，追加最新更改。", path)
		tie.InitFromLocal(tid)

	} else {
		tie.InitFromWeb(tid)
	}

	tie.Download()
}
