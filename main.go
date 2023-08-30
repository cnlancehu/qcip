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
	version        string         = "0.0.0"           // 程序版本号
	goos           string         = runtime.GOOS      // 程序运行的操作系统
	goarch         string         = runtime.GOARCH    // 程序运行的操作系统架构
	buildTime      string         = "buildTime"       // 程序编译时间
	action         string                             // 程序运行的行为
	notifa         bool                               // 是否启用 windows 通知
	notifyHelpMsg  string         = ""                // 帮助信息中的通知信息
	ua             string         = "qcip/" + version // 请求的 User-Agent
	confPath       string         = "config.json"     // 默认配置文件路径
	errmsglist     map[int]string                     // 错误信息列表
	errhandletimes int            = 0                 // 错误输出的次数
	httpClient     *http.Client   = &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("tcp4", addr)
			},
		},
	}
	notify = func(title, msg string, succeed bool) {} // 默认禁用的通知函数
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
	errmsglist = make(map[int]string)
}

func main() {
	if len(os.Args) == 1 {
		keyfunc()
		return
	} else {
		for i, arg := range os.Args {
			if arg == "-h" || arg == "--help" {
				if action != "" {
					errhandle("Error arguments: " + arg + " cannot be used with other arguments\nRun \033[33mqcip -h\033[31m for help")
					return
				}
				action = "help"
			} else if arg == "-v" || arg == "--version" {
				if action != "" {
					errhandle("Error arguments: " + arg + " cannot be used with other arguments\nRun \033[33mqcip -h\033[31m for help")
					return
				}
				action = "version"
			} else if arg == "-c" || arg == "--config" {
				if action != "" {
					errhandle("Error arguments: " + arg + " cannot be used with other arguments\nRun \033[33mqcip -h\033[31m for help")
					return
				}
				action = "run"
				if i == len(os.Args)-1 {
					errhandle("Error arguments: config path not defined\nRun \033[33mqcip -h\033[31m for help")
					return
				}
				confPath = os.Args[i+1]
			} else if arg == "-n" || arg == "--winnotify" {
				if goos == "windows" {
					notifa = true
				} else {
					errhandle("Error arguments: " + arg + " is only available on Windows")
					return
				}
			}
		}
		if action == "run" {
			keyfunc()
		} else if action == "version" {
			if notifa {
				errhandle("Error arguments: you can only enable notifacation when the program runs\nRun \033[33mqcip -h\033[31m for help")
				return
			}
			showversion()
		} else if action == "help" {
			if notifa {
				errhandle("Error arguments: you can only enable notifacation when the program runs\nRun \033[33mqcip -h\033[31m for help")
				return
			}
			fmt.Printf("QCIP \033[1;32mv%s\033[0m\nUsuage:	qcip [options] [<value>]\nOptions:\n  -c, --config <path>\tSpecify the location of the configuration file and run\n  -v, --version\t\tShow version information\n  -h, --help\t\tShow this help page%s\nExamples:\n  \033[33mqcip\033[0m\tRun the program with config.json\n  \033[33mqcip -c qcipconf.json\033[0m\tSpecify to use the configuration file qcipconf.json and run the program\nVisit our Github repo for more helps\n  https://github.com/cnlancehu/qcip\n", version, notifyHelpMsg)
		} else if action == "" && notifa {
			keyfunc()
		} else {
			errhandle("Error arguments: unknown arguments\nRun \033[33mqcip -h\033[31m for help")
		}
	}
}

// 功能主函数
func keyfunc() {
	fmt.Printf("QCIP \033[1;32mv%s\033[0m\n", version)
	configData := getconfig(confPath)
	maxRetries, _ := strconv.Atoi(configData.MaxRetries)
	ip := getip(configData.GetIPAPI, maxRetries)
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
		if notifa {
			notify("QCIP | Succsss", "Successfully modified the firewall rules", true)
		}
	} else {
		fmt.Printf("IP is the same\n")
		if notifa {
			notify("QCIP | Succsss", "IP is the same", true)
		}
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
		if notifa {
			notify("QCIP | Succsss", "Successfully modified the firewall rules", true)
		}
	} else {
		fmt.Printf("IP is the same\n")
		if notifa {
			notify("QCIP | Succsss", "IP is the same", true)
		}
	}
}

