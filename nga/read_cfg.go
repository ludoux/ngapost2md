package nga

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Network *network
	Config  *version
	Post    *post
}
type version struct {
	Version string
}
type post struct {
	Ip               bool
	Reply            bool
	Local_smile      bool
	Local_smile_path string
	Title_dir_name   bool
	Title_md_name    bool
}
type network struct {
	Base_url   string
	Ua         string
	Uid        string
	Cid        string
	Thread     int
	Page_limit int
}

func Read_cfg() Config {
	f := "config.toml"
	if _, err := os.Stat(f); err != nil {
		println("error")
	}
	tomlBytes, err := os.ReadFile("config.toml")
	if err != nil {
		fmt.Println(err)
	}

	var cfg Config
	err = toml.Unmarshal(tomlBytes, &cfg)
	if err != nil {
		println("err")
	}
	return cfg
}
