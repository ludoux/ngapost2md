# -*- coding: UTF-8 -*-
import re
import requests
import os
import sys
import hashlib
import time
from contextlib import closing

smile_ac = {
    "blink": "ac0.png",
    "goodjob": "ac1.png",
    "上": "ac2.png",
    "中枪": "ac3.png",
    "偷笑": "ac4.png",
    "冷": "ac5.png",
    "凌乱": "ac6.png",
    "吓": "ac8.png",
    "吻": "ac9.png",
    "呆": "ac10.png",
    "咦": "ac11.png",
    "哦": "ac12.png",
    "哭": "ac13.png",
    "哭1": "ac14.png",
    "哭笑": "ac15.png",
    "喘": "ac17.png",
    "心": "ac23.png",

    "囧": "ac21.png",
    "晕": "ac33.png",
    "汗": "ac34.png",
    "瞎": "ac35.png",
    "羞": "ac36.png",
    "羡慕": "ac37.png",

    "委屈": "ac22.png",
    "忧伤": "ac24.png",
    "怒": "ac25.png",
    "怕": "ac26.png",
    "惊": "ac27.png",
    "愁": "ac28.png",
    "抓狂": "ac29.png",
    "哼": "ac16.png",
    "喷": "ac18.png",
    "嘲笑": "ac19.png",
    "嘲笑1": "ac20.png",

    "抠鼻": "ac30.png",
    "无语": "ac32.png",
    "衰": "ac40.png",

    "黑枪": "ac44.png",
    "花痴": "ac38.png",
    "闪光": "ac43.png",
    "擦汗": "ac31.png",
    "茶": "ac39.png",
    "计划通": "ac41.png",
    "反对": "ac7.png",
    "赞同": "ac42.png"}

smile_a2 = {
    "goodjob": "a2_02.png",
    "诶嘿": "a2_05.png",
    "偷笑": "a2_03.png",
    "怒": "a2_04.png",
    "笑": "a2_07.png",
    "那个…": "a2_08.png",
    "哦嗬嗬嗬": "a2_09.png",
    "舔": "a2_10.png",
    "鬼脸": "a2_14.png",
    "冷": "a2_16.png",
    "大哭": "a2_15.png",
    "哭": "a2_17.png",
    "恨": "a2_21.png",
    "中枪": "a2_23.png",
    "囧": "a2_24.png",
    "你看看你": "a2_25.png",
    "doge": "a2_27.png",
    "自戳双目": "a2_28.png",
    "偷吃": "a2_30.png",
    "冷笑": "a2_31.png",
    "壁咚": "a2_32.png",
    "不活了": "a2_33.png",
    "不明觉厉": "a2_36.png",
    "是在下输了": "a2_51.png",
    "你为猴这么": "a2_53.png",
    "干杯": "a2_54.png",
    "干杯2": "a2_55.png",
    "异议": "a2_47.png",
    "认真": "a2_48.png",
    "你已经死了": "a2_45.png",
    "你这种人…": "a2_49.png",
    "妮可妮可妮": "a2_18.png",
    "惊": "a2_19.png",
    "抢镜头": "a2_52.png",
    "yes": "a2_26.png",
    "有何贵干": "a2_11.png",
    "病娇": "a2_12.png",
    "lucky": "a2_13.png",
    "poi": "a2_20.png",
    "囧2": "a2_22.png",
    "威吓": "a2_42.png",
    "jojo立": "a2_37.png",
    "jojo立2": "a2_38.png",
    "jojo立3": "a2_39.png",
    "jojo立4": "a2_41.png",
    "jojo立5": "a2_40.png"
}

