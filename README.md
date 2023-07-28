# qcip

**全新Golang重写版 比[Python版](https://github.com/cnlancehu/qcip/tree/python)运行速度快了11.3753秒**


自动设置**腾讯云服务器**的防火墙来源限制，使服务器重要端口只能被运行者的ip访问

目前已支持云产品: 腾讯云服务器 腾讯云轻量应用服务器

### 实现原理
qcip依赖于腾讯云api实现其功能，简单来说，是一个调用腾讯云api的脚本。它利用腾讯云安全组的来源限制功能(即防火墙规则只对指定IP生效)的功能，把指定防火墙规则的来源设置设定为本机(即运行这个程序的设备)的公网IPV4地址设定为限制的来源。这样，开放重要端口的防火墙只会对本机生效。

### 使用
#### 下载程序
在[Release](https://github.com/cnlancehu/qcip/releases "Release")页面下载您的系统所对应的程序

> 0.3.0版本以前的发布都是Python版本，不建议使用

#### 编辑配置
请事先在[腾讯云访问管理](https://console.cloud.tencent.com/cam/capi "腾讯云访问管理")创建API密钥

在刚刚下载下来的压缩包中，你可以找到配置文件`config.json`

编辑配置文件

![配置文件](https://github.com/cnlancehu/qcip/assets/106385654/c5c16c7d-1a1f-4d74-81e3-80ad505849b9 "配置填写教程")

`InstanceRegion`和`SecurityGroupRegion`的填写请参见下表
```
华北地区(北京) ap-beijing
西南地区(成都) ap-chengdu
西南地区(重庆) ap-chongqing
华南地区(广州) ap-guangzhou
港澳台地区(中国香港) ap-hongkong
亚太地区(首尔) ap-seoul
华东地区(上海) ap-shanghai
东南亚地区(新加坡) ap-singapore
欧洲地区(法兰克福) eu-frankfurt
美国西部(硅谷) na-siliconvalley
北美地区(多伦多) na-toronto
亚太地区(孟买) ap-mumbai
美国东部(弗吉尼亚) na-ashburn
亚太地区(曼谷) ap-bangkok
亚太地区(东京) ap-tokyo
华东地区(南京) ap-nanjing
亚太地区(雅加达) ap-jakarta
南美地区(圣保罗) sa-saopaulo
```

目前可用的获取IP的API有 `LanceAPI` `IPIP` `SB`
```
LanceAPI // 推荐在海外使用
https://api.lance.fun/ip/

IPIP // 推荐在中国大陆使用
https://myip.ipip.net/ip

SB // 全球通用 但效率较慢
https://api-ipv4.ip.sb/ip
```

填写时请注意区分大小写

#### 运行
使用命令行运行
```bash
qcip [配置文件路径(可选，默认为config.json)]`
```

> **注意** 若你使用 桌面系统 直接双击打开程序，会出现命令行窗口和闪退现象，这并不代表程序运行失败，但是是你无法看到运行结果

#### 开机启动(Windows)

你也可以把程序文件放入电脑的启动项中，这样，每次开机时，脚本就会自动运行

你可以在下面的目录(启动项文件夹)中添加该脚本的快捷方式或运行程序的批处理文件(bat)
`C:\ProgramData\Microsoft\Windows\Start Menu\Programs\Startup`

使用以下脚本运行以隐藏命令行窗口

```
// qcip.vbs
Set WshShell = CreateObject("WScript.Shell")
WshShell.Run "cmd /c /*运行的命令*/", 0, False
```