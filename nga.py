# -*- coding: UTF-8 -*-
import re
import requests
import os
import sys
import time
from contextlib import closing
import json
import nga_format

# =============先修改
headers = {
    'User-agent': 'Nga_Official/80023'}
cookies = {
    'ngaPassportUid': '__',
    'ngaPassportCid': '__',
}


totalfloor = []  # int几层，int pid, str时间，str昵称，str内容，int赞数
tid = 0
title = 'title'
localmaxpage = 1
localmaxfloor = -1
# (在single里用)部分楼层有评论，content是挂在被评论楼层的，所以先放在这里，之后判断当前楼层是否是评论楼层（是的话没有content），是的话就直接读成这里 int pid，str时间，str昵称，str内容，int赞数
commentreply = []
errortext = ''


def single(page):
    print('trypage%d' % page)
    params = (
        ('tid', tid),
        ('page', page),
        ('lite', 'js')
    )
    ss1 = requests.Session()
    get = ss1.get('https://bbs.nga.cn/read.php', headers=headers,
                  params=params, cookies=cookies)
    get.encoding = 'GBK'
    content = get.text.replace('	', '')  # 过滤掉防止json解析出错

    usertext = re.search(r',"__U":(.+?),"__R":', content, flags=re.S).group(1)
    userdict = json.loads(usertext, strict=False)

    replytext = re.search(r',"__R":(.+?),"__T":', content, flags=re.S).group(1)
    replydict = json.loads(replytext, strict=False)

    ttext = re.search(r',"__T":(.+?),"__F":', content, flags=re.S).group(1)
    tdict = json.loads(ttext, strict=False)

    global title
    title = tdict['subject']

    global commentreply
    for i in range(len(replydict)):
        one = ''
        if 'comment' in replydict[str(i)]:  # 该楼层下挂有评论，先+comment，下面到正经楼层
            for one in replydict[str(i)]['comment']:
                commentreply.append([int(replydict[str(i)]['comment'][one]['pid']), replydict[str(i)]['comment'][one]['postdate'], userdict[str(
                    replydict[str(i)]['comment'][one]['authorid'])]['username'], '[评论] ' + str(replydict[str(i)]['comment'][one]['content']), int(replydict[str(i)]['comment'][one]['score'])])

        if 'content' in replydict[str(i)]:  # 正经楼层
            commentnumtxt = ''
            if one != '':
                commentnumtxt = '[评论数:' + str(int(one) + 1) + ']\n\n'
            totalfloor.append([int(replydict[str(i)]['lou']), int(replydict[str(i)]['pid']), replydict[str(i)]['postdate'], userdict[str(
                replydict[str(i)]['authorid'])]['username'], commentnumtxt + str(replydict[str(i)]['content']), int(replydict[str(i)]['score'])])
        else:  # 评论楼层，无content
            for one in commentreply:
                if one[0] == int(replydict[str(i)]['pid']):
                    totalfloor.append([int(replydict[str(i)]['lou']), int(
                        replydict[str(i)]['pid']), one[1], one[2], one[3], one[4]])
                    commentreply.remove(one)

    # lastposter 对不上 “且”不是只有主楼的情况
    return int(tdict['replies']) > totalfloor[len(totalfloor)-1][0] and not(len(totalfloor) == 1 and totalfloor[0][0] == 0)


def makefile():
    global localmaxfloor
    global errortext
    lastfloor = 0
    total = totalfloor[len(totalfloor)-1][0]
    with open(('./%d/post.md' % tid), 'a', encoding='utf-8') as f:
        for onefloor in totalfloor:
            if localmaxfloor < int(onefloor[0]):
                if onefloor[0] == 0:
                    f.write(
                        '### %s\n\n(c) ludoux [GitHub Repo](https://github.com/ludoux/ngapost2md)\n\n' % title)

                f.write('----\n##### %d.[%d] \<pid:%d\> %s by %s\n' %
                        (onefloor[0], onefloor[5], onefloor[1], onefloor[2], onefloor[3]))
                raw = str(onefloor[4])

                rt = nga_format.format(raw,tid,onefloor[0],total,errortext)
                raw = rt[0]
                errortext = rt[1]
                
                f.write(('%s\n\n' % raw))
                lastfloor = int(onefloor[0])
    return lastfloor

def main():
    global tid
    if cookies['ngaPassportUid'][0] == '_' or cookies['ngaPassportCid'][0] == '_':
        print('Please edit cookies info in the code file first...')
    tid = int(input('tid:'))
    try:
        holder()
    except Exception as e:
        print('Oops! %s' % e)
    input('press to exit.')


def holder():
    global localmaxpage
    global localmaxfloor
    global errortext
    print(tid)
    if not os.path.exists(('./%d' % tid)):
        os.mkdir(('./%d' % tid))
    elif os.path.exists('./%d/max.txt' % tid):
        with open('./%d/max.txt' % tid, 'r', encoding='utf-8') as f:
            r = f.read()
            localmaxpage = int(r.split()[0])
            localmaxfloor = int(r.split()[1])

    print('localmaxpage%d\nlocalmaxfloor%d' % (localmaxpage, localmaxfloor))
    cpage = localmaxpage
    while single(cpage) != False:
        time.sleep(0.1)
        cpage = cpage + 1

    lastfloor = makefile()

    with open(('./%d/max.txt' % tid), 'w', encoding='utf-8') as f:
        f.write('%d %s' % (cpage, totalfloor[len(totalfloor) - 1][0]))

    if os.path.exists('./%d/info.txt' % tid):
        with open(('./%d/info.txt' % tid), 'a', encoding='utf-8') as f:
            f.write('[%s]%d Err:%s\n' % (time.asctime(
                time.localtime(time.time())), len(totalfloor), errortext))
    else:
        with open(('./%d/info.txt' % tid), 'w', encoding='utf-8') as f:
            f.write(
                'tid:%d\ntitle:%s\n(c) ludoux https://github.com/ludoux/ngapost2md\n==========\n' % (tid, title))
            f.write(
                ('[%s]%d Err:%s\n' % (time.asctime(time.localtime(time.time())), len(totalfloor), errortext)))

    print('makeuntil:%d' % lastfloor)


if __name__ == '__main__':
    main()