func showversion() {
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
	errcheck := func() {
		if err != nil {
			fmt.Printf("\r\033[31mFailed to check updates\033[0m\n")
			errexit()
		}
	}
	latestverbyte, err := io.ReadAll(resp.Body)
	errcheck()
	latestver := string(latestverbyte)
	vernow, err := strconv.Atoi(strings.Replace(version, ".", "", -1))
	errcheck()
	verlatest, err := strconv.Atoi(strings.Replace(latestver, ".", "", -1))
	errcheck()
	if verlatest > vernow {
		fmt.Printf("\rNew version available: \033[1;32m%s\033[0m\nDownload it here: \n  https://github.com/cnlancehu/qcip/releases/tag/%s\n", latestver, latestver)
	} else {
		fmt.Printf("\r\033[1;32mYou are using the latest version\033[0m\n")
	}
}

// 读取配置文件
func getconfig(confPath string) Config {
	config, err := os.ReadFile(confPath)
	if err != nil {
		if os.IsNotExist(err) {
			errhandle("Config error: config file " + confPath + " does not exist")
			errexit()
		}
		errhandle("Config error: " + err.Error())
		errexit()
	}
	var configData Config
	if !json.Valid(config) {
		errhandle("Config error: config file is not valid json")
		errexit()
	}
	err = json.Unmarshal(config, &configData)
	if err != nil {
		decodeErr := json.Unmarshal(config, &configData)
		if decodeErr != nil {
			errhandle("Config error: config file format is incorrect")
			errexit()
		}
		errhandle("Config error: " + err.Error())
		errexit()
	}
	if configData.EnableWinNotify {
		if goos != "windows" {
			errhandle("Config error: EnableWinNotify is only available on Windows")
			errexit()
		}
		notifa = true
	}
	var requiredKeys []string
	if configData.MType == "lh" {
		requiredKeys = []string{"SecretId", "SecretKey", "InstanceId", "InstanceRegion", "Rules"}
	} else if configData.MType == "cvm" {
		requiredKeys = []string{"SecretId", "SecretKey", "SecurityGroupId", "SecurityGroupRegion", "Rules"}
	} else {
		if configData.MType == "" {
			errhandle("Config error: machine type is empty")
			errexit()
		} else {
			errhandle("Config error: machine type " + configData.MType + " is incorrect")
			errexit()
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
				errhandle("\t" + key + " not found")
			} else if reflect.ValueOf(configData).FieldByName(key).String() == "" {
				errhandle("\t" + key + " is empty")
			} else if reflect.ValueOf(configData).FieldByName(key).String() == key {
				errhandle("\t" + key + " is incorrect")
			}
		}
		errexit()
	}
	fmt.Printf("Config loaded\n")
	return configData
}

