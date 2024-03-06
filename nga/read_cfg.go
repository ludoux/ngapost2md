package nga

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Version *version
	Network *network
	Post    *post
}
type version struct {
	Version string `comment:"软件版本信息。"`
}
type network struct {
	Base_url   string `comment:"软件访问的nga域名。默认值 https://bbs.nga.cn"`
	Ua         string `comment:"浏览器User-Agent，通常来说填写你常用浏览器的UA即可，留空则使用默认值"`
	Uid        string `comment:"nga网站个人cookie项目"`
	Cid        string `comment:"nga网站个人cookie项目"`
	Thread     int    `comment:"网络并发数，提高理论上可以增加下载速度。仅支持1、2、3。若开启enhance_ori_reply，请将此值设定为1。默认值2"`
	Page_limit int    `comment:"[#56]每次下载限制新下载的大约页数。到上限后需要重新运行程序再追加下载，如此直至全部下载成功。允许范围-1（含）至100（含）。值为0或-1时则不限制。默认值100（约100页）"`
}
type post struct {
	Ip               bool   `comment:"[#44]是否查询用户基于IP的地理位置？若启用则会导致至高20倍于未启用的网络请求。默认值False（不启用）"`
	Reply            bool   `comment:"[#35]将被回复的楼层内容补充完整。开启此功能要求同步将thread线程数设置为1，否则可能会补充到未format的文本。默认值False（不启用）"`
	Local_smile      bool   `comment:"[#58]是否使用本地表情图片资源而不是引用在线资源。默认值False（不启用）"`
	Local_smile_path string `comment:"[#58]本地表情图片资源路径。支持绝对路径与相对路径。尾部需要以 / 结尾"`
	Title_dir_name   bool   `comment:"[#21]文件夹名是否包含标题。默认值False。修改后仅对全新拉取的tid生效"`
	Title_md_name    bool   `comment:"[#21]Markdown 文件名是否为标题。默认值False。修改后仅对全新拉取的tid生效"`
}

func Read_cfg() Config {
	f := "config.toml"
	if _, err := os.Stat(f); err != nil {
		panic(err)
	}
	tomlBytes, err := os.ReadFile("config.toml")
	if err != nil {
		panic(err)
	}

	var cfg Config
	err = toml.Unmarshal(tomlBytes, &cfg)
	if err != nil {
		panic(err)
	}
	return cfg
}

// output default config
func Gen_cfg() {
	cfg := &Config{
		Version: &version{
			Version: VERSION,
		},
		Network: &network{
			Base_url:   "https://bbs.nga.cn",
			Thread:     2,
			Page_limit: 100,
		},
		Post: &post{
			Local_smile_path: "../smile/",
		},
	}

	out, err := toml.Marshal(cfg)
	if err != nil {
		panic(err)
	}
	err = toml.Unmarshal(out, &cfg)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("config.toml", out, 0666)
	if err != nil {
		panic(err)
	}
	fmt.Println("导出默认配置文件 config.toml 成功")
}
