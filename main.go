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
	"runtime"
	"sort"
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
	version         = "0.0.0"           // 程序版本号
	goos            = runtime.GOOS      // 程序运行的操作系统
	goarch          = runtime.GOARCH    // 程序运行的操作系统架构
	buildTime       = "buildTime"       // 程序编译时间
	action          string              // 程序运行的行为
	EnableWinNotify = false             // 是否启用 windows 通知
	notifyHelpMsg   = ""                // 帮助信息中的通知信息
	ua              = "qcip/" + version // 请求的 User-Agent
	confPath        = "config.json"     // 默认配置文件路径
	errMsgList      map[int]string      // 错误信息列表
	errHandleTimes  = 0                 // 错误输出的次数
	httpClient      = &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("tcp4", addr)
			},
		},
	}
	notify = func(title, msg string, succeed bool) {} // 默认禁用的通知函数
	ipAddr string                                     // 用户的IP地址
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
	EnableWinNotify     bool
	Rules               []string
}

type IPIPResp struct {
	IP string `json:"ip"`
}

func init() {
	errMsgList = make(map[int]string)
}

func main() {
	if len(os.Args) == 1 {
		keyFunc()
		return
	} else {
		for i, arg := range os.Args {
			if arg == "-h" || arg == "--help" {
				if action != "" {
					errOutput("Error arguments: " + arg + " cannot be used with other arguments\nRun \033[33mqcip -h\033[31m for help")
					return
				}
				action = "help"
			} else if arg == "-v" || arg == "--version" {
				if action != "" {
					errOutput("Error arguments: " + arg + " cannot be used with other arguments\nRun \033[33mqcip -h\033[31m for help")
					return
				}
				action = "version"
			} else if arg == "-c" || arg == "--config" {
				if action != "" {
					errOutput("Error arguments: " + arg + " cannot be used with other arguments\nRun \033[33mqcip -h\033[31m for help")
					return
				}
				action = "run"
				if i == len(os.Args)-1 {
					errOutput("Error arguments: config path not defined\nRun \033[33mqcip -h\033[31m for help")
					return
				}
				confPath = os.Args[i+1]
			} else if arg == "-n" || arg == "--winnotify" {
				if goos == "windows" {
					EnableWinNotify = true
				} else {
					errOutput("Error arguments: " + arg + " is only available on Windows")
					return
				}
			} else if arg == "-ip" || arg == "--ipaddr" {
				if i == len(os.Args)-1 {
					errOutput("Error arguments: ip address not defined\nRun \033[33mqcip -h\033[31m for help")
					return
				}
				ipAddr = os.Args[i+1]
				if net.ParseIP(ipAddr) == nil {
					errOutput("Error arguments: ip address is incorrect\nRun \033[33mqcip -h\033[31m for help")
					return
				}
			}
		}
		if action == "run" {
			keyFunc()
		} else if action == "version" {
			if ipAddr != "" {
				errOutput("Error arguments: you can only specify ip address when the program runs\nRun \033[33mqcip -h\033[31m for help")
				return
			}
			if EnableWinNotify {
				errOutput("Error arguments: you can only enable notifacation when the program runs\nRun \033[33mqcip -h\033[31m for help")
				return
			}
			showVersionInfo()
		} else if action == "help" {
			if ipAddr != "" {
				errOutput("Error arguments: you can only specify ip address when the program runs\nRun \033[33mqcip -h\033[31m for help")
				return
			}
			if EnableWinNotify {
				errOutput("Error arguments: you can only enable notifacation when the program runs\nRun \033[33mqcip -h\033[31m for help")
				return
			}
			fmt.Printf("QCIP \033[1;32mv%s\033[0m\nUsuage:	qcip [options] [<value>]\nOptions:\n  -c  --config <path>\tSpecify the location of the configuration file and run\n  -v  --version\t\tShow version information\n  -h  --help\t\tShow this help page\n  -ip --ipaddr <ip>\tSpecify to use custom ip address%s\nExamples:\n  \033[33mqcip\033[0m\tRun the program with config.json\n  \033[33mqcip -c qcipconf.json\033[0m\tSpecify to use the configuration file qcipconf.json and run the program\n  \033[33mqcip -ip 1.1.1.1\033[0m\tSpecify to use ip 1.1.1.1 instead of autoget\nVisit our Github repo for more helps\n  https://github.com/cnlancehu/qcip\n", version, notifyHelpMsg)
		} else if action == "" && EnableWinNotify {
			keyFunc()
		} else if action == "" && ipAddr != "" {
			keyFunc()
		} else {
			errOutput("Error arguments: unknown arguments\nRun \033[33mqcip -h\033[31m for help")
			return
		}
	}
}

