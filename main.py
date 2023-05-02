import requests
import json
from sys import exit
from tencentcloud.common import credential
from tencentcloud.common.profile.client_profile import ClientProfile
from tencentcloud.common.profile.http_profile import HttpProfile
from tencentcloud.common.exception.tencent_cloud_sdk_exception import TencentCloudSDKException
from tencentcloud.lighthouse.v20200324 import lighthouse_client, models


def get_config(configpath):
    """
    Get configuration from the specified file path.

    Args:
        configpath (str): The path of the configuration file.

    Returns:
        dict: The configuration dictionary.

    Raises:
        FileNotFoundError: If the configuration file is not found.
        FileExistsError: If the configuration file exists error.
        json.decoder.JSONDecodeError: If the configuration file format is incorrect.
        UnicodeDecodeError: If the configuration file is not a text file using utf-8 encoding.
        Exception: If there is an unknown error.
    """
    try:
        with open(configpath, 'r', encoding='utf-8') as f:
            config = json.load(f)
        return config
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


def get_ip(config):
    """
    Get the IP address of the current machine.

    Args:
        config (dict): The configuration dictionary.

    Returns:
        str: The IP address of the current machine.

    Raises:
        Exception: If there is an unknown error.
    """
    if 'GetIPAPI' in config:
        if config['GetIPAPI'] == "LanceAPI":
            try:
                ip = requests.get('https://api.lance.fun/ip/').text
            except Exception as e:
                print('This api may not work anymore, please use another and try again')
                print('Detail: ' + str(e))
                exit()
        elif config['GetIPAPI'] == "IPIP":
            try:
                ip = json.loads(requests.get('https://myip.ipip.net/ip').text)['ip']
            except Exception as e:
                print('This api may not work anymore, please use another and try again')
                print('Detail: ' + str(e))
                exit()
    else:
        try:
            ip = json.loads(requests.get('https://myip.ipip.net/ip').text)['ip']
        except Exception as e:
            print('This api may not work anymore, please use another and try again')
            print('Detail: ' + str(e))
            exit()
    return ip


def check_config(config):
    """
    Check if the configuration is valid.

    Args:
        config (dict): The configuration dictionary.

    Returns:
        bool: True if the configuration is valid, False otherwise.
    """
    checkpassing = True
    if 'SecretId' not in config and 'SecretKey' not in config:
        print('Both SecretId and SecretKey not found')
        checkpassing = False
    elif 'SecretId' not in config:
        print('SecretId not found')
        checkpassing = False
    elif 'SecretKey' not in config:
        print('SecretKey not found')
        checkpassing = False
    if 'InstanceId' not in config:
        print('InstanceID not found')
        checkpassing = False
    if 'InstanceRegion' not in config:
        print('InstanceRegion not found')
        checkpassing = False
    if 'Rules' not in config:
        print('Rules not found')
        checkpassing = False
    if not checkpassing:
        exit()
    return checkpassing


def get_firewall_rules(cred, InstanceRegion, InstanceId):
    """
    Get the firewall rules of the specified instance.

    Args:
        cred (credential.Credential): The credential object.
        InstanceRegion (str): The region of the instance.
        InstanceId (str): The ID of the instance.

    Returns:
        list: The firewall rules of the specified instance.

    Raises:
        TencentCloudSDKException: If there is an error in the Tencent Cloud SDK.
        Exception: If there is an unknown error.
    """
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
        return resp
    except TencentCloudSDKException as err:
        print(err)
        exit()
    except Exception as e:
        print('Unknown error: ' + str(e))
        exit()


def modify_firewall_rules(cred, InstanceRegion, InstanceId, resp):
    """
    Modify the firewall rules of the specified instance.

    Args:
        cred (credential.Credential): The credential object.
        InstanceRegion (str): The region of the instance.
        InstanceId (str): The ID of the instance.
        resp (list): The new firewall rules.

    Raises:
        TencentCloudSDKException: If there is an error in the Tencent Cloud SDK.
        Exception: If there is an unknown error.
    """
    try:
        httpProfile = HttpProfile()
        httpProfile.endpoint = "lighthouse.tencentcloudapi.com"
        clientProfile = ClientProfile()
        clientProfile.httpProfile = httpProfile
        client = lighthouse_client.LighthouseClient(
            cred, InstanceRegion, clientProfile)
        req = models.ModifyFirewallRulesRequest()
        params = json.dumps({
            "InstanceId": InstanceId,
            "FirewallRules": "newrule"
        })
        params = params.replace('"newrule"', str(resp))

        req.from_json_string(params)
        resp = client.ModifyFirewallRules(req)
        print("Successfully modified the firewall rules")

    except TencentCloudSDKException as err:
        print(err)
        exit()
    except Exception as e:
        print('Unknown error: ' + str(e))
        exit()


def main():
    """
    The main function.
    """
    # Init
    needupdate = False

    # Get config
    config = get_config('config.json')
    check_config(config)
    InstanceId = config['InstanceId']
    InstanceRegion = config['InstanceRegion']
    cred = credential.Credential(config['SecretId'], config['SecretKey'])
    json.dumps(config['Rules'])
    print('Config load successfully')

    # Get IP
    ip = get_ip(config)

    # Get Firewall Rules
    resp = get_firewall_rules(cred, InstanceRegion, InstanceId)

    # Modify Firewall Rules
    for a in range(len(resp)):
        for b in range(len(config['Rules'])):
            if resp[a].FirewallRuleDescription == config['Rules'][b]['FirewallRuleDescription']:
                if resp[a].CidrBlock == ip:
                    pass
                else:
                    resp[a].CidrBlock = ip
                    needupdate = True
    if needupdate:
        print("IP is different, start updating")
        modify_firewall_rules(cred, InstanceRegion, InstanceId, resp)
    else:
        print("IP is the same, no need to update")


if __name__ == '__main__':
    main()
