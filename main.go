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
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	lighthouse "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"
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
	fmt.Printf("QCIP\n")
	var configpath string
	if len(os.Args) > 1 {
		configpath = os.Args[1]
	} else {
		configpath = "config.json"
	}
	configData := getConfig(configpath)
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

func getConfig(confPath string) Config {
	// Get config file
	config, err := os.ReadFile(confPath)
	if err != nil {
		fmt.Printf("Error reading config file: %s\n", err)
		os.Exit(1)
	}
	if !json.Valid(config) {
		fmt.Println("Error: config file is not valid json")
		os.Exit(1)
	}
	// Unmarshal config file into custom type
	var configData Config
	err = json.Unmarshal(config, &configData)
	if err != nil {
		fmt.Printf("Error unmarshalling config file: %s\n", err)
		os.Exit(1)
	}
	// Check if config file include all required fields
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
	fmt.Printf("Config load successfully\n")
	return configData
}

func getip(api string, maxretries int) string {
	if api == "LanceAPI" {
		for i := 0; i < maxretries; i++ {
			resp, err := http.Get("https://api.lance.fun/ip")
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
			resp, err := http.Get("https://myip.ipip.net/ip")
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
	} else {
		fmt.Println("Error: IP API not supported")
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
	// Convert slice of FirewallRule to slice of pointers to FirewallRule
	ptrRules := make([]*lighthouse.FirewallRule, len(rules))
	for i := range rules {
		ptrRules[i] = &lighthouse.FirewallRule{
			// AppType:                 common.StringPtr(rules[i].AppType),
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
	request.FirewallRules = ptrRules // use ptrRules instead of rules
	_, err := client.ModifyFirewallRules(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s\n", err)
		return
	}
	if err != nil {
		panic(err)
	}
}
