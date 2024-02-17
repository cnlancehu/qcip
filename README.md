# qcip

自动设置**腾讯云服务器** **阿里云服务器**的防火墙来源限制，使服务器重要端口只能被运行者的ip访问

目前已支持云产品: 腾讯云服务器 腾讯云轻量应用服务器 阿里云轻量应用服务器

### 状态

![状态](https://api.lance.fun/proj/qcip/status)

### 实现原理
qcip依赖于云服务商api实现其功能，简单来说，是一个调用腾讯云api的脚本。它利用腾讯云安全组的来源限制功能(即防火墙规则只对指定IP生效)的功能，把指定防火墙规则的来源设置设定为本机(即运行这个程序的设备)的公网IPv4地址。这样，开放重要端口的防火墙规则只会对本机生效。

### 使用
#### 下载程序
在[Release](https://github.com/cnlancehu/qcip/releases "Release")页面下载系统所对应的程序

> 0.3.0版本以前的发布都是[Python版本](https://github.com/cnlancehu/qcip/tree/python)，不建议使用

#### 编辑配置
请事先在[腾讯云访问管理](https://console.cloud.tencent.com/cam/capi "腾讯云访问管理")或者[阿里云访问管理](https://ram.console.aliyun.com/manage/ak "阿里云访问管理")创建API密钥

在刚刚下载下来的压缩包中，你可以找到配置文件`config.json`

编辑配置文件

![编辑配置文件](https://github.com/cnlancehu/qcip/assets/106385654/66a83ddc-f034-441f-879c-1c0f9fa19390 "配置填写教程")

其中 阿里云轻量应用服务器的 `MType` 字段为 `allh`，其他部分与腾讯云轻量应用服务器同理

`InstanceRegion`和`SecurityGroupRegion`的填写请参见下表

> **腾讯云**
> 
> 华北地区(北京) ap-beijing
>
> 西南地区(成都) ap-chengdu
>
> 西南地区(重庆) ap-chongqing
>
> 华南地区(广州) ap-guangzhou
>
> 港澳台地区(中国香港) ap-hongkong
>
> 亚太地区(首尔) ap-seoul
>
> 华东地区(上海) ap-shanghai
>
> 东南亚地区(新加坡) ap-singapore
>
> 欧洲地区(法兰克福) eu-frankfurt
>
> 美国西部(硅谷) na-siliconvalley
>
> 北美地区(多伦多) na-toronto
>
> 亚太地区(孟买) ap-mumbai
>
> 美国东部(弗吉尼亚) na-ashburn
>
> 亚太地区(曼谷) ap-bangkok
>
> 亚太地区(东京) ap-tokyo
>
> 华东地区(南京) ap-nanjing
>
> 亚太地区(雅加达) ap-jakarta
>
> 南美地区(圣保罗) sa-saopaulo
> 
> **阿里云**
> 
> 中国（青岛） cn-qingdao
>
> 中国（北京） cn-beijing
>
> 中国（张家口） cn-zhangjiakou
>
> 中国（呼和浩特） cn-huhehaote
>
> 中国（杭州） cn-hangzhou
>
> 中国（上海） cn-shanghai
>
> 中国（深圳） cn-shenzhen
>
> 中国（成都） cn-chengdu
>
> 中国（广州） cn-guangzhou
>
> 中国（香港） cn-hongkong
>
> 新加坡 ap-southeast-1


目前可用的获取IP的API有 `LanceAPI` `IPIP` `SB` `IPCONF`

> 填写时请注意区分大小写

>IPCONF 全球通用
>
>https://ifconfig.co/ip

>LanceAPI 全球通用
>
>https://api.lance.fun/ip

>IPIP 中国大陆通用
>
>https://myip.ipip.net/ip

>SB 全球通用
>
>https://api-ipv4.ip.sb/ip

#### 运行
使用**命令行**运行

```bash
使用方法: qcip [选项] [<值>]
选项:
    -c  --config <配置文件路径>     指定配置文件路径并运行程序
    -v  --version                 显示版本信息
    -h  --help                    显示帮助信息
    -n  --winnotify               使用Windows通知显示结果
    -ip --ipaddr <IP地址>          直接使用指定的IP地址替换，而不是自动获取 仅支持ipv4
示例:
    qcip # 使用配置文件config.json运行程序
    qcip -c qcipconf.json # 使用配置文件qcipconf.json运行程序
    qcip -ip 1.1.1.1      # 指定使用 ip 1.1.1.1
```

> **注意** 若你使用 **桌面系统** 双击打开程序，会出现命令行窗口和闪退现象，这并不代表运行失败，但是你无法看到运行结果


#### 开机启动(Windows)

你可以在下面的目录(启动项文件夹)中添加该程序的快捷方式或运行程序的批处理文件(bat)
`C:\ProgramData\Microsoft\Windows\Start Menu\Programs\Startup`

> 因为程序依赖于配置文件，请不要直接把程序本体拖入启动项文件夹


#### Windows 通知
以 Windows 通知卡的形式发送运行结果

<img align="right" width="250" src="https://github.com/cnlancehu/qcip/assets/106385654/6d9dc257-581e-49dc-8eb0-ad16d4f19820">

**屏幕截图如右图**

要启用 Windows 通知，你可以使用`-n`运行参数

例如
```bash
qcip -n
qcip -c config.json -n
```

> **注意** `-n`只能在程序主功能中运行，即不可以在`-v`或`-h`中使用

或者在配置文件中，启用以下参数以永久开启 Windows 通知

```json
// config.json
{
    "EnableWinNotify": true
}
```

该功能搭配[开机启动](#开机启动windows)和[隐藏命令行窗口](#隐藏命令行窗口windows)使用效果更佳


#### 隐藏命令行窗口(Windows)

若要隐藏命令行窗口，可以使用以下 vbs 脚本运行程序

```vb
// qcip.vbs
Set WshShell = CreateObject("WScript.Shell")
WshShell.Run "cmd /c [启动 qcip 的命令]", 0, False
```


### 外部链接
程序中调用了以下外部链接
```
获取最新版本号
https://api.lance.fun/proj/qcip/version

配置文件中指定的IP查询API
参见上表
```

### 常见问题
#### 在 Windows 命令提示符中运行时出现乱码
由于 Windows 命令提示符对ANSI编码的支持不完善，因此在 Windows 命令提示符中运行时会出现乱码

解决方法: 使用最新版本的 Windows Terminal 运行

1. 下载 [Windows Terminal](https://apps.microsoft.com/store/detail/windows-terminal/9N0DX20HK701)
2. 使用 Win+I 打开设置，在 隐私和安全性-开发者选项-终端 中选择 Windows 终端

此时命令提示符将由支持ANSI转义的 Windows Terminal 接管，运行程序时能够正确显示颜色

#### 在 Linux 中未找到 GLIBC 库
具体错误信息如下
```
version `GLIBC_2.32' not found
version `GLIBC_2.34' not found
```

在 `1.0.2` 及之前的构建默认启用了 CGO 编译，因此在 Linux 中运行时需要 GLIBC 2.32 及以上的版本

解决方法
- 使用最新的QCIP版本(`1.0.3` 及之后)

或

- 使用最新的 Linux 版本 