// 功能主函数
func keyFunc() {
	fmt.Printf("QCIP \033[1;32mv%s\033[0m\n", version)
	configData := getConfig(confPath)
	maxRetries, _ := strconv.Atoi(configData.MaxRetries)
	if ipAddr == "" {
		ipAddr = getIPaddr(configData.GetIPAPI, maxRetries)
	}
	if configData.MType == "lh" {
		lhMain(configData, ipAddr)
	} else if configData.MType == "cvm" {
		cvmMain(configData, ipAddr)
	}
	os.Exit(0)
}

// 轻量应用服务器主函数
func lhMain(configData Config, ip string) {
	credential := common.NewCredential(
		configData.SecretId,
		configData.SecretKey,
	)
	rules := LHGetRules(credential, configData.InstanceRegion, configData.InstanceId)
	res, needUpdate := LHMatch(rules, ip, configData)
	if needUpdate {
		fmt.Printf("IP is different, start updating\n")
		LHModifyRules(credential, configData.InstanceRegion, configData.InstanceId, res)
		fmt.Printf("Successfully modified the firewall rules\n")
		if EnableWinNotify {
			notify("QCIP | Success", "Successfully modified the firewall rules", true)
		}
	} else {
		fmt.Printf("IP is the same\n")
		if EnableWinNotify {
			notify("QCIP | Success", "IP is the same", true)
		}
	}
}

// 云服务器主函数
func cvmMain(configData Config, ip string) {
	credential := common.NewCredential(
		configData.SecretId,
		configData.SecretKey,
	)
	rules := SGGetRules(credential, configData.SecurityGroupId, configData.SecurityGroupRegion)
	res, needUpdate := SGMatch(rules, ip, configData)
	if needUpdate {
		fmt.Printf("IP is different, start updating\n")
		SGModifyRules(credential, configData.SecurityGroupId, configData.SecurityGroupRegion, res)
		fmt.Printf("Successfully modified the firewall rules\n")
		if EnableWinNotify {
			notify("QCIP | Success", "Successfully modified the firewall rules", true)
		}
	} else {
		fmt.Printf("IP is the same\n")
		if EnableWinNotify {
			notify("QCIP | Success", "IP is the same", true)
		}
	}
}

func showVersionInfo() {
	fmt.Printf("QCIP \033[1;32mv%s\033[0m | \033[1;33m%s %s\033[0m\nBuild time: %s\nChecking for update...", version, goos, goarch, buildTime)
	req, _ := http.NewRequest("GET", "https://api.lance.fun/proj/qcip/version", nil)
	req.Header.Set("User-Agent", ua)
	resp, err := httpClient.Do(req)
	if err != nil || (resp.StatusCode >= 400 && resp.StatusCode <= 599) {
		fmt.Printf("\r\033[31mFailed to check updates\033[0m\n")
		return
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			fmt.Printf("\r\033[31mFailed to check updates\033[0m\n")
			return
		}
	}(resp.Body)
	errCheck := func() {
		if err != nil {
			fmt.Printf("\r\033[31mFailed to check updates\033[0m\n")
			errExit()
		}
	}
	latestverbyte, err := io.ReadAll(resp.Body)
	errCheck()
	latestVersion := string(latestverbyte)
	currentVersion, err := strconv.Atoi(strings.Replace(version, ".", "", -1))
	errCheck()
	verlatest, err := strconv.Atoi(strings.Replace(latestVersion, ".", "", -1))
	errCheck()
	if verlatest > currentVersion {
		fmt.Printf("\rNew version available: \033[1;32m%s\033[0m\nDownload it here: \n  https://github.com/cnlancehu/qcip/releases/tag/%s\n", latestVersion, latestVersion)
	} else {
		fmt.Printf("\r\033[1;32mYou are using the latest version\033[0m\n")
	}
}

