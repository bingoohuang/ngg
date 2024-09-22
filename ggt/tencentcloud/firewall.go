package firewall

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bingoohuang/ngg/cmd"
	"github.com/bingoohuang/ngg/ggt/root"
	"github.com/bingoohuang/ngg/gum"
	"github.com/bingoohuang/ngg/ss"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	lighthouse "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"
)

func init() {
	fc := &subCmd{}
	c := &cobra.Command{
		Use:   "tencentcloud",
		Short: "tencentcloud firewall",
		RunE:  fc.run,
	}

	root.AddCommand(c, fc)
}

type subCmd struct {
	InstanceId string `short:"i" help:"InstanceId."`
	File       string `short:"f" help:"防火墙规则JSON文件, e.g. firewall-xxx.json"`
}

func (r *subCmd) run(cmd *cobra.Command, args []string) error {
	if r.File != "" {
		return r.modifyRules(r.File)
	}
	return r.listRules()
}

func (r *subCmd) modifyRules(file string) error {
	if _, err := os.Stat(file); err != nil {
		return err
	}

	fileData, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	var rules InstanceFirewallRules
	if err := json.Unmarshal(fileData, &rules); err != nil {
		return err
	}

	rq := lighthouse.NewModifyFirewallRulesRequest()
	rq.InstanceId = rules.InstanceId
	for _, rule := range rules.Rules {
		for _, protocol := range rule.Protocol {
			rq.FirewallRules = append(rq.FirewallRules, &lighthouse.FirewallRule{
				Protocol:                protocol,
				Port:                    rule.Port,
				CidrBlock:               rule.CidrBlock,
				Action:                  rule.Action,
				FirewallRuleDescription: rule.FirewallRuleDescription,
			})
		}
	}

	rsp, err := getClient().ModifyFirewallRules(rq)
	if err != nil {
		return err
	}

	resp, _ := json.Marshal(rsp)
	log.Printf("ModifyFirewallRules: %s", resp)
	return nil
}

func (r *subCmd) listRules() error {
	rq := lighthouse.NewDescribeFirewallRulesRequest()
	if r.InstanceId == "" {
		r.InstanceId = LightHouse.InstanceId
	}
	rq.InstanceId = &r.InstanceId

	// https://console.cloud.tencent.com/api/explorer?Product=lighthouse&Version=2020-03-24&Action=DescribeFirewallRules
	// 返回的resp是一个DescribeFirewallRulesResponse的实例，与请求对象对应
	response, err := getClient().DescribeFirewallRules(rq)
	if err != nil {
		return err
	}

	rules := InstanceFirewallRules{
		InstanceId: rq.InstanceId,
	}
	for _, rule := range response.Response.FirewallRuleSet {
		rules.Rules = append(rules.Rules, FirewallRule{
			Protocol:                []*string{rule.Protocol},
			Port:                    rule.Port,
			CidrBlock:               rule.CidrBlock,
			Action:                  rule.Action,
			FirewallRuleDescription: rule.FirewallRuleDescription,
		})
	}
	rules.mergeRules()

	jsonRules, err := json.MarshalIndent(rules, "", "    ")
	if err != nil {
		return err
	}

	// Create a temporary file
	file, err := ss.WriteTempFile("", "*.json", jsonRules, false)
	if err != nil {
		return err
	}
	defer os.Remove(file)

	log.Printf("cmd: %s", cmd.ShellQuoteMust(os.Args[0], "firewall", "-f", file))

	editorCmd, err := FindAvailableCmd("/usr/local/bin/zed", "/usr/local/bin/code")
	if err != nil {
		return nil
	}
	c := cmd.New(cmd.ShellQuoteMust(editorCmd, file))
	if err = c.Run(context.Background()); err != nil {
		return nil
	}

	yes, err := gum.Confirm("确认修改防火墙规则么?")
	if err != nil {
		return err
	}

	if !yes {
		return nil
	}

	if err := r.modifyRules(file); err != nil {
		return err
	}

	return nil
}

type FirewallRule struct {
	// 协议，取值：TCP，UDP，ICMP，ALL。
	Protocol []*string `json:"Protocol,omitempty" name:"Protocol"`

	// 端口，取值：ALL，单独的端口，逗号分隔的离散端口，减号分隔的端口范围。
	Port *string `json:"Port,omitempty" name:"Port"`

	// IPv4网段或 IPv4地址(互斥)。
	// 示例值：0.0.0.0/0。
	//
	// 和Ipv6CidrBlock互斥，两者都不指定时，如果Protocol不是ICMPv6，则取默认值0.0.0.0/0。
	CidrBlock *string `json:"CidrBlock,omitempty" name:"CidrBlock"`

	// 取值：ACCEPT，DROP。默认为 ACCEPT。
	Action *string `json:"Action,omitempty" name:"Action"`

	// 防火墙规则描述。
	FirewallRuleDescription *string `json:"FirewallRuleDescription,omitempty" name:"FirewallRuleDescription"`

	merged bool
}

