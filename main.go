package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	lighthouse "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

var (
	version    = "Dev"
	ua         = "QCIP/" + version
	httpClient = &http.Client{
		Timeout: time.Second * 10,
	}
)

type Config struct {
	MType               string
	SecretId            string
	SecretKey           string
	GetIPAPI            string
	InstanceId          string
	InstanceRegion      string
	SecurityGroupId     string
	SecurityGroupRegion string
	MaxRetries          string
	Rules               []string
}

type IPIPResp struct {
	IP string `json:"ip"`
}

// 主函数
func main() {
	fmt.Printf("QCIP \033[1;32m%s\033[0m\n", version)
	configData := getconfig()
	maxRetries, _ := strconv.Atoi(configData.MaxRetries)
	ip := getip(configData.GetIPAPI, int(maxRetries))
	if configData.MType == "lh" {
		lhmain(configData, ip)
	} else if configData.MType == "cvm" {
		cvmmain(configData, ip)
	}
	os.Exit(0)
}

// 轻量应用服务器主函数
func lhmain(configData Config, ip string) {
	credential := common.NewCredential(
		configData.SecretId,
		configData.SecretKey,
	)
	rules := lhgetrules(credential, configData.InstanceRegion, configData.InstanceId)
	res, needUpdate := lhmatch(rules, ip, configData)
	if needUpdate {
		fmt.Printf("IP is different, start updating\n")
		lhmodifyrules(credential, configData.InstanceRegion, configData.InstanceId, res)
		fmt.Printf("Successfully modified the firewall rules\n")
	} else {
		fmt.Printf("IP is the same\n")
	}
}

// 云服务器主函数
func cvmmain(configData Config, ip string) {
	credential := common.NewCredential(
		configData.SecretId,
		configData.SecretKey,
	)
	rules := sggetrules(credential, configData.SecurityGroupId, configData.SecurityGroupRegion)
	res, needUpdate := sgmatch(rules, ip, configData)
	if needUpdate {
		fmt.Printf("IP is different, start updating\n")
		sgmodifyrules(credential, configData.SecurityGroupId, configData.SecurityGroupRegion, res)
		fmt.Printf("Successfully modified the firewall rules\n")
	} else {
		fmt.Printf("IP is the same\n")
	}
}

// 读取配置文件
func getconfig() Config {
	var confPath string

	if len(os.Args) > 1 {
		confPath = os.Args[1]
	} else {
		confPath = "config.json"
	}
	config, err := os.ReadFile(confPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Config file does not exist")
			os.Exit(1)
		}
		fmt.Printf("Unknown error: %s\n", err)
		os.Exit(1)
	}
	var configData Config
	if !json.Valid(config) {
		fmt.Println("Error: config file is not valid json")
		os.Exit(1)
	}
	err = json.Unmarshal(config, &configData)
	if err != nil {
		decodeErr := json.Unmarshal(config, &configData)
		if decodeErr != nil {
			fmt.Println("Incorrect configuration file format")
			os.Exit(1)
		}
		fmt.Printf("Unknown error: %s\n", err)
		os.Exit(1)
	}
	var requiredKeys []string
	if configData.MType == "lh" {
		requiredKeys = []string{"SecretId", "SecretKey", "InstanceId", "InstanceRegion", "Rules"}
	} else if configData.MType == "cvm" {
		requiredKeys = []string{"SecretId", "SecretKey", "SecurityGroupId", "SecurityGroupRegion", "Rules"}
	} else {
		if configData.MType == "" {
			fmt.Println("Machine type is empty")
			os.Exit(1)
		} else {
			fmt.Printf("Error machine type: %s\n", configData.MType)
			os.Exit(1)
		}
	}
	checkPassing := true
	for _, key := range requiredKeys {
		if _, ok := reflect.TypeOf(configData).FieldByName(key); !ok {
			fmt.Println(key + " not found in config file")
			checkPassing = false
		}
		if reflect.ValueOf(configData).FieldByName(key).String() == "" {
			fmt.Println(key + " is empty")
			checkPassing = false
		}
		if reflect.ValueOf(configData).FieldByName(key).String() == key {
			fmt.Println(key + " is incorrect")
			checkPassing = false
		}
	}
	if !checkPassing {
		os.Exit(1)
	}
	fmt.Printf("Config loaded\n")
	return configData
}