smile_pst = {
    "举手": "pt00.png",
    "亲": "pt01.png",
    "偷笑": "pt02.png",
    "偷笑2": "pt03.png",
    "偷笑3": "pt04.png",
    "傻眼": "pt05.png",
    "傻眼2": "pt06.png",
    "兔子": "pt07.png",
    "发光": "pt08.png",
    "呆": "pt09.png",
    "呆2": "pt10.png",
    "呆3": "pt11.png",
    "呕": "pt12.png",
    "呵欠": "pt13.png",
    "哭": "pt14.png",
    "哭2": "pt15.png",
    "哭3": "pt16.png",
    "嘲笑": "pt17.png",
    "基": "pt18.png",
    "宅": "pt19.png",
    "安慰": "pt20.png",
    "幸福": "pt21.png",
    "开心": "pt22.png",
    "开心2": "pt23.png",
    "开心3": "pt24.png",
    "怀疑": "pt25.png",
    "怒": "pt26.png",
    "怒2": "pt27.png",
    "怨": "pt28.png",
    "惊吓": "pt29.png",
    "惊吓2": "pt30.png",
    "惊呆": "pt31.png",
    "惊呆2": "pt32.png",
    "惊呆3": "pt33.png",
    "惨": "pt34.png",
    "斜眼": "pt35.png",
    "晕": "pt36.png",
    "汗": "pt37.png",
    "泪": "pt38.png",
    "泪2": "pt39.png",
    "泪3": "pt40.png",
    "泪4": "pt41.png",
    "满足": "pt42.png",
    "满足2": "pt43.png",
    "火星": "pt44.png",
    "牙疼": "pt45.png",
    "电击": "pt46.png",
    "看戏": "pt47.png",
    "眼袋": "pt48.png",
    "眼镜": "pt49.png",
    "笑而不语": "pt50.png",
    "紧张": "pt51.png",
    "美味": "pt52.png",
    "背": "pt53.png",
    "脸红": "pt54.png",
    "脸红2": "pt55.png",
    "腐": "pt56.png",
    "星星眼": "pt57.png",
    "谢": "pt58.png",
    "醉": "pt59.png",
    "闷": "pt60.png",
    "闷2": "pt61.png",
    "音乐": "pt62.png",
    "黑脸": "pt63.png",
    "鼻血": "pt64.png"
}

smile_dt = {
    "ROLL": "dt01.png",
    "上": "dt02.png",
    "傲娇": "dt03.png",
    "叉出去": "dt04.png",
    "发光": "dt05.png",
    "呵欠": "dt06.png",
    "哭": "dt07.png",
    "啃古头": "dt08.png",
    "嘲笑": "dt09.png",
    "心": "dt10.png",
    "怒": "dt11.png",
    "怒2": "dt12.png",
    "怨": "dt13.png",
    "惊": "dt14.png",
    "惊2": "dt15.png",
    "无语": "dt16.png",
    "星星眼": "dt17.png",
    "星星眼2": "dt18.png",
    "晕": "dt19.png",
    "注意": "dt20.png",
    "注意2": "dt21.png",
    "泪": "dt22.png",
    "泪2": "dt23.png",
    "烧": "dt24.png",
    "笑": "dt25.png",
    "笑2": "dt26.png",
    "笑3": "dt27.png",
    "脸红": "dt28.png",
    "药": "dt29.png",
    "衰": "dt30.png",
    "鄙视": "dt31.png",
    "闲": "dt32.png",
    "黑脸": "dt33.png"
}

smile_pg = {
    "战斗力": "pg01.png",
    "哈啤": "pg02.png",
    "满分": "pg03.png",
    "衰": "pg04.png",
    "拒绝": "pg05.png",
    "心": "pg06.png",
    "严肃": "pg07.png",
    "吃瓜": "pg08.png",
    "嘣": "pg09.png",
    "嘣2": "pg10.png",
    "冻": "pg11.png",
    "谢": "pg12.png",
    "哭": "pg13.png",
    "响指": "pg14.png",
    "转身": "pg15.png"
}

errortext = ''
appendpid = []  # 里面是int


def util_down(url, path, filename, prestr=''):
    time.sleep(0.1)
    global errortext
    fullpath = path + '/' + prestr + filename
    try:
        with closing(requests.get(url, stream=True)) as response:
            chunk_size = 1024  # 单次请求最大值
            with open(fullpath, 'wb') as file:
                for data in response.iter_content(chunk_size=chunk_size):
                    file.write(data)
    except Exception as e:
        print('Failed to down url:%s, path:%s:%s' % (url, fullpath, e))
        errortext = errortext + \
            '<Failed to down url:%s, path:%s>' % (url, fullpath)


