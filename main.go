package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
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
)

var (
	version    = "Development"
	ua         = "QCIP/" + version
	httpClient = &http.Client{
		Timeout: time.Second * 10,
	}
)

type Config struct {
	SecretId       string
	SecretKey      string
	GetIPAPI       string
	InstanceId     string
	InstanceRegion string
	MaxRetries     string
	Rules          []string
}

type IPResponse struct {
	IP string `json:"ip"`
}

type FirewallRule struct {
	AppType                 string `json:"AppType"`
	Protocol                string `json:"Protocol"`
	Port                    string `json:"Port"`
	CidrBlock               string `json:"CidrBlock"`
	Action                  string `json:"Action"`
	FirewallRuleDescription string `json:"FirewallRuleDescription"`
}
type Response struct {
	TotalCount      int            `json:"TotalCount"`
	FirewallRuleSet []FirewallRule `json:"FirewallRuleSet"`
	FirewallVersion int            `json:"FirewallVersion"`
	RequestId       string         `json:"RequestId"`
}
type Rules struct {
	Response Response `json:"Response"`
}

func main() {
	fmt.Printf("QCIP \033[32mv%s\033[0m\n", version)
	configData := getconfig()
	maxRetries, _ := strconv.Atoi(configData.MaxRetries)
	ip := getip(configData.GetIPAPI, int(maxRetries))
	credential := common.NewCredential(
		configData.SecretId,
		configData.SecretKey,
	)
	rules := getrules(credential, configData.InstanceRegion, configData.InstanceId)
	res, needUpdate := match(rules, ip, configData)
	if needUpdate {
		fmt.Printf("IP is different, start updating\n")
		modifyrules(credential, configData.InstanceRegion, configData.InstanceId, res)
		fmt.Printf("Successfully modified the firewall rules\n")
		os.Exit(1)
	} else {
		fmt.Printf("IP is the same\n")
		os.Exit(1)
	}
}

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
	requiredKeys := []string{"SecretId", "SecretKey", "InstanceId", "InstanceRegion", "Rules"}
	checkPassing := true
	for _, key := range requiredKeys {
		if _, ok := reflect.TypeOf(configData).FieldByName(key); !ok {
			fmt.Println(key + " not found in config file")
			checkPassing = false
		}
	}
	if !checkPassing {
		os.Exit(1)
	}
	fmt.Printf("Config loaded\n")
	return configData
}

func getip(api string, maxretries int) string {
	if api == "LanceAPI" {
		for i := 0; i < maxretries; i++ {
			req, _ := http.NewRequest("GET", "https://api.lance.fun/ip", nil)
			req.Header.Set("User-Agent", ua)
			resp, err := httpClient.Do(req)
			if err != nil {
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
			if err != nil {
				fmt.Printf("IP API call failed, retrying %d time...\n", i+1)
			} else {
				defer resp.Body.Close()
				respn, _ := io.ReadAll(resp.Body)
				var r IPResponse
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
			if err != nil {
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

func getrules(credential *common.Credential, InstanceRegion string, InstanceId string) []FirewallRule {
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
	orrules := response.ToJsonString()
	data := []byte(string(orrules))
	var rules Rules
	nerr := json.Unmarshal(data, &rules)
	if nerr != nil {
		log.Fatalln(nerr)
		os.Exit(1)
	}
	return rules.Response.FirewallRuleSet
}

func match(rules []FirewallRule, ip string, config Config) ([]FirewallRule, bool) {
	needUpdate := false
	for a := range rules {
		for b := range config.Rules {
			if rules[a].FirewallRuleDescription == config.Rules[b] {
				if rules[a].CidrBlock == ip {
					continue
				} else {
					rules[a].CidrBlock = ip
					needUpdate = true
				}
			}
		}
	}
	return rules, needUpdate
}

func modifyrules(credential *common.Credential, InstanceRegion string, InstanceId string, rules []FirewallRule) {
	ptrRules := make([]*lighthouse.FirewallRule, len(rules))
	for i := range rules {
		ptrRules[i] = &lighthouse.FirewallRule{
			Protocol:                common.StringPtr(rules[i].Protocol),
			Port:                    common.StringPtr(rules[i].Port),
			CidrBlock:               common.StringPtr(rules[i].CidrBlock),
			Action:                  common.StringPtr(rules[i].Action),
			FirewallRuleDescription: common.StringPtr(rules[i].FirewallRuleDescription),
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