// 读取配置文件
func getConfig(confPath string) Config {
	config, err := os.ReadFile(confPath)
	if err != nil {
		if os.IsNotExist(err) {
			errOutput("Config error: config file " + confPath + " does not exist")
			errExit()
		}
		errOutput("Config error: " + err.Error())
		errExit()
	}
	var configData Config
	if !json.Valid(config) {
		errOutput("Config error: config file is not valid json")
		errExit()
	}
	err = json.Unmarshal(config, &configData)
	if err != nil {
		decodeErr := json.Unmarshal(config, &configData)
		if decodeErr != nil {
			errOutput("Config error: config file format is incorrect")
			errExit()
		}
		errOutput("Config error: " + err.Error())
		errExit()
	}
	if configData.EnableWinNotify {
		if goos != "windows" {
			errOutput("Config error: EnableWinNotify is only available on Windows")
			errExit()
		}
		EnableWinNotify = true
	}
	var requiredKeys []string
	if configData.MType == "lh" {
		requiredKeys = []string{"SecretId", "SecretKey", "InstanceId", "InstanceRegion", "Rules"}
	} else if configData.MType == "cvm" {
		requiredKeys = []string{"SecretId", "SecretKey", "SecurityGroupId", "SecurityGroupRegion", "Rules"}
	} else {
		if configData.MType == "" {
			errOutput("Config error: machine type is empty")
			errExit()
		} else {
			errOutput("Config error: machine type " + configData.MType + " is incorrect")
			errExit()
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
		errOutput("Config error:")
		for _, key := range requiredKeys {
			if _, ok := reflect.TypeOf(configData).FieldByName(key); !ok {
				errOutput("\t" + key + " not found")
			} else if reflect.ValueOf(configData).FieldByName(key).String() == "" {
				errOutput("\t" + key + " is empty")
			} else if reflect.ValueOf(configData).FieldByName(key).String() == key {
				errOutput("\t" + key + " is incorrect")
			}
		}
		errExit()
	}
	fmt.Printf("Config loaded\n")
	return configData
}

// 获取自身公网IP
func getIPaddr(api string, maxRetries int) string {
	if maxRetries < 0 || maxRetries > 10 {
		errOutput("Config error: maxRetries should be an integer greater than or equal to 0 and less than or equal to 10")
	}
	fetchApi := func(apiURL string) []byte {
		var (
			req    *http.Request
			resp   *http.Response
			err    error
			failed bool
		)
		for i := 0; i <= maxRetries; i++ {
			req, _ = http.NewRequest("GET", apiURL, nil)
			req.Header.Set("User-Agent", ua)
			resp, err = httpClient.Do(req)
			if err != nil || (resp.StatusCode >= 400 && resp.StatusCode <= 599) {
				failed = true
				if i == 0 {
					errOutput("IP API calling error:")
					errOutput("\tError detail: " + err.Error())
					continue
				}
				if maxRetries != 0 {
					fmt.Printf("\r\033[31m%s\033[0m", "	retrying "+strconv.Itoa(i)+"/"+strconv.Itoa(maxRetries)+" time")
					time.Sleep(1 * time.Second)
				}
			} else {
				break
			}
		}
		if failed {
			if maxRetries != 0 {
				errOutput("\nIP API call failed " + fmt.Sprint(maxRetries) + " times, exiting...")
			} else {
				errOutput("IP API call failed, exiting...")
			}
			errExit()
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				errOutput("IP API calling error")
				errOutput("\tError detail: " + err.Error())
				errExit()
				return
			}
		}(resp.Body)
		respcontent, err := io.ReadAll(resp.Body)
		if err != nil {
			errOutput("IP API calling error: " + err.Error())
			errExit()
		}
		return respcontent
	}
	if api == "LanceAPI" {
		return string(fetchApi("https://api.lance.fun/ip"))
	} else if api == "IPIP" {
		var r IPIPResp
		err := json.Unmarshal(fetchApi("https://myip.ipip.net/ip"), &r)
		if err != nil {
			errOutput("IP API calling error: " + err.Error())
			errOutput("\tError detail: " + err.Error())
			errExit()
		}
		return r.IP
	} else if api == "SB" {
		return strings.TrimRight(string(fetchApi("https://api-ipv4.ip.sb/ip")), "\n")
	} else if api == "IPCONF" || api == "" {
		return strings.TrimSpace(string(fetchApi("https://ifconfig.co/ip")))
	} else {
		errOutput("IP API calling error: unknown API " + api)
		errExit()
		return ""
	}
}

