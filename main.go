package main

import (
	"fmt"
	"os"

	"github.com/ludoux/ngapost2md/nga"
	"github.com/spf13/cast"
	"gopkg.in/ini.v1"
)

func main() {
	const Url = "https://bbs.nga.cn"
	const Ua  = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36"

	fmt.Println("ngapost2md (c) ludoux [ GitHub: https://github.com/ludoux/ngapost2md/tree/neo ]")
	fmt.Println("Version: " + nga.VERSION)
	fmt.Println("")
	cfg, err := ini.Load("assets/config.ini")
	if err != nil {
		fmt.Printf("Fail to find or read config.ini file: %v", err)
		os.Exit(1)
	}

	//默认设置
	nga.BASE_URL = Url
	nga.UA = Ua

	//从assets文件夹下的config.ini文件，修改ngaPassportUid，ngaPassportCid为自己的值
	var ngaPassportUid = cfg.Section("network").Key("ngaPassportUid").String()
	var ngaPassportCid = cfg.Section("network").Key("ngaPassportCid").String()
	nga.COOKIE = fmt.Sprintf("ngaPassportUid="+ngaPassportUid+";"+"ngaPassportCid="+ngaPassportCid)

	//默认线程数为2,仅支持1~3
	nga.THREAD_COUNT = cfg.Section("network").Key("thread").InInt(2, []int{1, 2, 3})
	nga.GET_IP_LOCATION = cfg.Section("post").Key("get_ip_location").MustBool()
	nga.ENHANCE_ORI_REPLY = cfg.Section("post").Key("enhance_ori_reply").MustBool()
	nga.Client = nga.NewNgaClient()

	tie := nga.Tiezi{}

	if len(os.Args) != 2 && len(os.Args) != 3 {
		fmt.Println("传参数目错误")
		os.Exit(1)
	}

	force_max_page := -1
	if len(os.Args) == 3 {
		force_max_page = cast.ToInt(os.Args[2])
	}

	if _, err := os.Stat(cast.ToString(os.Args[1])); os.IsNotExist(err) {
		tie.InitFromWeb(cast.ToInt(os.Args[1]), force_max_page)
	} else {
		fmt.Println("存在 tid 文件夹")
		tie.InitFromLocal(cast.ToInt(os.Args[1]), force_max_page)
	}

	tie.Download()
}
