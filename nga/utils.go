package nga

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/imroc/req/v3"
	"github.com/spf13/cast"
)

func getSmile(key string) string {
	smileStr := `{"[s:ac:blink]":"ac0.png","[s:ac:goodjob]":"ac1.png","[s:ac:上]":"ac2.png","[s:ac:中枪]":"ac3.png","[s:ac:偷笑]":"ac4.png","[s:ac:冷]":"ac5.png","[s:ac:凌乱]":"ac6.png","[s:ac:吓]":"ac8.png","[s:ac:吻]":"ac9.png","[s:ac:呆]":"ac10.png","[s:ac:咦]":"ac11.png","[s:ac:哦]":"ac12.png","[s:ac:哭]":"ac13.png","[s:ac:哭1]":"ac14.png","[s:ac:哭笑]":"ac15.png","[s:ac:喘]":"ac17.png","[s:ac:心]":"ac23.png","[s:ac:囧]":"ac21.png","[s:ac:晕]":"ac33.png","[s:ac:汗]":"ac34.png","[s:ac:瞎]":"ac35.png","[s:ac:羞]":"ac36.png","[s:ac:羡慕]":"ac37.png","[s:ac:委屈]":"ac22.png","[s:ac:忧伤]":"ac24.png","[s:ac:怒]":"ac25.png","[s:ac:怕]":"ac26.png","[s:ac:惊]":"ac27.png","[s:ac:愁]":"ac28.png","[s:ac:抓狂]":"ac29.png","[s:ac:哼]":"ac16.png","[s:ac:喷]":"ac18.png","[s:ac:嘲笑]":"ac19.png","[s:ac:嘲笑1]":"ac20.png","[s:ac:抠鼻]":"ac30.png","[s:ac:无语]":"ac32.png","[s:ac:衰]":"ac40.png","[s:ac:黑枪]":"ac44.png","[s:ac:花痴]":"ac38.png","[s:ac:闪光]":"ac43.png","[s:ac:擦汗]":"ac31.png","[s:ac:茶]":"ac39.png","[s:ac:计划通]":"ac41.png","[s:ac:反对]":"ac7.png","[s:ac:赞同]":"ac42.png","[s:a2:goodjob]":"a2_02.png","[s:a2:诶嘿]":"a2_05.png","[s:a2:偷笑]":"a2_03.png","[s:a2:怒]":"a2_04.png","[s:a2:笑]":"a2_07.png","[s:a2:那个…]":"a2_08.png","[s:a2:哦嗬嗬嗬]":"a2_09.png","[s:a2:舔]":"a2_10.png","[s:a2:鬼脸]":"a2_14.png","[s:a2:冷]":"a2_16.png","[s:a2:大哭]":"a2_15.png","[s:a2:哭]":"a2_17.png","[s:a2:恨]":"a2_21.png","[s:a2:中枪]":"a2_23.png","[s:a2:囧]":"a2_24.png","[s:a2:你看看你]":"a2_25.png","[s:a2:doge]":"a2_27.png","[s:a2:自戳双目]":"a2_28.png","[s:a2:偷吃]":"a2_30.png","[s:a2:冷笑]":"a2_31.png","[s:a2:壁咚]":"a2_32.png","[s:a2:不活了]":"a2_33.png","[s:a2:不明觉厉]":"a2_36.png","[s:a2:是在下输了]":"a2_51.png","[s:a2:你为猴这么]":"a2_53.png","[s:a2:干杯]":"a2_54.png","[s:a2:干杯2]":"a2_55.png","[s:a2:异议]":"a2_47.png","[s:a2:认真]":"a2_48.png","[s:a2:你已经死了]":"a2_45.png","[s:a2:你这种人…]":"a2_49.png","[s:a2:妮可妮可妮]":"a2_18.png","[s:a2:惊]":"a2_19.png","[s:a2:抢镜头]":"a2_52.png","[s:a2:yes]":"a2_26.png","[s:a2:有何贵干]":"a2_11.png","[s:a2:病娇]":"a2_12.png","[s:a2:lucky]":"a2_13.png","[s:a2:poi]":"a2_20.png","[s:a2:囧2]":"a2_22.png","[s:a2:威吓]":"a2_42.png","[s:a2:jojo立]":"a2_37.png","[s:a2:jojo立2]":"a2_38.png","[s:a2:jojo立3]":"a2_39.png","[s:a2:jojo立4]":"a2_41.png","[s:a2:jojo立5]":"a2_40.png","[s:pst:举手]":"pt00.png","[s:pst:亲]":"pt01.png","[s:pst:偷笑]":"pt02.png","[s:pst:偷笑2]":"pt03.png","[s:pst:偷笑3]":"pt04.png","[s:pst:傻眼]":"pt05.png","[s:pst:傻眼2]":"pt06.png","[s:pst:兔子]":"pt07.png","[s:pst:发光]":"pt08.png","[s:pst:呆]":"pt09.png","[s:pst:呆2]":"pt10.png","[s:pst:呆3]":"pt11.png","[s:pst:呕]":"pt12.png","[s:pst:呵欠]":"pt13.png","[s:pst:哭]":"pt14.png","[s:pst:哭2]":"pt15.png","[s:pst:哭3]":"pt16.png","[s:pst:嘲笑]":"pt17.png","[s:pst:基]":"pt18.png","[s:pst:宅]":"pt19.png","[s:pst:安慰]":"pt20.png","[s:pst:幸福]":"pt21.png","[s:pst:开心]":"pt22.png","[s:pst:开心2]":"pt23.png","[s:pst:开心3]":"pt24.png","[s:pst:怀疑]":"pt25.png","[s:pst:怒]":"pt26.png","[s:pst:怒2]":"pt27.png","[s:pst:怨]":"pt28.png","[s:pst:惊吓]":"pt29.png","[s:pst:惊吓2]":"pt30.png","[s:pst:惊呆]":"pt31.png","[s:pst:惊呆2]":"pt32.png","[s:pst:惊呆3]":"pt33.png","[s:pst:惨]":"pt34.png","[s:pst:斜眼]":"pt35.png","[s:pst:晕]":"pt36.png","[s:pst:汗]":"pt37.png","[s:pst:泪]":"pt38.png","[s:pst:泪2]":"pt39.png","[s:pst:泪3]":"pt40.png","[s:pst:泪4]":"pt41.png","[s:pst:满足]":"pt42.png","[s:pst:满足2]":"pt43.png","[s:pst:火星]":"pt44.png","[s:pst:牙疼]":"pt45.png","[s:pst:电击]":"pt46.png","[s:pst:看戏]":"pt47.png","[s:pst:眼袋]":"pt48.png","[s:pst:眼镜]":"pt49.png","[s:pst:笑而不语]":"pt50.png","[s:pst:紧张]":"pt51.png","[s:pst:美味]":"pt52.png","[s:pst:背]":"pt53.png","[s:pst:脸红]":"pt54.png","[s:pst:脸红2]":"pt55.png","[s:pst:腐]":"pt56.png","[s:pst:星星眼]":"pt57.png","[s:pst:谢]":"pt58.png","[s:pst:醉]":"pt59.png","[s:pst:闷]":"pt60.png","[s:pst:闷2]":"pt61.png","[s:pst:音乐]":"pt62.png","[s:pst:黑脸]":"pt63.png","[s:pst:鼻血]":"pt64.png","[s:dt:ROLL]":"dt01.png","[s:dt:上]":"dt02.png","[s:dt:傲娇]":"dt03.png","[s:dt:叉出去]":"dt04.png","[s:dt:发光]":"dt05.png","[s:dt:呵欠]":"dt06.png","[s:dt:哭]":"dt07.png","[s:dt:啃古头]":"dt08.png","[s:dt:嘲笑]":"dt09.png","[s:dt:心]":"dt10.png","[s:dt:怒]":"dt11.png","[s:dt:怒2]":"dt12.png","[s:dt:怨]":"dt13.png","[s:dt:惊]":"dt14.png","[s:dt:惊2]":"dt15.png","[s:dt:无语]":"dt16.png","[s:dt:星星眼]":"dt17.png","[s:dt:星星眼2]":"dt18.png","[s:dt:晕]":"dt19.png","[s:dt:注意]":"dt20.png","[s:dt:注意2]":"dt21.png","[s:dt:泪]":"dt22.png","[s:dt:泪2]":"dt23.png","[s:dt:烧]":"dt24.png","[s:dt:笑]":"dt25.png","[s:dt:笑2]":"dt26.png","[s:dt:笑3]":"dt27.png","[s:dt:脸红]":"dt28.png","[s:dt:药]":"dt29.png","[s:dt:衰]":"dt30.png","[s:dt:鄙视]":"dt31.png","[s:dt:闲]":"dt32.png","[s:dt:黑脸]":"dt33.png","[s:pg:战斗力]":"pg01.png","[s:pg:哈啤]":"pg02.png","[s:pg:满分]":"pg03.png","[s:pg:衰]":"pg04.png","[s:pg:拒绝]":"pg05.png","[s:pg:心]":"pg06.png","[s:pg:严肃]":"pg07.png","[s:pg:吃瓜]":"pg08.png","[s:pg:嘣]":"pg09.png","[s:pg:嘣2]":"pg10.png","[s:pg:冻]":"pg11.png","[s:pg:谢]":"pg12.png","[s:pg:哭]":"pg13.png","[s:pg:响指]":"pg14.png","[s:pg:转身]":"pg15.png"}`

	r, _ := jsonparser.GetString([]byte(smileStr), key)
	return r
}

