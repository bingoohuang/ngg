package main

import (
	"cmp"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strings"

	"github.com/bingoohuang/ngg/cmd"
	"github.com/bingoohuang/ngg/ggt/root"
	_ "github.com/bingoohuang/ngg/gnet/android"
	"github.com/bingoohuang/ngg/gum"
	"github.com/bingoohuang/ngg/ss"
	"github.com/spf13/cobra"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	lighthouse "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/lighthouse/v20200324"
	"golang.org/x/crypto/hkdf"
)

func main() {
	c := root.CreateCmd(nil, "tencentcloud", "tencentcloud firewall", &subCmd{})
	if err := c.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
	}
}

type subCmd struct {
	InstanceId string `short:"i" help:"InstanceId."`
	File       string `short:"f" help:"防火墙规则JSON文件, e.g. firewall-xxx.json"`
	Config     string `short:"c" help:"腾讯云Credential, e.g. tencent.json"`

	Desc string `short:"d" help:"防火墙规则描述关键字, 必须与 IP 一起使用时, 直接添加防火墙规则, 允许指定IP访问"`
	IP   string `help:"防火墙规则描述关键字允许的IP"`
}

func (r *subCmd) Run(cmd *cobra.Command, args []string) error {
	if r.Desc != "" && r.IP == "" || r.IP != "" && r.Desc == "" {
		return fmt.Errorf("desc and ip must be set together")
	}

	conf, err := ParseLightHouseConf(r.Config)
	if err != nil {
		return err
	}

	client, err := conf.NewClient()
	if err != nil {
		return err
	}

	if r.File != "" {
		return r.modifyRules(client, r.File)
	}

	return r.listRules(conf, client)
}

func (r *subCmd) modifyRules(client *lighthouse.Client, file string) error {
	f, err := ss.ExpandFilename(file)
	if err != nil {
		return err
	}

	fileData, err := os.ReadFile(f)
	if err != nil {
		return err
	}

	var rules InstanceFirewallRules
	if err := json.Unmarshal(fileData, &rules); err != nil {
		return err
	}

	return r.submitRules(client, rules)
}

func (r *subCmd) submitRules(client *lighthouse.Client, rules InstanceFirewallRules) error {
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

	rsp, err := client.ModifyFirewallRules(rq)
	if err != nil {
		return err
	}

	resp, _ := json.Marshal(rsp)
	log.Printf("ModifyFirewallRules: %s", resp)
	return nil
}

