import json
from requests import get as reqget
from sys import exit, argv
from time import sleep
from termcolor import colored
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
        json.load(open(configpath, 'r', encoding='utf-8'))
    except FileNotFoundError:
        print(colored('Config file not found', 'light_red'))
        exit()
    except FileExistsError:
        print(colored('Config file exists error'), 'light_red')
        exit()
    except json.decoder.JSONDecodeError:
        print(colored('Incorrect configuration file format', 'light_red'))
        exit()
    except UnicodeDecodeError:
        print(colored('Incorrect configuration file, it should be a text file using utf-8 encoding', 'light_red'))
        exit()
    except Exception as e:
        e = 'Unknown error: ' + str(e)
        print(colored(e, 'light_red'))
        exit()
    else:
        config = json.load(open(configpath, 'r', encoding='utf-8'))
    return config


def get_ip(api, maxretries=3):
    """
    Get the IP address of the current machine.

    Args:
        api (str): The API used to get the IP address.
        maxretries (int): The maximum number of retries.

    Returns:
        str: The IP address of the current machine.

    Raises:
        Exception: If there is an unknown error.
    """
    for i in range(maxretries):
        try:
            if api == "LanceAPI":
                ip = reqget('https://api.lance.fun/ip/').text
            elif api == "IPIP":
                ip = json.loads(reqget(
                    'https://myip.ipip.net/ip').text)['ip']
            else:
                ip = json.loads(reqget(
                    'https://myip.ipip.net/ip').text)['ip']
        except Exception as e:
            print(f'IP API call failed, retrying {i+1} time...')
            if i == maxretries - 1:
                print(colored('Max retries exceeded, try to change an api', 'light_red'))
                exit()
            sleep(0.5)
            continue
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
        print(colored('Both SecretId and SecretKey not found', 'light_red')) 
        checkpassing = False 
    elif 'SecretId' not in config: 
        print(colored('SecretId not found', 'light_red')) 
        checkpassing = False 
    elif 'SecretKey' not in config: 
        print(colored('SecretKey not found', 'light_red')) 
        checkpassing = False 
    if 'InstanceId' not in config: 
        print(colored('InstanceID not found', 'light_red')) 
        checkpassing = False 
    if 'InstanceRegion' not in config: 
        print(colored('InstanceRegion not found', 'light_red')) 
        checkpassing = False 
    if 'Rules' not in config: 
        print(colored('Rules not found', 'light_red')) 
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
        print(colored(err, 'light_red'))
        exit()
    except Exception as e:
        e = 'Unknown error: ' + str(e)
        print(colored(e, 'light_red'))
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
        print(colored(err, 'light_red'))
        exit()
    except Exception as e:
        e = 'Unknown error: ' + str(e)
        print(colored(e, 'light_red'))
        exit()


def main():
    """
    The main function.
    """
    try:
        try:
            version
        except:
            print('QCIP ' + colored('Python version ', 'light_yellow'))
        else:
            print('QCIP ' + version)
        # Init
        needupdate = False

        # Get config
        try:
            argv[1]
        except IndexError:
            config = get_config('config.json')
        else:
            config = get_config(argv[1])
        check_config(config)
        try:
            maxretries = int(config['MaxRetries'])
        except KeyError:
            maxretries = 3
        InstanceId = config['InstanceId']
        InstanceRegion = config['InstanceRegion']
        cred = credential.Credential(config['SecretId'], config['SecretKey'])
        json.dumps(config['Rules'])
        print('Config load successfully')

        # Get IP
        try:
            config['GetIPAPI']
        except KeyError:
            ip = get_ip('IPIP', maxretries)
        else:
            ip = get_ip(config['GetIPAPI'], maxretries)

        # Get Firewall Rules
        resp = get_firewall_rules(cred, InstanceRegion, InstanceId)

        # Modify Firewall Rules
        for a in range(len(resp)):
            for b in range(len(config['Rules'])):
                if resp[a].FirewallRuleDescription == config['Rules'][b]:
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
        exit()
    except KeyboardInterrupt:
        print(colored('KeyboardInterrupt', 'light_red'))
        exit()
    except Exception as err:
        err = 'Error: ' + err
        print(colored(err, 'light_red'))
        exit()


if __name__ == '__main__':
    main()