// 轻量应用服务器部分
func LHGetRules(credential *common.Credential, InstanceRegion string, InstanceId string) []*lighthouse.FirewallRuleInfo {
	cpf := profile.NewClientProfile()
	cpf.NetworkFailureMaxRetries = 3
	cpf.HttpProfile.Endpoint = "lighthouse.tencentcloudapi.com"
	client, _ := lighthouse.NewClient(credential, InstanceRegion, cpf)
	request := lighthouse.NewDescribeFirewallRulesRequest()
	request.InstanceId = common.StringPtr(InstanceId)
	request.Offset = common.Int64Ptr(0)
	request.Limit = common.Int64Ptr(100)
	response, err := client.DescribeFirewallRules(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		errOutput("Error while fetching rules for lighthouse:")
		errOutput("\t" + err.Error())
		errExit()
	}
	return response.Response.FirewallRuleSet
}

func LHMatch(rules []*lighthouse.FirewallRuleInfo, ip string, config Config) ([]*lighthouse.FirewallRuleInfo, bool) {
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

func LHModifyRules(credential *common.Credential, InstanceRegion string, InstanceId string, rules []*lighthouse.FirewallRuleInfo) {
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
	cpf.NetworkFailureMaxRetries = 3
	cpf.HttpProfile.Endpoint = "lighthouse.tencentcloudapi.com"
	client, _ := lighthouse.NewClient(credential, InstanceRegion, cpf)
	request := lighthouse.NewModifyFirewallRulesRequest()
	request.InstanceId = common.StringPtr(InstanceId)
	request.FirewallRules = ptrRules
	_, err := client.ModifyFirewallRules(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		errOutput("Error while modifying rules for lighthouse:")
		errOutput("\t" + err.Error())
		errExit()
		return
	}
}

// 云服务器安全组部分
func SGGetRules(credential *common.Credential, SecurityGroupId string, SecurityGroupRegion string) *vpc.SecurityGroupPolicySet {
	cpf := profile.NewClientProfile()
	cpf.NetworkFailureMaxRetries = 3
	cpf.HttpProfile.Endpoint = "vpc.tencentcloudapi.com"
	client, _ := vpc.NewClient(credential, SecurityGroupRegion, cpf)
	request := vpc.NewDescribeSecurityGroupPoliciesRequest()
	request.SecurityGroupId = common.StringPtr(SecurityGroupId)
	response, err := client.DescribeSecurityGroupPolicies(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		errOutput("Error while fetching rules for security group:")
		errOutput("\t" + err.Error())
		errExit()
	}
	return response.Response.SecurityGroupPolicySet
}

func SGMatch(rules *vpc.SecurityGroupPolicySet, ip string, config Config) (*vpc.SecurityGroupPolicySet, bool) {
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

func SGModifyRules(credential *common.Credential, SecurityGroupId string, SecurityGroupRegion string, rules *vpc.SecurityGroupPolicySet) {
	cpf := profile.NewClientProfile()
	cpf.NetworkFailureMaxRetries = 3
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
	FinalRules := processRules(ptrRules).(*vpc.SecurityGroupPolicySet)
	request.SecurityGroupPolicySet = FinalRules
	_, err := client.ModifySecurityGroupPolicies(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		errOutput("Error while modifying rules for security group:")
		errOutput("\t" + err.Error())
		errExit()
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

func errOutput(errMsg string) {
	if EnableWinNotify {
		errHandleTimes++
		errMsgList[errHandleTimes] = errMsg
	}
	fmt.Printf("\033[31m%s\033[0m\n", errMsg)
}

func errExit() {
	if EnableWinNotify {
		var (
			keys      []int
			allErrMsg string
		)
		for k := range errMsgList {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		for _, k := range keys {
			allErrMsg += errMsgList[k] + "\n"
		}
		allErrMsg = strings.ReplaceAll(allErrMsg, "\t", "    ")
		notify("QCIP | Error", allErrMsg, false)
	}
	os.Exit(1)
}