type InstanceFirewallRules struct {
	InstanceId *string
	Rules      []FirewallRule
}

func (r *InstanceFirewallRules) mergeRules() {
	for i, ji := range r.Rules {
		for j := i + 1; j < len(r.Rules); j++ {
			jr := r.Rules[j]
			if !jr.merged && *jr.Action == *ji.Action &&
				*jr.Port == *ji.Port && *jr.CidrBlock == *ji.CidrBlock {
				r.Rules[j].merged = true
				r.Rules[i].Protocol = append(ji.Protocol, jr.Protocol...)
				if !strings.Contains(*ji.FirewallRuleDescription, *jr.FirewallRuleDescription) {
					*r.Rules[i].FirewallRuleDescription += "; " + *jr.FirewallRuleDescription
				}
			}
		}
	}

	rules := make([]FirewallRule, 0, len(r.Rules))
	for _, ji := range r.Rules {
		if !ji.merged {
			rules = append(rules, ji)
		}
	}

	r.Rules = rules
}

var (
	_clientOnce sync.Once
	_client     *lighthouse.Client
)

func getClient() *lighthouse.Client {
	_clientOnce.Do(func() {
		// 实例化一个认证对象，入参需要传入腾讯云账户 SecretId 和 SecretKey，此处还需注意密钥对的保密
		// 代码泄露可能会导致 SecretId 和 SecretKey 泄露，并威胁账号下所有资源的安全性。以下代码示例仅供参考，
		// 建议采用更安全的方式来使用密钥，请参见：https://cloud.tencent.com/document/product/1278/85305
		// 密钥可前往官网控制台 https://console.cloud.tencent.com/cam/capi 进行获取
		c := common.NewCredential(LightHouse.SecretID, LightHouse.SecretKey)
		// 实例化一个client选项，可选的，没有特殊需求可以跳过
		p := profile.NewClientProfile()
		p.HttpProfile.Endpoint = LightHouse.Endpoint
		// 实例化要请求产品的client对象,clientProfile是可选的
		var err error
		_client, err = lighthouse.NewClient(c, LightHouse.Region, p)
		if err != nil {
			panic(err)
		}
	})

	return _client
}

const jsonFile = ".lighthouse.json"

var LightHouse = func() (lh LightHouseConf) {
	jsonFileRewrite := false

	ReadHomeJsonFile(jsonFile, &lh)

	if env := os.Getenv("LIGHTHOUSE_SECRET"); env != "" {
		parts := strings.Split(env, ":")
		if len(parts) < 2 {
			log.Printf("bad $LIGHTHOUSE_SECRET")
		} else {
			lh.SecretID = parts[0]
			lh.SecretKey = parts[1]
			if len(parts) > 2 {
				lh.InstanceId = parts[2]
			}
			if len(parts) > 3 {
				lh.Region = parts[3]
			}
			jsonFileRewrite = true
		}
	}

	if lh.Region == "" {
		lh.Region = "ap-beijing"
		jsonFileRewrite = true
	}
	if lh.Endpoint == "" {
		lh.Endpoint = "lighthouse.tencentcloudapi.com"
		jsonFileRewrite = true
	}

	if jsonFileRewrite {
		if err := WriteHomeJsonFile(jsonFile, lh); err != nil {
			log.Printf("write %s error: %v", jsonFile, err)
		}
	}

	return lh
}()

type LightHouseConf struct {
	SecretID   string `json:"secretID"`
	SecretKey  string `json:"secretKey"`
	InstanceId string `json:"instanceId"`
	Region     string `json:"region"`
	Endpoint   string `json:"endpoint"`
}

func FindAvailableCmd(cmds ...string) (string, error) {
	for _, cmd := range cmds {
		if _, err := os.Stat(cmd); err == nil {
			return cmd, nil
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("stat %s: %w", cmd, err)
		}
	}

	return "", errors.New("cmc not found")
}

func WriteHomeJsonFile[T any](name string, v T) error {
	home, err := homedir.Dir()
	if err != nil {
		return fmt.Errorf("home dir: %w", err)
	}

	f := filepath.Join(home, name)
	return WriteJSONFile(f, v)
}

func ReadHomeJsonFile[T any](name string, v *T) (*T, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, fmt.Errorf("home dir: %w", err)
	}

	f := filepath.Join(home, name)
	return ReadJSONFile(f, v)
}

func WriteJSONFile[T any](file string, v T) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err := os.WriteFile(file, data, os.ModePerm); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

func ReadJSONFile[T any](file string, v *T) (*T, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", file, err)
	}

	if err := json.Unmarshal(data, v); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", file, err)
	}

	return v, nil
}
