package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
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
	ua         = "qcip/" + version
	httpClient = &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("tcp4", addr)
			},
		},
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
	fmt.Printf("QCIP \033[1;32mv%s\033[0m\n", version)
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
			errhandle("Config error: config file " + confPath + " does not exist")
			os.Exit(1)
		}
		errhandle("Config error: " + err.Error())
		os.Exit(1)
	}
	var configData Config
	if !json.Valid(config) {
		errhandle("Config error: config file is not valid json")
		os.Exit(1)
	}
	err = json.Unmarshal(config, &configData)
	if err != nil {
		decodeErr := json.Unmarshal(config, &configData)
		if decodeErr != nil {
			errhandle("Config error: config file format is incorrect")
			os.Exit(1)
		}
		errhandle("Config error: " + err.Error())
		os.Exit(1)
	}
	var requiredKeys []string
	if configData.MType == "lh" {
		requiredKeys = []string{"SecretId", "SecretKey", "InstanceId", "InstanceRegion", "Rules"}
	} else if configData.MType == "cvm" {
		requiredKeys = []string{"SecretId", "SecretKey", "SecurityGroupId", "SecurityGroupRegion", "Rules"}
	} else {
		if configData.MType == "" {
			errhandle("Config error: machine type is empty")
			os.Exit(1)
		} else {
			errhandle("Config error: machine type " + configData.MType + " is incorrect")
			os.Exit(1)
		}
	}
	checkPassing := true
	for _, key := range requiredKeys {
		if _, ok := reflect.TypeOf(configData).FieldByName(key); !ok {
			checkPassing = false
		}
		if reflect.ValueOf(configData).FieldByName(key).String() == "" {
			checkPassing = false
		}
		if reflect.ValueOf(configData).FieldByName(key).String() == key {
			checkPassing = false
		}
	}
	if !checkPassing {
		errhandle("Config error:")
		for _, key := range requiredKeys {
			if _, ok := reflect.TypeOf(configData).FieldByName(key); !ok {
				errhandle("	" + key + " not found")
			} else if reflect.ValueOf(configData).FieldByName(key).String() == "" {
				errhandle("	" + key + " is empty")
			} else if reflect.ValueOf(configData).FieldByName(key).String() == key {
				errhandle("	" + key + " is incorrect")
			}
		}
		os.Exit(1)
	}
	fmt.Printf("Config loaded\n")
	return configData
}

