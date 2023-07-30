package config

import (
        "fmt"
        "gopkg.in/ini.v1"
)

func UpdateConfig() {
        // 定义默认配置
        var defaultConfig = map[string]map[string]string{
                "config": {
                        "version": "1.4.0",
                },
                "network": {
                        "base_url":            "https://bbs.nga.cn",
                        "ua":                  `"""MODIFY_ME"""`,
                        "ngaPassportUid":      "MODIFY_ME",
                        "ngaPassportCid":      "MODIFY_ME",
                        "thread":              "2",
                        "page_download_limit": "100",
                },
                "post": {
                        "enable_post_title": "False",
                        "get_ip_location":   "False",
                        "enhance_ori_reply": "False",
                },
        }

        // 打开旧的INI配置文件
        cfg, err := ini.Load("config.ini")
        if err != nil {
                fmt.Printf("无法加载配置文件: %v", err)
                return
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
                        defaultValue := defaultConfig[sectionName][keyName]

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
                                _, err := section.NewKey(key, defaultValue)
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
