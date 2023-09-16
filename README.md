# PacMan
windows系统下go语言实现一个终端运行吃豆人小游戏

系统：windows10

语言：golang

编译器：GoLand


![QQ图片20230429145011](https://user-images.githubusercontent.com/111231983/235289192-15794acf-99ed-43de-b042-11e217df7936.png)

玩法：吃豆人吃地图中的豆子，吃的过程中会有怪物来袭击，躲过怪物袭击并且吃完所有豆子游戏即获得胜利。

吃豆人只有三条生命，被怪兽成功袭击一次就减少一条生命值，生命值为0时，游戏就结束。

地图中吃一个普通豆子得分加1，吃一个能量药丸得分加10，且怪物会变为10秒内无法攻击吃豆人的状态。

操作：键盘的UP、DOWN、LEFT、RIGHT键控制吃豆人的上下左右移动


文件介绍：

config.json:为地图配置彩色表情

config.noemoji.json:无彩色表情配置

go.mod:版本控制，引入的外部库

main.go:游戏的总代码
代码中每个功能都有相应的注释

maze01.txt:字符串地图，可以根据自己需要更改

地图中各字符含义：

“#” 代表墙

“.” 代表豆子

“P” 代表玩家

“G” 代表怪兽

“X” 代表能量药丸


PacGp.exe:代码编译过的exe可执行文件





