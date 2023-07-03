# qcip

**全新Golang重写版 比Python版运行速度快了11.3753秒**


自动设置腾讯云轻量服务器的防火墙来源限制，使服务器重要端口只能被运行者的ip访问

> **注意** 该脚本仅支持控制腾讯云**轻量应用服务器**的防火墙

## 原理
从配置文件中读取SecretId、SecretKey、InstanceId等信息以访问腾讯云API，获取本机的IP地址并检查防火墙规则是否需要更新。若IP发生改变，修改防火墙规则的**来源限制**，以确保实例的指定端口只允许运行者的IP访问。 

## 使用方法

### 下载程序
到Release页面下载你的系统对应的程序

### 更改配置文件

请事先在[腾讯云访问管理](https://console.cloud.tencent.com/cam/capi)创建API密钥

```jsonc
// config.json
// 请注意填写时的大小写规范
{
    "SecretId": "SecretId", // 腾讯云API密钥ID
    "SecretKey": "SecretKey", // 腾讯云API密钥Key
    "GetIPAPI": "IPIP", // 获取IP的API，选填 LanceAPI IPIP SB
    "InstanceId": "InstanceId", // 服务器的实例ID
    "InstanceRegion": "InstanceRegion", // 服务器的地域，参见下文附录
    "MaxRetries": "3", // 获取IP地址时出现错误的最大重试次数
    "Rules": ["%第一个防火墙配置的备注%", "%第二个%"] // 需要修改的防火墙策略的备注，可填写多个
}
```
如下图，把下图划线处的内容(即防火墙策略的备注)填入配置文件的 Rules 对应的内容中，那么这条防火墙规则将会添加到自动更新的列表中
![image](https://user-images.githubusercontent.com/106385654/214570514-90e46714-c3a3-450f-ba37-36f8dcb9089a.png)
即
```jsonc
// config.json
{
    "Rules": ["ssh"]
    // 在此处以列表形式填入防火墙策略的备注
}
```

### 运行

请使用命令行运行

```bash
qcip {config.json的路径(可选，默认为与程序相同目录下的config.json)}
```

### 开机启动
你也可以把程序文件放入电脑的启动项中，这样，每次开机时，脚本就会自动运行

你可以在下面的目录(启动项文件夹)中添加该脚本的**快捷方式**或**运行程序的批处理文件(bat)**

`C:\ProgramData\Microsoft\Windows\Start Menu\Programs\Startup`

使用以下脚本运行以隐藏命令行窗口
```vbs
// qcip.vbs
Set WshShell = CreateObject("WScript.Shell")
WshShell.Run "cmd /c /*运行的命令*/", 0, False
```

## 附录

地域参照表
```
华北地区(北京) ap-beijing
西南地区(成都) ap-chengdu
华南地区(广州) ap-guangzhou
港澳台地区(中国香港) ap-hongkong
亚太地区(首尔) ap-seoul
华东地区(上海) ap-shanghai
东南亚地区(新加坡) ap-singapore
```

各API地址
```
LanceAPI // 推荐在海外使用
https://api.lance.fun/ip/

IPIP // 推荐在中国大陆使用
https://myip.ipip.net/ip

SB // 全球通用 但效率较慢
https://api-ipv4.ip.sb/ip
```

> 更多IP正在适配中