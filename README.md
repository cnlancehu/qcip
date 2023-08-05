# qcliteautorip 

### 状态

![状态](https://api.lance.fun/msg/qcip)


腾讯云轻量服务器防火墙来源自动设置脚本

自动设置腾讯云轻量服务器的防火墙来源限制，使服务器重要端口只能被开发者的ip访问

> **注意** 该脚本仅支持控制腾讯云**轻量应用服务器**的防火墙

**仍在迭代中**

## 原理
从配置文件中读取SecretId、SecretKey、InstanceId等信息以访问腾讯云API，然后获取本机的IP地址并检查防火墙规则是否需要更新。如果IP发生更改，它会修改适当的防火墙规则，以确保实例的指定端口只允许本机地址访问。 

## 食用教程
### 安装
#### 方法一 直接下载可执行文件(推荐)
从[Releases](https://github.com/cnlancehu/qcliteautorip/releases)中下载适合你的系统的可执行文件，解压

若没有适合你的系统的二进制文件，请使用[方法二](#方法二-使用python运行)
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
    "SecretId": "SecretId", // 腾讯云API密钥ID
    "SecretKey": "SecretKey", // 腾讯云API密钥Key
    "GetIPAPI": "IPIP", // 获取IP的API，选填 LanceAPI 或 IPIP ，默认为IPIP， 中国大陆用户请使用 IPIP
    "InstanceId": "InstanceId", // 服务器的实例ID
    "InstanceRegion": "InstanceRegion", // 服务器的地域，参见下文附录
    "MaxRetries": "3", // 获取IP地址时出现错误的最大重试次数
    "Rules": ["%第一个防火墙配置的备注%", "%第二个%"] // 需要修改的防火墙策略的备注，可填写多个
}
```
如下图，把下图划线处的内容(即防火墙策略的备注)填入配置文件的 Rules 对应的内容中，那么这条防火墙规则将会添加到自动更新的列表中
![image](https://user-images.githubusercontent.com/106385654/214570514-90e46714-c3a3-450f-ba37-36f8dcb9089a.png)
即
```json
// config.json
{
    "Rules": ["ssh"]
    // 在此处以列表形式填入防火墙策略的备注
}
```

### 运行脚本
若你使用的是**方法一**，现在你可以直接运行可执行文件了
> **注意** 若你使用 桌面系统 直接双击打开脚本，会出现cmd窗口和闪退现象，这是正常的，但你无法看到程序的运行结果

你也可以使用命令行运行以查看运行结果

```bash
qcip.exe config.json # Windows

qcip config.json # Linux & MacOS

{程序路径} {配置文件路径(可选，默认为 config.json)}
```


若你采用的是**方法二**，请使用以下脚本运行
```bash
python3 main.py config.json

python3 {python文件路径} {配置文件路径(可选，默认为 config.json)}

```
这样，脚本就可以自动获取你的ip，并将指定防火墙策略的来源限制为你的ip

但这只是一次性的，如果你的ip发生变化，你需要再次运行脚本

### 开机启动
你也可以把程序文件放入电脑的启动项中，这样，每次开机时，脚本就会自动运行

你可以在下面的目录(启动项文件夹)中添加该脚本的**快捷方式**

`C:\ProgramData\Microsoft\Windows\Start Menu\Programs\Startup`


```vbs
// qcip.vbs
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
https://api.lance.fun/ip/

IPIP // 推荐在中国大陆使用
https://myip.ipip.net/ip
```


#### Q & A

1. > Q: 为什么不能让用户自己指定获取IP的API地址?
   >
   > A: 由于每个API的返回格式不同，所以需要对每个API进行适配，你可以适当修改代码以适配新的API，或者提交PR来让我们支持更多的API

2. > Q: 我能不能自行编译二进制文件?
   >
   > A: 当然可以，你可以使用 `pyinstaller` 来编译，但是我们预购建的二进制文件使用 `Nuitka` 构建，更安全，更效率，但由于`Nuitka`极其依赖使用环境，我们推荐您自己使用 `pyinstaller` 编译，我们将在以后推出使用 `Nuitka` 构建的教程
