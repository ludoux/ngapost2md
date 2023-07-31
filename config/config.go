package config

import (
	"fmt"

	"gopkg.in/ini.v1"
)

func UpdateConfig() {
	// 定义默认配置
	var defaultConfig = map[string]map[string][2]string{
		"config": {
			"version": {"1.4.0", "请不要修改此处"},
		},
		"network": {
			"base_url":            {"https://bbs.nga.cn", ""},
			"ua":                  {"<;MODIFY_ME;>", "浏览器User-Agent，通常来说填写你常用浏览器的UA即可。修改时请将两个分号和内部英文、空格等均替换删除修改。"},
			"ngaPassportUid":      {"<;MODIFY_ME;>", "nga网站个人cookie项目。修改时请将两个尖括号及内部所有等均替换删除修改。"},
			"ngaPassportCid":      {"<;MODIFY_ME;>", "nga网站个人cookie项目。修改时请将两个尖括号及内部所有等均替换删除修改。"},
			"thread":              {"2", "线程数，提高理论上可以增加下载速度。仅支持1、2、3。若开启enhance_ori_reply，请将此值设定为1。默认值2。"},
			"page_download_limit": {"100", "每次下载限制新下载的大约页数。到上限后需要重新运行程序再追加下载，如此直至全部下载成功。允许范围-1（含）至100（含）。值为0或-1时则不限制。默认值100（约100页）。"},
		},
		"post": {
			"get_ip_location":           {"False", "是否查询用户基于IP的地理位置？若启用则会导致至高20倍于未启用的网络请求。默认值False（不启用）。"},
			"enhance_ori_reply":         {"False", "将被回复的楼层内容补充完整。见issue#35 。开启此功能要求同步将thread线程数设置为1，否则可能会补充到未format的文本。默认值False（不启用）。"},
			"use_local_smile_pic":       {"False", "是否使用本地表情图片资源而不是引用在线资源。默认值False（不启用）。"},
			"local_smile_pic_path":      {"../smile/", "本地表情图片资源路径。支持绝对路径与相对路径。尾部需要以 / 结尾。"},
			"use_title_as_folder_name":  {"False", "文件夹名是否包含标题。默认值False。修改后仅对全新拉取的tid生效。"},
			"use_title_as_md_file_name": {"False", "Markdown 文件名是否为标题。默认值False。修改后仅对全新拉取的tid生效。"},
		},
	}

	// 打开旧的INI配置文件
	cfg, err := ini.Load("config.ini")
	if err != nil {
		fmt.Printf("无法加载配置文件: %v", err)
		return
	}
	oldCfgVersion := cfg.Section("config").Key("version").String()

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
		cfg.Section("post").NewKey("use_title_as_md_file_name", oldValue)
	}

	// 遍历旧配置的节和配置项，并与默认配置进行对比
	for _, section := range cfg.Sections() {
		sectionName := section.Name()

		// 如果默认配置中不存在该节，则删除该节
		if _, ok := defaultConfig[sectionName]; !ok {
			cfg.DeleteSection(sectionName)
			continue
		}

		// 遍历旧配置的配置项
		for _, key := range section.Keys() {
			keyName := key.Name()
			defaultValue := defaultConfig[sectionName][keyName][0]

			// 如果默认配置中不存在该配置项，则删除该配置项
			if defaultValue == "" {
				section.DeleteKey(keyName)
				continue
			}

		}
	}

	// 遍历默认配置的节和配置项，如果旧配置中不存在该节或配置项，则添加
	for sectionName, sectionConfig := range defaultConfig {
		section, err := cfg.GetSection(sectionName)
		if err != nil {
			section, err = cfg.NewSection(sectionName)
			if err != nil {
				fmt.Printf("无法创建新的节: %v", err)
				return
			}
		}

		for key, defaultValue := range sectionConfig {
			if !section.HasKey(key) {
				k, err := section.NewKey(key, defaultValue[0])
				k.Comment = defaultValue[1]
				if err != nil {
					fmt.Printf("无法创建新的配置项: %v", err)
					return
				}
			}
		}
	}

	err = cfg.SaveTo("config.ini")
	if err != nil {
		fmt.Printf("无法保存更新后的配置文件: %v", err)
		return
	}

}