func (r *subCmd) listRules(conf *LightHouseConf, client *lighthouse.Client) error {
	rq := lighthouse.NewDescribeFirewallRulesRequest()
	if r.InstanceId == "" {
		r.InstanceId = conf.InstanceId
	}
	rq.InstanceId = &r.InstanceId

	// https://console.cloud.tencent.com/api/explorer?Product=lighthouse&Version=2020-03-24&Action=DescribeFirewallRules
	// 返回的resp是一个DescribeFirewallRulesResponse的实例，与请求对象对应
	response, err := client.DescribeFirewallRules(rq)
	if err != nil {
		return err
	}

	rules := InstanceFirewallRules{
		InstanceId: rq.InstanceId,
	}
	for _, rule := range response.Response.FirewallRuleSet {
		rules.Rules = append(rules.Rules, &FirewallRule{
			Protocol:                []*string{rule.Protocol},
			Port:                    rule.Port,
			CidrBlock:               rule.CidrBlock,
			Action:                  rule.Action,
			FirewallRuleDescription: rule.FirewallRuleDescription,
		})
	}
	rules.mergeRules()

	if r.IP != "" && r.Desc != "" {
		// 直接根据描述和 IP，添加防火墙规则
		found := false
		for _, rule := range rules.Rules {
			if strings.Contains(*rule.FirewallRuleDescription, r.Desc) {
				rule.Protocol = common.StringPtrs([]string{"UDP", "TCP"})
				rule.Port = common.StringPtr("ALL")
				rule.CidrBlock = common.StringPtr(r.IP)
				rule.Action = common.StringPtr("ACCEPT")
				found = true
			}
		}
		if !found {
			rules.Rules = append(rules.Rules, &FirewallRule{
				Protocol:                common.StringPtrs([]string{"UDP", "TCP"}),
				Port:                    common.StringPtr("ALL"),
				CidrBlock:               common.StringPtr(r.IP),
				Action:                  common.StringPtr("ACCEPT"),
				FirewallRuleDescription: common.StringPtr(r.Desc),
			})
		}

		return r.submitRules(client, rules)
	}

	jsonRules, err := json.MarshalIndent(rules, "", "    ")
	if err != nil {
		return err
	}
	// Create a temporary file
	file, err := ss.WriteTempFile("", "*.json", jsonRules, false)
	if err != nil {
		return err
	}
	// defer os.Remove(file)

	log.Printf("cmd: %s", cmd.ShellQuoteMust(os.Args[0], "firewall", "-f", file))

	editorCmd, err := FindAvailableCmd("/usr/local/bin/zed", "/usr/local/bin/code", "vim")
	if err != nil {
		return err
	}
	c := cmd.New(cmd.ShellQuoteMust(editorCmd, file))
	if err = c.Run(context.Background()); err != nil {
		return err
	}

	yes, err := gum.Confirm("确认修改防火墙规则么?")
	if err != nil {
		return err
	}

	if !yes {
		return nil
	}

	return r.modifyRules(client, file)
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
	Rules      []*FirewallRule
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

	rules := make([]*FirewallRule, 0, len(r.Rules))
	for _, ji := range r.Rules {
		if !ji.merged {
			rules = append(rules, ji)
		}
	}

	r.Rules = rules
}

func (l *LightHouseConf) NewClient() (*lighthouse.Client, error) {
	// 实例化一个认证对象，入参需要传入腾讯云账户 SecretId 和 SecretKey，此处还需注意密钥对的保密
	// 代码泄露可能会导致 SecretId 和 SecretKey 泄露，并威胁账号下所有资源的安全性。以下代码示例仅供参考，
	// 建议采用更安全的方式来使用密钥，请参见：https://cloud.tencent.com/document/product/1278/85305
	// 密钥可前往官网控制台 https://console.cloud.tencent.com/cam/capi 进行获取
	c := common.NewCredential(l.SecretID, l.SecretKey)
	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	p := profile.NewClientProfile()
	p.HttpProfile.Endpoint = l.Endpoint
	// 实例化要请求产品的client对象,clientProfile是可选的
	return lighthouse.NewClient(c, l.Region, p)
}

func ParseLightHouseConf(tencentCredenialJsonFile string) (*LightHouseConf, error) {
	tencentCredenialJsonFile = ss.Or(tencentCredenialJsonFile, "~/.lighthouse.json")

	var lh LightHouseConf
	optionFunc := WithSesame("Tx~F*qy1s1")
	if _, err := ReadJSONFile(tencentCredenialJsonFile, &lh, optionFunc); err != nil {
		return nil, err
	}

	lh.Region = cmp.Or(lh.Region, "ap-beijing")
	lh.Endpoint = cmp.Or(lh.Endpoint, "lighthouse.tencentcloudapi.com")

	if err := WriteJSONFile(tencentCredenialJsonFile, lh, optionFunc); err != nil {
		return nil, err
	}

	return &lh, nil
}

type LightHouseConf struct {
	SecretID   string `json:"secretID" sesame:"yes"`
	SecretKey  string `json:"secretKey" sesame:"yes"`
	InstanceId string `json:"instanceId"`
	Region     string `json:"region"`
	Endpoint   string `json:"endpoint"`
}