func ts2t(ts int64) string {
	return time.Unix(ts, 0).Format(time.DateTime)
}

func ts() int64 {
	return time.Now().Unix()
}

func anony(it string) string {
	// Special thanks to @crella6
	anonyRune1 := []rune("甲乙丙丁戊己庚辛壬癸子丑寅卯辰巳午未申酉戌亥")
	anonyRune2 := []rune("王李张刘陈杨黄吴赵周徐孙马朱胡林郭何高罗郑梁谢宋唐许邓冯韩曹曾彭萧蔡潘田董袁于余叶蒋杜苏魏程吕丁沈任姚卢傅钟姜崔谭廖范汪陆金石戴贾韦夏邱方侯邹熊孟秦白江阎薛尹段雷黎史龙陶贺顾毛郝龚邵万钱严赖覃洪武莫孔汤向常温康施文牛樊葛邢安齐易乔伍庞颜倪庄聂章鲁岳翟殷詹申欧耿关兰焦俞左柳甘祝包宁尚符舒阮柯纪梅童凌毕单季裴霍涂成苗谷盛曲翁冉骆蓝路游辛靳管柴蒙鲍华喻祁蒲房滕屈饶解牟艾尤阳时穆农司卓古吉缪简车项连芦麦褚娄窦戚岑景党宫费卜冷晏席卫米柏宗瞿桂全佟应臧闵苟邬边卞姬师和仇栾隋商刁沙荣巫寇桑郎甄丛仲虞敖巩明佘池查麻苑迟邝")
	i := 6
	res := ""
	for j := 0; j < 6; j++ {
		if j == 0 || j == 3 {
			n, _ := strconv.ParseUint(("0" + it[i+1:i+2]), 16, 32)
			if int(n) < len(anonyRune1) {
				res += string(anonyRune1[n : n+1])
			}
		} else {
			n, _ := strconv.ParseUint(("" + it[i:i+2]), 16, 32)
			if int(n) < len(anonyRune2) {
				res += string(anonyRune2[n : n+1])
			}
		}
		i += 2
	}
	res += "?"
	return res
}

