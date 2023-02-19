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
    json.load(open(configpath, 'r', encoding='utf-8'))
except FileNotFoundError:
    print('Config file not found')
    exit()
except FileExistsError:
    print('Config file exists error')
    exit()
except json.decoder.JSONDecodeError:
    print('Incorrect configuration file format')
    exit()
except UnicodeDecodeError:
    print('Incorrect configuration file, it should be a text file using utf-8 encoding')
    exit()
except Exception as e:
    print('Unknown error: ' + str(e))
    exit()
with open(configpath, 'r', encoding='utf-8') as f:
    config = json.load(f)
    f.close()
# Get IP
if 'GetIPAPI' in config:
    if config['GetIPAPI'] == "LanceAPI":
        try:
            ip = requests.get('https://get.lance.fun/ip/').text
        except Exception as e:
            print('This api may not work anymore, please replace it and try again')
            print('Detail: ' + str(e))
            exit()
    elif config['GetIPAPI'] == "IPIP":
        try:
            ip = json.loads(requests.get('https://myip.ipip.net/ip').text)['ip']
        except Exception as e:
            print('This api may not work anymore, please replace it and try again')
            print('Detail: ' + str(e))
            exit()
else:
    try:
        ip = requests.get('https://get.lance.fun/ip/').text
    except Exception as e:
        print('This api may not work anymore, please replace it and try again')
        print('Detail: ' + str(e))
        exit()
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
if 'InstanceId' not in config:
    print('InstanceID not found')
    checkpassing = False
# Check Region
if 'InstanceRegion' not in config:
    print('InstanceRegion not found')
    checkpassing = False
if 'Rules' not in config:
    print('Rules not found')
    checkpassing = False
if checkpassing == False:
    exit()
InstanceId = config['InstanceId']
InstanceRegion = config['InstanceRegion']
cred = credential.Credential(config['SecretId'], config['SecretKey'])
rules = json.dumps(config['Rules'])
print('Config load successfully')


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
except Exception as e:
    print('Unknown error: ' + str(e))
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
    except Exception as e:
        print('Unknown error: ' + str(e))
        exit()
elif needupdate == False:
    print("IP相同 无需更新")