func FindAvailableCmd(cmds ...string) (string, error) {
	for _, cmd := range cmds {
		if path, err := exec.LookPath(cmd); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("cmds %v not found", cmds)
}

func WriteJSONFile[T any](file string, v T, opts ...JSONFileOptionFunc) error {
	opt := &JSONFileOption{}
	for _, o := range opts {
		o(opt)
	}

	if err := opt.ToggleSesame(&v, false); err != nil {
		return fmt.Errorf("toggle sesame: %w", err)
	}

	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	f, err := ss.ExpandFilename(file)
	if err != nil {
		return err
	}

	if err := os.WriteFile(f, data, os.ModePerm); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	return nil
}

func ReadJSONFile[T any](file string, v *T, opts ...JSONFileOptionFunc) (*T, error) {
	opt := &JSONFileOption{}
	for _, o := range opts {
		o(opt)
	}

	f, err := ss.ExpandFilename(file)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(f)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", file, err)
	}

	if err := json.Unmarshal(data, v); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", file, err)
	}

	if err := opt.ToggleSesame(v, true); err != nil {
		return nil, fmt.Errorf("toggle sesame: %w", err)
	}

	return v, nil
}

type JSONFileOption struct {
	Sesame string
}

type JSONFileOptionFunc func(opts *JSONFileOption)

func WithSesame(sesame string) JSONFileOptionFunc {
	return func(opts *JSONFileOption) {
		opts.Sesame = sesame
	}
}

func (o *JSONFileOption) ToggleSesame(obj any, open bool) error {
	if o.Sesame == "" {
		return nil
	}

	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("obj must be a pointer to struct")
	}

	v = v.Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}

		if field.Kind() != reflect.String {
			continue
		}

		tag, ok := t.Field(i).Tag.Lookup("sesame")
		if !ok || tag == "-" {
			continue
		}

		str := field.String()
		if str == "" {
			continue
		}

		f := Encrypt
		if open {
			f = Decrypt
		}

		val, err := f([]byte(o.Sesame), []byte(t.Field(i).Name), []byte("Sesame"), str)
		if err != nil {
			return fmt.Errorf("encrypt field %s: %w", t.Field(i).Name, err)
		}

		if str != val {
			field.SetString(val)
		}
	}

	return nil
}

func Encrypt(secret, salt, info []byte, data string) (string, error) {
	// Let's start up our key factory
	keyFactory := hkdf.New(sha256.New, secret, salt, info)

	// Now, let's produce two 32-byte keys
	key1 := make([]byte, 32)

	if _, err := io.ReadFull(keyFactory, key1); err != nil {
		return "", fmt.Errorf("create key: %w", err)
	}

	block, err := aes.NewCipher(key1)
	if err != nil {
		return "", fmt.Errorf("AES: %w", err)
	}

	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("rand.Read: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("cipher.NewGCM: %w", err)
	}

	encrypted := aesgcm.Seal(nil, nonce, []byte(data), nil)
	encrypted = append(nonce, encrypted...)

	return "{Sesame}" + base64.RawURLEncoding.EncodeToString(encrypted), nil
}

func Decrypt(secret, salt, info []byte, data string) (string, error) {
	if !strings.HasPrefix(data, "{Sesame}") {
		return data, nil
	}

	data = data[len("{Sesame}"):]
	if len(data) <= 12 {
		return "", fmt.Errorf("data too short")
	}

	// Let's start up our key factory
	keyFactory := hkdf.New(sha256.New, secret, salt, info)

	// Now, let's produce two 32-byte keys
	key1 := make([]byte, 32)

	if _, err := io.ReadFull(keyFactory, key1); err != nil {
		return "", fmt.Errorf("create key: %w", err)
	}

	block, err := aes.NewCipher(key1)
	if err != nil {
		return "", fmt.Errorf("AES: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("cipher.NewGCM: %w", err)
	}

	encrypted, err := base64.RawURLEncoding.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("base64.StdEncoding.DecodeString: %w", err)
	}

	nonce := encrypted[:12]
	encrypted = encrypted[12:]
	decrypted, err := aesgcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return "", fmt.Errorf("aesgcm.Open: %w", err)
	}

	return string(decrypted), nil
}