// 获取自身公网IP
func getip(api string, maxretries int) string {
	if api == "LanceAPI" {
		for i := 0; i < maxretries; i++ {
			req, _ := http.NewRequest("GET", "https://api.lance.fun/ip", nil)
			req.Header.Set("User-Agent", ua)
			resp, err := httpClient.Do(req)
			if err != nil || (resp.StatusCode >= 400 && resp.StatusCode <= 599) {
				fmt.Printf("IP API call failed, retrying %d time...\n", i+1)
				time.Sleep(1 * time.Second)
			} else {
				defer resp.Body.Close()
				ip, _ := io.ReadAll(resp.Body)
				return string(ip)
			}
		}
		fmt.Printf("IP API call failed %d times, exiting...\n", maxretries)
		os.Exit(1)
	} else if api == "IPIP" {
		for i := 0; i < maxretries; i++ {
			req, _ := http.NewRequest("GET", "https://myip.ipip.net/ip", nil)
			req.Header.Set("User-Agent", ua)
			resp, err := httpClient.Do(req)
			if err != nil || (resp.StatusCode >= 400 && resp.StatusCode <= 599) {
				fmt.Printf("IP API call failed, retrying %d time...\n", i+1)
			} else {
				defer resp.Body.Close()
				respn, _ := io.ReadAll(resp.Body)
				var r IPIPResp
				err := json.Unmarshal(respn, &r)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				ip := r.IP
				return string(ip)
			}
		}
		fmt.Printf("IP API call failed %d times, exiting...\n", maxretries)
		os.Exit(1)
	} else if api == "SB" {
		for i := 0; i < maxretries; i++ {
			req, _ := http.NewRequest("GET", "https://api-ipv4.ip.sb/ip", nil)
			req.Header.Set("User-Agent", ua)
			resp, err := httpClient.Do(req)
			if err != nil || (resp.StatusCode >= 400 && resp.StatusCode <= 599) {
				fmt.Printf("IP API call failed, retrying %d time...\n", i+1)
				time.Sleep(1 * time.Second)
			} else {
				defer resp.Body.Close()
				ipo, _ := io.ReadAll(resp.Body)
				ip := strings.TrimRight(string(ipo), "\n")
				return ip
			}
		}
		fmt.Printf("IP API call failed %d times, exiting...\n", maxretries)
		os.Exit(1)
	} else {
		fmt.Printf("Error: IP API %s not supported\n", api)
		os.Exit(1)
	}
	return ""
}

// 轻量应用服务器部分
func lhgetrules(credential *common.Credential, InstanceRegion string, InstanceId string) []*lighthouse.FirewallRuleInfo {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "lighthouse.tencentcloudapi.com"
	client, _ := lighthouse.NewClient(credential, InstanceRegion, cpf)
	request := lighthouse.NewDescribeFirewallRulesRequest()
	request.InstanceId = common.StringPtr(InstanceId)
	request.Offset = common.Int64Ptr(0)
	request.Limit = common.Int64Ptr(100)
	response, err := client.DescribeFirewallRules(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s\n", err)
		os.Exit(1)
	}
	if err != nil {
		panic(err)
	}
	return response.Response.FirewallRuleSet
}

func lhmatch(rules []*lighthouse.FirewallRuleInfo, ip string, config Config) ([]*lighthouse.FirewallRuleInfo, bool) {
	needUpdate := false
	for a := range rules {
		for b := range config.Rules {
			if *rules[a].FirewallRuleDescription == config.Rules[b] {
				if *rules[a].CidrBlock == ip {
					continue
				} else {
					*rules[a].CidrBlock = ip
					needUpdate = true
				}
			}
		}
	}
	return rules, needUpdate
}