def smile(raw):
    rex = re.findall(r'\[s\:ac\:(.+?)\]', raw)
    for ritem in rex:
        raw = raw.replace('[s:ac:%s]' % (
            ritem), '![%s](https://img4.nga.178.com/ngabbs/post/smile/%s)' % (ritem, smile_ac[ritem]))

    rex = re.findall(r'\[s\:a2\:(.+?)\]', raw)
    for ritem in rex:
        raw = raw.replace('[s:a2:%s]' % (
            ritem), '![%s](https://img4.nga.178.com/ngabbs/post/smile/%s)' % (ritem, smile_a2[ritem]))

    rex = re.findall(r'\[s\:pst\:(.+?)\]', raw)
    for ritem in rex:
        raw = raw.replace('[s:pst:%s]' % (
            ritem), '![%s](https://img4.nga.178.com/ngabbs/post/smile/%s)' % (ritem, smile_pst[ritem]))

    rex = re.findall(r'\[s\:dt\:(.+?)\]', raw)
    for ritem in rex:
        raw = raw.replace('[s:dt:%s]' % (
            ritem), '![%s](https://img4.nga.178.com/ngabbs/post/smile/%s)' % (ritem, smile_dt[ritem]))

    rex = re.findall(r'\[s\:pg\:(.+?)\]', raw)
    for ritem in rex:
        raw = raw.replace('[s:pg:%s]' % (
            ritem), '![%s](https://img4.nga.178.com/ngabbs/post/smile/%s)' % (ritem, smile_pg[ritem]))

    return raw


def newline(raw):
    return raw.replace('<br/>', '\n').replace('<br>', '\n').replace('&lt;br/&gt;','\n').replace('&lt;br&gt;','\n')


def pic(raw, tid, floorindex, total):
    rex = re.findall(r'(?<=\[img\]).+?(?=\[/img\])', raw)
    for ritem in rex:
        url = str(ritem)
        if url[0:2] == './':
            url = 'https://img.nga.178.com/attachments/' + url[2:]
        url = url.replace('.medium.jpg', '')
        filename = hashlib.md5(
            bytes(url, encoding='utf-8')).hexdigest()[2:8] + url[-6:]
        if os.path.exists('./%d/%s' % (tid, filename)) == False:
            util_down(url, ('./%d' % tid), filename, '')
            print('down pic:./%d/%s Floor[%d/%d]' %
                  (tid, filename, floorindex, total))
        raw = raw.replace(('[img]%s[/img]' %
                           ritem), ('![img](./%s)' % filename))
    return raw


