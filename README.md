# qcliteautorip
自动设置腾讯云轻量服务器的防火墙来源限制，使服务器重要端口只能被开发者的ip访问

> 注意 该脚本支持控制腾讯云轻量应用服务器的防火墙
### 仍在施工中

## 原理
通过腾讯云的API以获取服务器的安全组信息，将服务器安全组的**来源**设置为你的IP，使服务器的重要端口只能被开发者的ip访问。

## 食用教程
### 环境要求
> Python 3.6+ (仅在Python3.9 3.10上测试过)

必要的Python Module:

TencentSDK
```bash
pip install tencentcloud-sdk-python
```

Requests
```bash
pip install requests
```

### 更改配置文件

请事先在[腾讯云访问管理](https://console.cloud.tencent.com/cam/capi)创建API密钥

```json
// config.json
{
    "SecretId": "SecretId", //对应上文创建API密钥中的 SecretId
    "SecretKey": "SecretKey", //对应上文创建API密钥中的 SecretKey
    "GetIPAPI": "LanceAPI", //获取IP的API，目前仅支持LanceAPI (可不填)
    "InstanceId": "InstanceId", //服务器的实例ID
    "InstanceRegion": "ap-hongkong", //服务器的地域，可在附录参见
    "Rules": [
        {
            "FirewallRuleDescription": "http" //填入你要修改来源的防火墙策略的描述 
        },
        {
            "FirewallRuleDescription": "ssh"
        }
    ]
}
```
如下图，把下图划线处的内容填入配置文件的FirewallRuleDescription对应的内容中
![image](https://user-images.githubusercontent.com/106385654/214570514-90e46714-c3a3-450f-ba37-36f8dcb9089a.png)
即
```json
"Rules": [
    {
        "FirewallRuleDescription": "ssh"
    }
]
```

### 运行脚本
```bash
python main.py
```
这样，脚本就可以自动获取你的ip，并将服务器的重要端口的来源限制为你的ip

但这只是一次性的，如果你的ip发生变化，你需要再次运行脚本

### 开机启动
另外，你也可以把程序文件放入电脑的启动项中，这样，每次开机时，脚本就会自动运行

但请注意不要把程序文件拖入启动项文件夹中，因为脚本需要config.json才能运作

你可以在下面的目录(启动项文件夹)中添加该脚本

`C:\ProgramData\Microsoft\Windows\Start Menu\Programs\Startup`


```vbs
Set WshShell = CreateObject("WScript.Shell")
WshShell.Run "cmd /c python /*程序的地址*/", 0, False
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