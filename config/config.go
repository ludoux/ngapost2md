package config

import (
	"fmt"

	"gopkg.in/ini.v1"
)

// 定义默认配置。使用 slice 保证顺序
var sectionList = []string{"config", "network", "post"}

var defaultConfig = map[string][][3]string{
	"config": {
		{"version", "1.4.0", "请不要修改此处。请不要使用配置文件注释项来记录信息，软件加载时会强制覆盖删除。"},
	},
	"network": {
		{"base_url", "https://bbs.nga.cn", "软件访问的nga域名。默认值 https://bbs.nga.cn"},
		{"ua", "<;MODIFY_ME;>", "浏览器User-Agent，通常来说填写你常用浏览器的UA即可。修改时请将两个分号和内部英文、空格等均替换删除修改。"},
		{"ngaPassportUid", "<;MODIFY_ME;>", "nga网站个人cookie项目。修改时请将 两个尖括号及内部所有文本 均替换删除修改。"},
		{"ngaPassportCid", "<;MODIFY_ME;>", "nga网站个人cookie项目。修改时请将 两个尖括号及内部所有文本 均替换删除修改。"},
		{"thread", "2", "网络并发数，提高理论上可以增加下载速度。仅支持1、2、3。若开启enhance_ori_reply，请将此值设定为1。默认值2。"},
		{"page_download_limit", "100", "[#56]每次下载限制新下载的大约页数。到上限后需要重新运行程序再追加下载，如此直至全部下载成功。允许范围-1（含）至100（含）。值为0或-1时则不限制。默认值100（约100页）。"},
	},
	"post": {
		{"get_ip_location", "False", "[#44]是否查询用户基于IP的地理位置？若启用则会导致至高20倍于未启用的网络请求。默认值False（不启用）。"},
		{"enhance_ori_reply", "False", "[#35]将被回复的楼层内容补充完整。开启此功能要求同步将thread线程数设置为1，否则可能会补充到未format的文本。默认值False（不启用）。"},
		{"use_local_smile_pic", "False", "[#58]是否使用本地表情图片资源而不是引用在线资源。默认值False（不启用）。"},
		{"local_smile_pic_path", "../smile/", "[#58]本地表情图片资源路径。支持绝对路径与相对路径。尾部需要以 / 结尾。"},
		{"use_title_as_folder_name", "False", "[#21]文件夹名是否包含标题。默认值False。修改后仅对全新拉取的tid生效。"},
		{"use_title_as_md_file_name", "False", "[#21]Markdown 文件名是否为标题。默认值False。修改后仅对全新拉取的tid生效。"},
	},
}

// 会自动更新、格式化配置文件并保存
func GetConfig() (*ini.File, error) {
	//不要等号对齐，那样子好难看
	ini.PrettyFormat = false
	// 打开旧的INI配置文件
	cfg, err := ini.Load("config.ini")
	if err != nil {
		return nil, fmt.Errorf("无法加载配置文件: %v", err)
	}
	oldCfgVersion := cfg.Section("config").Key("version").String()
	//此为默认配置
	defaultcfg := genDefaultConfig()

	// 针对相同功能，新配置相比于旧配置名字不同，需要进行自动迁移
	switch oldCfgVersion {
	case "1.2.0":
		//从1.2.0->1.4.0时，enable_post_title 更名为 use_title_as_md_file_name
		var oldValue string
		if cfg.Section("post").Key("enable_post_title").MustBool() {
			oldValue = "True"
		} else {
			oldValue = "False"
		}
		defaultcfg.Section("post").Key("use_title_as_md_file_name").SetValue(oldValue)
	}

	//基于默认配置，往默认配置内填已存在配置的信息
	for _, section := range defaultcfg.Sections() {
		for _, key := range section.Keys() {
			if cfg.HasSection(section.Name()) && cfg.Section(section.Name()).HasKey(key.Name()) {
				//读取配置内有此项，将value值填入默认配置内
				cfgValue := cfg.Section(section.Name()).Key(key.Name()).Value()
				key.SetValue(cfgValue)
			}
		}
	}

	err = defaultcfg.SaveTo("config.ini")
	if err != nil {
		return nil, fmt.Errorf("无法保存更新后的配置文件: %v", err)
	}
	return cfg, nil
}

func genDefaultConfig() *ini.File {
	// 不要等号对齐，那样子好难看
	ini.PrettyFormat = false
	cfg := ini.Empty()
	for _, section := range sectionList {
		sectionNode, _ := cfg.NewSection(section)
		sectionDetail := defaultConfig[section]
		for _, v := range sectionDetail {
			kNode, _ := sectionNode.NewKey(v[0], v[1])
			kNode.Comment = v[2]
		}
	}
	return cfg
}

func GenDefaultConfigFile() error {
	return genDefaultConfig().SaveTo("config.ini")
}
