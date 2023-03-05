# qcliteautorip
自动设置腾讯云轻量服务器的防火墙来源限制，使服务器重要端口只能被开发者的ip访问

> **注意** 该脚本仅支持控制腾讯云轻量应用服务器的防火墙
### 仍在施工中

## 原理
通过腾讯云的API以获取服务器的安全组信息，将服务器指定防火墙策略的**来源**ip限制为运行该程序的设备的公网ip，使服务器的重要端口只能被开您的ip访问。

## 食用教程
### 安装
#### 方法一 直接下载可执行文件(推荐)
从[Releases](https://github.com/cnlancehu/qcliteautorip/releases)中下载适合你的系统的可执行文件，解压

若没有适合你的系统的二进制文件，请使用[方法二](#方法二-使用python运行)或[自行构建](#自行构建)
#### 方法二 使用Python运行
**环境要求**
> Python 3.6+ (建议使用 3.10)

必要的Python Module:

```bash
# 腾讯云SDK
pip3 install tencentcloud-sdk-python
# Requests
pip3 install requests 
```

### 更改配置文件

请事先在[腾讯云访问管理](https://console.cloud.tencent.com/cam/capi)创建API密钥

```json
// config.json
// 请注意填写时的大小写规范
{
    "SecretId": "SecretId", // SecretId
    "SecretKey": "SecretKey", // SecretKey
    "GetIPAPI": "IPIP", // 获取IP的API，选填 LanceAPI 或 IPIP ，默认为IPIP， 中国大陆用户请使用 IPIP
    "InstanceId": "InstanceId", // 服务器的实例ID
    "InstanceRegion": "ap-hongkong", // 服务器的地域，参见下文附录
    "Rules": [
        // 第一个策略
        {
            "FirewallRuleDescription": "http" // 填入你要修改来源的防火墙策略的描述
        },
        // 第二个策略，如此类推，可填写多个
        {
            "FirewallRuleDescription": "ssh" 
        }
    ]
}
```
如下图，把下图划线处的内容(即防火墙策略的备注)填入配置文件的FirewallRuleDescription对应的内容中，那么这条防火墙规则将会添加到自动更新的列表中
![image](https://user-images.githubusercontent.com/106385654/214570514-90e46714-c3a3-450f-ba37-36f8dcb9089a.png)
即
```json
// config.json
{
    "Rules": [
        {
            "FirewallRuleDescription": "ssh"
        }
    ]
}
```

### 运行脚本
若你使用的是**方法一**，现在你可以直接运行可执行文件了

若你采用的是**方法二**，请使用以下脚本运行
```bash
python3 main.py
```
这样，脚本就可以自动获取你的ip，并将指定防火墙策略的来源限制为你的ip

但这只是一次性的，如果你的ip发生变化，你需要再次运行脚本

### 开机启动
另外，你也可以把程序文件放入电脑的启动项中，这样，每次开机时，脚本就会自动运行

你可以在下面的目录(启动项文件夹)中添加该脚本的**快捷方式**

> **注意** 不要在启动文件夹中直接放入程序文件，因为需要配置文件配合其运行

`C:\ProgramData\Microsoft\Windows\Start Menu\Programs\Startup`


```vbs
// qclarip.vbs
// 该脚本适用于方法二，如果你使用的是方法一，请适当修改后使用
Set WshShell = CreateObject("WScript.Shell")
WshShell.Run "cmd /c python3 /*程序的地址*/", 0, False
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
https://get.lance.fun/ip/

IPIP // 推荐在中国大陆使用
http://myip.ipip.net/ip
```

### 自行构建

#### 环境要求
> Python 3.10

必要的Python Module

```bash
pip3 install pyinstaller
pip3 install requests
pip3 install tencentcloud-sdk-python
pip3 install pillow
```

#### 构建

克隆本仓库到本地
```bash
git clone https://github.com/cnlancehu/qcliteautorip.git
cd qcliteautorip
```

构建
```bash
python3 -m PyInstaller main.py -F -i icon.ico
mv dist/main.exe dist/qcip.exe
mv config.json dist/config.json
```

二进制文件和配置文件将会生成在 `dist` 文件夹中