def quote(raw):
    global appendpid
    # [quote][pid=446671245,23044506,3]Reply[/pid] [b]Post by [uid]#anony_e8992fb55425a90ff0ff409eb4981c96[/uid][color=gray](43楼)[/color] (2020-08-21 03:32):[/b]<br/><br/>被霸凌欺负过，磕过安眠药洗过胃。<br/>你问我为什么不极限一换一？<br/>家人/老师/同学都认为是你的错，<br/>你也以为是自己的错，扭曲的环境下，<br/>我只有自杀一条路可以走。[/quote]
    # [0]人名 [1]时间 [2]圈的内容
    # 引用 [quote][tid=0000000]Topic[/tid] [b]Post by [uid=000000]whowhowho[/uid] (2020-03-26 01:07):[/b]
    ro0 = re.compile(
        r'\[quote\]\[tid=.+?\[uid.*?\](.+?)\[/uid\].*?\((\d{4}.+?)\):\[/b\](.+?)\[/quote\]((?:\n){0,2})', flags=re.S)  # 这个是圈主帖的
    rex = ro0.findall(raw)
    for ritem in rex:
        quotetext = ritem[2]
        quotetext = quotetext.replace('\n', '\n>')
        quoteauthor = ritem[0]
        if quoteauthor[:7] == '#anony_':
            # TODO: https://img4.nga.178.com/common_res/js_commonui.js commonui.anonyName 之后再整
            quoteauthor = '匿' + quoteauthor[-6:]
        raw = ro0.sub('>[jump](#pid0) %s(%s) said:%s\n\n' %
                      (quoteauthor, ritem[1], quotetext), raw)

    quoteCount = raw.count("[quote]")
    for x in range(quoteCount):
        end = raw.find('[/quote]')
        start = 0
        for m in re.finditer('\[quote\]', raw):
            cur = m.start()
            if cur < end:
                start = cur
            else:
                break

        clip = raw[start:end+8]
        ro1 = re.compile(
            r'\[quote\]\[pid=(\d+?),.+?\[uid.*?\](.+?)\[/uid\].*?\((\d{4}.+?)\):\[/b\](.+?)\[/quote\]((?:\n){0,2})', flags=re.S)
        # [0]pid [1]原作者 [2]时间 [3]说的东西
        rex = ro1.findall(clip)
        for ritem in rex:
            quotetext = ritem[3]
            quotetext = quotetext.replace('\n', '\n>')
            quoteauthor = ritem[1]
            # appendpid.append(int(ritem[0])) #这里会有原文的，就不append了
            if quoteauthor[:7] == '#anony_':
                # TODO: https://img4.nga.178.com/common_res/js_commonui.js commonui.anonyName 之后再整
                quoteauthor = '匿' + quoteauthor[-6:]
            raw = raw.replace(clip,'>[jump](#pid%s) %s(%s) said:%s\n\n' %
                        (ritem[0], quoteauthor, ritem[2], quotetext))
            # raw = raw.replace(re.search(r'\[quote\].+?\[uid.*?\](.+?)\[/uid\].*?\((\d{4}.+?)\):\[/b\](.+?)\[/quote\]',
            # raw, flags=re.S).group(), '>%s(%s) said:%s' % (quoteauthor, ritem[1], quotetext))

    ro2 = re.compile(
        r'\[b\]Reply to \[pid=(\d+?),.+? Post by \[uid.*?\](.+?)\[\/uid\].+?\((.+?)\)\[\/b\]((?:\n){0,2})', flags=re.S)
    # [0]pid [1]原作者 [2]时间
    rex = ro2.findall(raw)
    for ritem in rex:
        appendpid.append(int(ritem[0]))
        raw = ro2.sub('>[jump](#pid%s) Reply to %s(%s):\n\n' %
                      (ritem[0], ritem[1], ritem[2]), raw)

    return raw


def strikeout(raw):
    return raw.replace('[del]', '~~').replace('[/del]', '~~')


def url(raw):
    rex = re.findall(r'\[url\](.+?)\[\/url\]', raw)
    for ritem in rex:
        raw = raw.replace('[url]%s[/url]' % ritem, '[url](%s)' % ritem)
    rex = re.findall(r'\[url=(.+?)\](.+?)\[\/url\]', raw)
    for ritem in rex:
        raw = raw.replace('[url=%s]%s[/url]' % (ritem[0],
                                                ritem[1]), '[%s](%s)' % (ritem[1], ritem[0]))
    return raw


def align(raw):
    rex = re.findall(r'\[align=(.+?)\](.+?)\[\/align\]', raw)
    for ritem in rex:
        raw = raw.replace('[align=%s]%s[/align]' % (ritem[0], ritem[1]),
                          '<div style="text-align:%s">%s</div>' % (ritem[0], ritem[1]))
    return raw


def collapse(raw):
    rex = re.findall(
        r'\[collapse(=.+?)?\](.+?)\[\/collapse\]', raw, flags=re.S)
    rt = ''
    for ritem in rex:
        if ritem[0] == '':
            rt = '<details>\n  <summary>已折叠，点击展开</summary>\n  <pre>' + \
                ritem[0].replace('\n', '<br>') + '</pre>\n</details>'
            raw = raw.replace('[collapse]%s[/collapse]' % ritem[1], rt)
        else:
            rt = '<details>\n  <summary>' + \
                ritem[0][1:] + '</summary>\n  <pre>' + \
                ritem[1].replace('\n', '<br>') + '</pre>\n</details>'
            raw = raw.replace('[collapse%s]%s[/collapse]' %
                              (ritem[0], ritem[1]), rt)
    return raw


