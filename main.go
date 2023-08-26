package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/imroc/req/v3"
	"github.com/jessevdk/go-flags"
	"github.com/ludoux/ngapost2md/config"
	"github.com/ludoux/ngapost2md/nga"
	"github.com/spf13/cast"
)

type Option struct {
	ForceNoCheckUpdate bool `long:"force-no-check-update" description:"在编译时间后的不少于 60 天内，不检测版本更新"`
	DumpUpdateInfo     bool `long:"dump-update-info" description:"当检测到新版本时，将版本信息原始文件写入 NEED_UPDATE 文件"`
	Version            bool `short:"v" long:"version" description:"显示版本信息并退出"`
	Help               bool `short:"h" long:"help" description:"显示此帮助信息并退出"`
	GenConfigFile      bool `long:"gen-config-file" description:"生成默认配置文件于 config.ini 并退出"`
}

func checkUpdate(dump bool) {
	resp, _ := req.C().R().Get("https://gitee.com/ludoux/check-update/raw/master/ngapost2md/version_neo.txt")
	//版本更新配置文件改为 DO_NOT_CHECK ，软件则不会强制使用最新版本
	if resp.String() != nga.VERSION && resp.String() != "DO_NOT_CHECK" {
		if dump {
			f, _ := os.OpenFile("NEED_UPDATE", os.O_CREATE|os.O_WRONLY, 0666)
			_, _ = f.Write(resp.Bytes())
			defer f.Close()
		}
		log.Printf("目前版本: %s 最新版本: %s", nga.VERSION, resp.String())
		log.Fatalln("请去 GitHub Releases 页面下载最新版本。软件即将退出……")
	}
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
	} else if opts.GenConfigFile {
		err := config.SaveDefaultConfigFile()
		if err != nil {
			log.Fatalln(err.Error())
		}
		log.Println("导出默认配置文件 config.ini 成功。")
		os.Exit(0)
	} else if opts.Help {
		fmt.Println("使用: ngapost2md tid [--force-no-check-update] [--dump-update-info]")
		fmt.Println("选项与参数说明: ")
		fmt.Println("tid: 待下载的帖子 tid 号")
		fmt.Println("--force-no-check-update     ", parser.FindOptionByLongName("force-no-check-update").Description)
		fmt.Println("--dump-update-info          ", parser.FindOptionByLongName("dump-update-info").Description)
		fmt.Println("")
		fmt.Println("ngapost2md -v, --version    ", parser.FindOptionByLongName("version").Description)
		fmt.Println("ngapost2md -h, --help       ", parser.FindOptionByLongName("help").Description)
		fmt.Println("ngapost2md --gen-config-file", parser.FindOptionByLongName("gen-config-file").Description)
		os.Exit(0)
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
	fmt.Println("此版本永久禁用检查更新。请手动前往 GitHub 项目主页查看，谢谢。")
	/*
		//DEBUG_MODE不为1时(不是DEBUG_MODE)情况下，未有 ForceNoCheckUpdate 时检测更新
		if nga.DEBUG_MODE == "1" {
			fmt.Println("==debug mode===")
			fmt.Println("DEBUG MODE 下恒不检查更新")
			fmt.Println("==debug mode===")
		} else {
			if !opts.ForceNoCheckUpdate {
				checkUpdate(opts.DumpUpdateInfo)
			} else {
				//如果有这个标，在大于指定天数后也要检查更新
				curTs := time.Now().Local().Unix()
				builtTs, err := cast.ToInt64E(nga.BUILD_TS)
				if err != nil {
					log.Fatalln("BUILD_TS", nga.BUILD_TS, "无法转为数字，可能编译时 ldflags 有误")
				}

				if curTs < builtTs {
					fmt.Println("由于本地时间有误，仍将检查更新。")
					checkUpdate(opts.DumpUpdateInfo)
				} else if curTs-builtTs > 61*86400 {
					//61天
					fmt.Println("距离此版本编译时间已过去", (curTs-builtTs)/86400, "天，仍将检查更新。")
					checkUpdate(opts.DumpUpdateInfo)
				} else {
					fmt.Println("由于使用了 --force-no-check-update 标志，且距离编译时间仍在时限内，本次不检查更新。距离此版本编译时间已过", (curTs-builtTs)/86400, "天")
				}
			}
		}
	*/
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

	path := nga.FindFolderNameByTid(tid)
	if path != "" {
		log.Printf("本地存在此 tid (%s) 文件夹，追加最新更改。", path)
		tie.InitFromLocal(tid)

	} else {
		tie.InitFromWeb(tid)
	}

	tie.Download()
}