func lhmodifyrules(credential *common.Credential, InstanceRegion string, InstanceId string, rules []*lighthouse.FirewallRuleInfo) {
	ptrRules := make([]*lighthouse.FirewallRule, len(rules))
	for i := range rules {
		ptrRules[i] = &lighthouse.FirewallRule{
			Protocol:                common.StringPtr(*rules[i].Protocol),
			Port:                    common.StringPtr(*rules[i].Port),
			CidrBlock:               common.StringPtr(*rules[i].CidrBlock),
			Action:                  common.StringPtr(*rules[i].Action),
			FirewallRuleDescription: common.StringPtr(*rules[i].FirewallRuleDescription),
		}
	}
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "lighthouse.tencentcloudapi.com"
	client, _ := lighthouse.NewClient(credential, InstanceRegion, cpf)
	request := lighthouse.NewModifyFirewallRulesRequest()
	request.InstanceId = common.StringPtr(InstanceId)
	request.FirewallRules = ptrRules
	_, err := client.ModifyFirewallRules(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s\n", err)
		os.Exit(1)
		return
	}
	if err != nil {
		panic(err)
	}
}

// 云服务器安全组部分
func sggetrules(credential *common.Credential, SecurityGroupId string, SecurityGroupRegion string) *vpc.SecurityGroupPolicySet {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "vpc.tencentcloudapi.com"
	client, _ := vpc.NewClient(credential, SecurityGroupRegion, cpf)
	request := vpc.NewDescribeSecurityGroupPoliciesRequest()
	request.SecurityGroupId = common.StringPtr(SecurityGroupId)
	response, err := client.DescribeSecurityGroupPolicies(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s", err)
		os.Exit(1)
	}
	if err != nil {
		panic(err)
	}
	return response.Response.SecurityGroupPolicySet
}

func sgmatch(rules *vpc.SecurityGroupPolicySet, ip string, config Config) (*vpc.SecurityGroupPolicySet, bool) {
	needUpdate := false
	for a := range rules.Ingress {
		for b := range config.Rules {
			if *rules.Ingress[a].PolicyDescription == config.Rules[b] {
				if *rules.Ingress[a].CidrBlock == ip {
					continue
				} else {
					rules.Ingress[a].CidrBlock = &ip
					needUpdate = true
				}
			}
		}
	}
	return rules, needUpdate
}

func sgmodifyrules(credential *common.Credential, SecurityGroupId string, SecurityGroupRegion string, rules *vpc.SecurityGroupPolicySet) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "vpc.tencentcloudapi.com"
	client, _ := vpc.NewClient(credential, SecurityGroupRegion, cpf)
	request := vpc.NewModifySecurityGroupPoliciesRequest()
	request.SecurityGroupId = common.StringPtr(SecurityGroupId)
	ptrRules := &vpc.SecurityGroupPolicySet{}
	ptrRules.Version = rules.Version
	ptrRules.Egress = nil
	ptrRules.Ingress = rules.Ingress
	for a := range ptrRules.Ingress {
		ptrRules.Ingress[a].SecurityGroupId = &SecurityGroupId
		ptrRules.Ingress[a].PolicyIndex = nil
		ptrRules.Ingress[a].Ipv6CidrBlock = nil
		ptrRules.Ingress[a].SecurityGroupId = nil
		ptrRules.Ingress[a].AddressTemplate = nil
	}
	rulesfin := processRules(ptrRules).(*vpc.SecurityGroupPolicySet)
	request.SecurityGroupPolicySet = rulesfin
	_, err := client.ModifySecurityGroupPolicies(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s", err)
		os.Exit(1)
	}
	if err != nil {
		panic(err)
	}
}

// 把ptrRules中内容为空的值替换为nil
func processRules(ptrRules interface{}) interface{} {
	v := reflect.ValueOf(ptrRules)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.Ptr && !field.IsNil() && field.Elem().Kind() == reflect.String && field.Elem().String() == "" {
			field.Set(reflect.Zero(field.Type()))
		} else if field.Kind() == reflect.Struct {
			processRules(field.Addr().Interface())
		} else if field.Kind() == reflect.Ptr && !field.IsNil() && field.Elem().Kind() == reflect.Struct {
			processRules(field.Interface())
		} else if field.Kind() == reflect.Slice && field.Type().Elem().Kind() == reflect.Ptr && field.Type().Elem().Elem().Kind() == reflect.Struct {
			for j := 0; j < field.Len(); j++ {
				elem := field.Index(j)
				processRules(elem.Interface())
			}
		}
	}
	return ptrRules
}