def anony(raw):
    # Special thanks to @crella6
    anony_string1 = '甲乙丙丁戊己庚辛壬癸子丑寅卯辰巳午未申酉戌亥'
    anony_string2 = '王李张刘陈杨黄吴赵周徐孙马朱胡林郭何高罗郑梁谢宋唐许邓冯韩曹曾彭萧蔡潘田董袁于余叶蒋杜苏魏程吕丁沈任姚卢傅钟姜崔谭廖范汪陆金石戴贾韦夏邱方侯邹熊孟秦白江阎薛尹段雷黎史龙陶贺顾毛郝龚邵万钱严赖覃洪武莫孔汤向常温康施文牛樊葛邢安齐易乔伍庞颜倪庄聂章鲁岳翟殷詹申欧耿关兰焦俞左柳甘祝包宁尚符舒阮柯纪梅童凌毕单季裴霍涂成苗谷盛曲翁冉骆蓝路游辛靳管柴蒙鲍华喻祁蒲房滕屈饶解牟艾尤阳时穆农司卓古吉缪简车项连芦麦褚娄窦戚岑景党宫费卜冷晏席卫米柏宗瞿桂全佟应臧闵苟邬边卞姬师和仇栾隋商刁沙荣巫寇桑郎甄丛仲虞敖巩明佘池查麻苑迟邝'
    rex = re.findall(r'#anony_.{32}', raw)
    for aname in rex:
        i = 6
        res = ''
        for j in range(6):
            if j == 0 or j == 3:
                if int('0x0' + aname[i+1], 16) < len(anony_string1):
                    res = res + anony_string1[int('0x0' + aname[i+1], 16)]
            else:
                if int('0x' + aname[i:i+2], 16) < len(anony_string2):
                    res = res + anony_string2[int('0x' + aname[i:i+2], 16)]
            i = i+2
        res = res + '?'
        raw = raw.replace(aname, res)
    return raw


def audio(raw, tid, floorindex, total):
    rex = re.findall(r'(?<=\[flash=audio\]).+?(?=\[/flash\])', raw)
    for ritem in rex:
        dura = ritem[ritem.find('?duration=')+10:]
        ori = ritem
        ritem = ritem[:ritem.find('?duration=')]
        url = str(ritem)
        if url[0:2] == './':
            url = 'https://img.nga.178.com/attachments/' + url[2:]
        filename = hashlib.md5(
            bytes(url, encoding='utf-8')).hexdigest()[2:8] + url[-6:]
        if os.path.exists('./%d/%s' % (tid, filename)) == False:
            util_down(url, ('./%d' % tid), filename, str(floorindex) + '_')
            print('down audio:./%d/%s Floor[%d/%d]' %
                  (tid, filename, floorindex, total))
        raw = raw.replace(('[flash=audio]%s[/flash]' %
                           ori), ('<存在一音频: %s , %s>' % (str(floorindex) + '_' + filename, dura)))
    return raw


def format(raw, tid, floorindex, total, errtxt):
    global errortext
    global appendpid  # 需要主程序追加在后面的pid的正文，这个在quote里面修改（里面都是int
    errortext = errtxt
    appendpid.clear()
    try:
        raw = newline(raw)
        raw = anony(raw)
        raw = pic(raw, tid, floorindex, total)
        raw = audio(raw, tid, floorindex, total)
        raw = smile(raw)
        raw = quote(raw)
        raw = strikeout(raw)
        raw = url(raw)
        raw = align(raw)
        raw = collapse(raw)
    except Exception as e:
        print('Error occured (@F.%d): %s' % (floorindex, e))
        errortext = errortext + 'Error occured (@F.%d).' % floorindex
        return raw, errortext, appendpid
    else:
        return raw, errortext, appendpid
