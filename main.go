package main

import (
	"fmt"
	"os"

	"github.com/ludoux/ngapost2md/nga"
	"github.com/spf13/cast"
	"gopkg.in/ini.v1"
)

func main() {
	cfg, err := ini.Load("config.ini")
	if err != nil {
		fmt.Printf("Fail to find or read config.ini file: %v", err)
		os.Exit(1)
	}
	nga.BASE_URL = cfg.Section("network").Key("baseurl").String()
	nga.UA = cfg.Section("network").Key("ua").String()
	nga.THREAD_COUNT, _ = cfg.Section("network").Key("thread").Int()
	nga.Client = nga.NewNgaClient()

	tie := nga.Tiezi{}

	if len(os.Args) != 2 {
		fmt.Println("传参数目错误")
		os.Exit(1)
	}
	if _, err := os.Stat(cast.ToString(os.Args[1])); os.IsNotExist(err) {
		tie.InitFromWeb(cast.ToInt(os.Args[1]))
	} else {
		fmt.Println("存在 tid 文件夹")
		tie.InitFromLocal(cast.ToInt(os.Args[1]))
	}

	//tie.InitFromWeb(36229275)
	//tie.InitFromLocal(36229275)

	tie.Download()
}