// 获取自身公网IP
func getip(api string, maxretries int) string {
	if maxretries < 0 || maxretries > 10 {
		errhandle("Config error: maxretries should be an integer greater than or equal to 0 and less than or equal to 10")
	}
	fetchapi := func(apiaddress string) []byte {
		var (
			req    *http.Request
			resp   *http.Response
			err    error
			failed bool
		)
		for i := 0; i <= maxretries; i++ {
			req, _ = http.NewRequest("GET", apiaddress, nil)
			req.Header.Set("User-Agent", ua)
			resp, err = httpClient.Do(req)
			if err != nil || (resp.StatusCode >= 400 && resp.StatusCode <= 599) {
				failed = true
				if i == 0 {
					errhandle("IP API calling error:")
					errhandle("\tError detail: " + err.Error())
					continue
				}
				if maxretries != 0 {
					fmt.Printf("\r\033[31m%s\033[0m", "	retrying "+strconv.Itoa(i)+"/"+strconv.Itoa(maxretries)+" time")
					time.Sleep(1 * time.Second)
				}
			} else {
				break
			}
		}
		if failed {
			if maxretries != 0 {
				errhandle("\nIP API call failed " + fmt.Sprint(maxretries) + " times, exiting...")
			} else {
				errhandle("IP API call failed, exiting...")
			}
			errexit()
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				errhandle("IP API calling error")
				errhandle("\tError detail: " + err.Error())
				errexit()
				return
			}
		}(resp.Body)
		respcontent, err := io.ReadAll(resp.Body)
		if err != nil {
			errhandle("IP API calling error: " + err.Error())
			errexit()
		}
		return respcontent
	}
	if api == "LanceAPI" {
		return string(fetchapi("https://api.lance.fun/ip"))
	} else if api == "IPIP" {
		var r IPIPResp
		err := json.Unmarshal(fetchapi("https://myip.ipip.net/ip"), &r)
		if err != nil {
			errhandle("IP API calling error: " + err.Error())
			errhandle("\tError detail: " + err.Error())
			errexit()
		}
		return r.IP
	} else if api == "SB" {
		return strings.TrimRight(string(fetchapi("https://api-ipv4.ip.sb/ip")), "\n")
	} else if api == "IPCONF" || api == "" {
		return strings.TrimSpace(string(fetchapi("https://ifconfig.co/ip")))
	} else {
		errhandle("IP API calling error: unknown API " + api)
		errexit()
		return ""
	}
}

// 轻量应用服务器部分
func lhgetrules(credential *common.Credential, InstanceRegion string, InstanceId string) []*lighthouse.FirewallRuleInfo {
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
		errhandle("Error while fetching rules for lighthouse:")
		errhandle("\t" + err.Error())
		errexit()
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
	cpf.NetworkFailureMaxRetries = 3
	cpf.HttpProfile.Endpoint = "lighthouse.tencentcloudapi.com"
	client, _ := lighthouse.NewClient(credential, InstanceRegion, cpf)
	request := lighthouse.NewModifyFirewallRulesRequest()
	request.InstanceId = common.StringPtr(InstanceId)
	request.FirewallRules = ptrRules
	_, err := client.ModifyFirewallRules(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		errhandle("Error while modifying rules for lighthouse:")
		errhandle("\t" + err.Error())
		errexit()
		return
	}
}

// 云服务器安全组部分
func sggetrules(credential *common.Credential, SecurityGroupId string, SecurityGroupRegion string) *vpc.SecurityGroupPolicySet {
	cpf := profile.NewClientProfile()
	cpf.NetworkFailureMaxRetries = 3
	cpf.HttpProfile.Endpoint = "vpc.tencentcloudapi.com"
	client, _ := vpc.NewClient(credential, SecurityGroupRegion, cpf)
	request := vpc.NewDescribeSecurityGroupPoliciesRequest()
	request.SecurityGroupId = common.StringPtr(SecurityGroupId)
	response, err := client.DescribeSecurityGroupPolicies(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		errhandle("Error while fetching rules for security group:")
		errhandle("\t" + err.Error())
		errexit()
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
	rulesfin := processRules(ptrRules).(*vpc.SecurityGroupPolicySet)
	request.SecurityGroupPolicySet = rulesfin
	_, err := client.ModifySecurityGroupPolicies(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		errhandle("Error while modifying rules for security group:")
		errhandle("\t" + err.Error())
		errexit()
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
	if notifa {
		errhandletimes++
		errmsglist[errhandletimes] = errmsg
	}
	fmt.Printf("\033[31m%s\033[0m\n", errmsg)
}

func errexit() {
	if notifa {
		var (
			keys      []int
			allerrmsg string
		)
		for k := range errmsglist {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		for _, k := range keys {
			allerrmsg += errmsglist[k] + "\n"
		}
		allerrmsg = strings.ReplaceAll(allerrmsg, "\t", "    ")
		notify("QCIP | Error", allerrmsg, false)
	}
	os.Exit(1)
}