// 获取自身公网IP
func getip(api string, maxretries int) string {
	if maxretries < 0 {
		errhandle("Config error: maxretries should be an integer greater than or equal to 0")
		os.Exit(1)
	}
	if api == "LanceAPI" {
		if maxretries == 0 {
			req, _ := http.NewRequest("GET", "https://api.lance.fun/ip", nil)
			req.Header.Set("User-Agent", ua)
			resp, err := httpClient.Do(req)
			if err != nil || (resp.StatusCode >= 400 && resp.StatusCode <= 599) {
				errhandle("IP API calling error")
				errhandle("	Error detail: " + err.Error())
				os.Exit(1)
			}
			defer resp.Body.Close()
			ip, _ := io.ReadAll(resp.Body)
			return string(ip)
		}

		for i := 0; i <= maxretries; i++ {
			req, _ := http.NewRequest("GET", "https://api.lance.fun/ip", nil)
			req.Header.Set("User-Agent", ua)
			resp, err := httpClient.Do(req)
			if err != nil || (resp.StatusCode >= 400 && resp.StatusCode <= 599) {
				if i == 0 {
					errhandle("IP API calling error:")
					errhandle("	Error detail: " + err.Error())
				}
				if i == 0 && maxretries == 1 {
					fmt.Printf("\r\033[31m%s\033[0m\n", "	retrying "+strconv.Itoa(i)+"/"+strconv.Itoa(maxretries)+" time")
				} else if i == 1 {
					fmt.Printf("\r\033[31m%s\033[0m", "	retrying "+strconv.Itoa(i)+"/"+strconv.Itoa(maxretries)+" time")
				} else if i > 1 && maxretries > 1 {
					fmt.Printf("\r\033[31m%s\033[0m", "	retrying "+strconv.Itoa(i)+"/"+strconv.Itoa(maxretries)+" times")
				} else if i == maxretries {
					fmt.Printf("\r\033[31m%s\033[0m\n", "	retrying "+strconv.Itoa(i)+"/"+strconv.Itoa(maxretries)+" times")
				}
				time.Sleep(1 * time.Second)
			} else {
				defer resp.Body.Close()
				ip, _ := io.ReadAll(resp.Body)
				return string(ip)
			}
		}
		errhandle("\nIP API call failed " + fmt.Sprint(maxretries) + " times, exiting...")
		os.Exit(1)
	} else if api == "IPIP" {
		if maxretries == 0 {
			req, _ := http.NewRequest("GET", "https://myip.ipip.net/ip", nil)
			req.Header.Set("User-Agent", ua)
			resp, err := httpClient.Do(req)
			if err != nil || (resp.StatusCode >= 400 && resp.StatusCode <= 599) {
				errhandle("IP API calling error")
				errhandle("	Error detail: " + err.Error())
				os.Exit(1)
			}
			defer resp.Body.Close()
			respn, _ := io.ReadAll(resp.Body)
			var r IPIPResp
			err = json.Unmarshal(respn, &r)
			if err != nil {
				errhandle("IP API calling error: " + err.Error())
				errhandle("	Error detail: " + err.Error())
				os.Exit(1)
			}
			ip := r.IP
			return string(ip)
		}

		for i := 0; i <= maxretries; i++ {
			req, _ := http.NewRequest("GET", "https://myip.ipip.net/ip", nil)
			req.Header.Set("User-Agent", ua)
			resp, err := httpClient.Do(req)
			if err != nil || (resp.StatusCode >= 400 && resp.StatusCode <= 599) {
				if i == 0 {
					errhandle("IP API calling error:")
					errhandle("	Error detail: " + err.Error())
				}
				if i == 0 && maxretries == 1 {
					fmt.Printf("\r\033[31m%s\033[0m\n", "	retrying "+strconv.Itoa(i)+"/"+strconv.Itoa(maxretries)+" time")
				} else if i == 1 {
					fmt.Printf("\r\033[31m%s\033[0m", "	retrying "+strconv.Itoa(i)+"/"+strconv.Itoa(maxretries)+" time")
				} else if i > 1 && maxretries > 1 {
					fmt.Printf("\r\033[31m%s\033[0m", "	retrying "+strconv.Itoa(i)+"/"+strconv.Itoa(maxretries)+" times")
				} else if i == maxretries {
					fmt.Printf("\r\033[31m%s\033[0m\n", "	retrying "+strconv.Itoa(i)+"/"+strconv.Itoa(maxretries)+" times")
				}
				time.Sleep(1 * time.Second)
			} else {
				defer resp.Body.Close()
				respn, _ := io.ReadAll(resp.Body)
				var r IPIPResp
				err := json.Unmarshal(respn, &r)
				if err != nil {
					errhandle("IP API calling error: " + err.Error())
					errhandle("	Error detail: " + err.Error())
					os.Exit(1)
				}
				ip := r.IP
				return string(ip)
			}
		}
		errhandle("\nIP API call failed " + fmt.Sprint(maxretries) + " times, exiting...")
		os.Exit(1)
	} else if api == "SB" {
		if maxretries == 0 {
			req, _ := http.NewRequest("GET", "https://api-ipv4.ip.sb/ip", nil)
			req.Header.Set("User-Agent", ua)
			resp, err := httpClient.Do(req)
			if err != nil || (resp.StatusCode >= 400 && resp.StatusCode <= 599) {
				errhandle("IP API calling error")
				errhandle("	Error detail: " + err.Error())
				os.Exit(1)
			}
			defer resp.Body.Close()
			ipo, _ := io.ReadAll(resp.Body)
			ip := strings.TrimRight(string(ipo), "\n")
			return ip
		}
		for i := 0; i <= maxretries; i++ {
			req, _ := http.NewRequest("GET", "https://api-ipv4.ip.sb/ip", nil)
			req.Header.Set("User-Agent", ua)
			resp, err := httpClient.Do(req)
			if err != nil || (resp.StatusCode >= 400 && resp.StatusCode <= 599) {
				if i == 0 {
					errhandle("IP API calling error:")
					errhandle("	Error detail: " + err.Error())
				}
				if i == 0 && maxretries == 1 {
					fmt.Printf("\r\033[31m%s\033[0m\n", "	retrying "+strconv.Itoa(i)+"/"+strconv.Itoa(maxretries)+" time")
				} else if i == 1 {
					fmt.Printf("\r\033[31m%s\033[0m", "	retrying "+strconv.Itoa(i)+"/"+strconv.Itoa(maxretries)+" time")
				} else if i > 1 && maxretries > 1 {
					fmt.Printf("\r\033[31m%s\033[0m", "	retrying "+strconv.Itoa(i)+"/"+strconv.Itoa(maxretries)+" times")
				} else if i == maxretries {
					fmt.Printf("\r\033[31m%s\033[0m\n", "	retrying "+strconv.Itoa(i)+"/"+strconv.Itoa(maxretries)+" times")
				}
				time.Sleep(1 * time.Second)
			} else {
				defer resp.Body.Close()
				ipo, _ := io.ReadAll(resp.Body)
				ip := strings.TrimRight(string(ipo), "\n")
				return ip
			}
		}
		errhandle("\nIP API call failed " + fmt.Sprint(maxretries) + " times")
		os.Exit(1)
	} else if api == "IPCONF" {
		if maxretries == 0 {
			req, _ := http.NewRequest("GET", "https://ifconfig.co/ip", nil)
			req.Header.Set("User-Agent", ua)
			resp, err := httpClient.Do(req)
			if err != nil || (resp.StatusCode >= 400 && resp.StatusCode <= 599) {
				errhandle("IP API calling error")
				errhandle("	Error detail: " + err.Error())
				os.Exit(1)
			}
			defer resp.Body.Close()
			ip, _ := io.ReadAll(resp.Body)
			return string(ip)
		}

		for i := 0; i <= maxretries; i++ {
			req, _ := http.NewRequest("GET", "https://ifconfig.co/ip", nil)
			req.Header.Set("User-Agent", ua)
			resp, err := httpClient.Do(req)
			if err != nil || (resp.StatusCode >= 400 && resp.StatusCode <= 599) {
				if i == 0 {
					errhandle("IP API calling error:")
					errhandle("	Error detail: " + err.Error())
				}
				if i == 0 && maxretries == 1 {
					fmt.Printf("\r\033[31m%s\033[0m\n", "	retrying "+strconv.Itoa(i)+"/"+strconv.Itoa(maxretries)+" time")
				} else if i == 1 {
					fmt.Printf("\r\033[31m%s\033[0m", "	retrying "+strconv.Itoa(i)+"/"+strconv.Itoa(maxretries)+" time")
				} else if i > 1 && maxretries > 1 {
					fmt.Printf("\r\033[31m%s\033[0m", "	retrying "+strconv.Itoa(i)+"/"+strconv.Itoa(maxretries)+" times")
				} else if i == maxretries {
					fmt.Printf("\r\033[31m%s\033[0m\n", "	retrying "+strconv.Itoa(i)+"/"+strconv.Itoa(maxretries)+" times")
				}
				time.Sleep(1 * time.Second)
			} else {
				defer resp.Body.Close()
				ip, _ := io.ReadAll(resp.Body)
				return strings.TrimSpace(string(ip))
			}
		}
		errhandle("\nIP API call failed " + fmt.Sprint(maxretries) + " times, exiting...")
		os.Exit(1)
	} else {
		errhandle("IP API calling error: unknown API " + api)
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
		errhandle("Error while fetching rules for lighthouse:")
		errhandle("	" + err.Error())
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
		errhandle("Error while modifying rules for lighthouse:")
		errhandle("	" + err.Error())
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
		errhandle("Error while fetching rules for security group:")
		errhandle("	" + err.Error())
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
		errhandle("Error while modifying rules for security group:")
		errhandle("	" + err.Error())
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

func errhandle(errmsg string) {
	fmt.Printf("\033[31m%s\033[0m\n", errmsg)
}