func downloadAssets(url string, fileName string) {
	client := req.C()

	// Download to the absolute file path.
	client.R().SetOutputFile(fileName).Get(url)
}

func ToSaveFilename(in string) string {
	//https://stackoverflow.com/questions/1976007/what-characters-are-forbidden-in-windows-and-linux-directory-names
	rp := strings.NewReplacer(
		"/", " ",
		"\\", " ",
		"<", " ",
		">", " ",
		":", " ",
		"\"", " ",
		"|", " ",
		"?", " ",
		"*", " ",
	)
	rt := rp.Replace(in)
	return rt
}

// 找不到时会直接返回 ""
func FindFolderNameByTid(tid int) string {
	fi, err := os.Stat(cast.ToString(tid))
	if err == nil && fi.IsDir() {
		return cast.ToString(tid)
	} else if err != nil && os.IsNotExist(err) {
		return ""
	} else if err != nil {
		log.Fatalln(err.Error())
	}
	matches, err := filepath.Glob(fmt.Sprintf("./%d*-", tid))
	if err != nil {
		log.Fatalln(err.Error())
		return ""
	}
	if len(matches) == 1 {
		return matches[0]
	} else if len(matches) > 1 {
		log.Fatalln("有多个文件夹匹配:", fmt.Sprintf("./%d*-", tid))
		return ""
	}
	return ""
}
