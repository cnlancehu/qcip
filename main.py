import requests
import json
from sys import exit
from tencentcloud.common import credential
from tencentcloud.common.profile.client_profile import ClientProfile
from tencentcloud.common.profile.http_profile import HttpProfile
from tencentcloud.common.exception.tencent_cloud_sdk_exception import TencentCloudSDKException
from tencentcloud.lighthouse.v20200324 import lighthouse_client, models

# Init
needupdate = False
checkpassing = True

# Get config
try:
    configpath
except NameError:
    configpath = 'config.json' 
try:
    open(configpath, 'r')
except FileNotFoundError:
    print('Config file not found')
    exit()
except FileExistsError:
    print('Config file exists error')
    exit()
with open(configpath, 'r') as f:
    config = json.load(f)
    f.close()
# Get IP
if 'GetIPAPI' in config:
    if config['GetIPAPI'] == "LanceAPI":
        ip = requests.get('https://get.lance.fun/ip/').text
    elif config['GetIPAPI'] == "IPIP":
        ip = requests.get('http://myip.ipip.net/').text.split(' ')[1][3:]
else:
    ip = requests.get('https://get.lance.fun/ip/').text
# Check Secret
if 'SecretId' not in config and 'SecretKey' not in config:
    print('Both SecretId and SecretKey not found')
    checkpassing = False
elif 'SecretId' not in config:
    print('SecretId not found')
    checkpassing = False
elif 'SecretKey' not in config:
    print('SecretKey not found')
    checkpassing = False
# Check InstanceID
elif 'InstanceId' not in config:
    print('InstanceID not found')
    checkpassing = False
# Check Region
elif 'InstanceRegion' not in config:
    print('InstanceRegion not found')
    checkpassing = False
elif 'Rules' not in config:
    print('Rules not found')
    checkpassing = False
else:
    InstanceId = config['InstanceId']
    InstanceRegion = config['InstanceRegion']
    cred = credential.Credential(config['SecretId'], config['SecretKey'])
    rules = json.dumps(config['Rules'])
    print('Config load successfully')
if checkpassing == False:
    exit()


# Get Firewall Rules
try:
    httpProfile = HttpProfile()
    httpProfile.endpoint = "lighthouse.tencentcloudapi.com"
    clientProfile = ClientProfile()
    clientProfile.httpProfile = httpProfile
    client = lighthouse_client.LighthouseClient(
        cred, InstanceRegion, clientProfile)
    req = models.DescribeFirewallRulesRequest()
    params = {
        "InstanceId": InstanceId,
        "Offset": 0,
        "Limit": 100
    }
    req.from_json_string(json.dumps(params))
    resp = client.DescribeFirewallRules(req).FirewallRuleSet

except TencentCloudSDKException as err:
    print(err)
    exit()


# Modify Firewall Rules
for a in range(0, len(resp)):
    for b in range(0, len(config['Rules'])):
        if resp[a].FirewallRuleDescription == config['Rules'][b]['FirewallRuleDescription']:
            if resp[a].CidrBlock == ip:
                pass
            else:
                resp[a].CidrBlock = ip
                needupdate = True
if needupdate == True:
    print("IP不同 开始更新")
    try:
        httpProfile = HttpProfile()
        httpProfile.endpoint = "lighthouse.tencentcloudapi.com"
        clientProfile = ClientProfile()
        clientProfile.httpProfile = httpProfile
        client = lighthouse_client.LighthouseClient(
            cred, "ap-hongkong", clientProfile)
        req = models.ModifyFirewallRulesRequest()
        params = json.dumps({
            "InstanceId": InstanceId,
            "FirewallRules": "newrule"
        })
        params = params.replace('"newrule"', str(resp))

        req.from_json_string(params)
        resp = client.ModifyFirewallRules(req)
        print("成功")

    except TencentCloudSDKException as err:
        print(err)
        exit()
elif needupdate == False:
    print("IP相同 无需更新")
