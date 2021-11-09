package main

import (
	"bytes"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os/exec"
	"time"
)

var c *config

type config struct {
	AccessKeyId  string `yaml:"access_key_id"`
	AccessSecret string `yaml:"access_secret"`
	DnsDomain    string `yaml:"dns_domain"`
	AliyunDomain string `yaml:"aliyun_domain"`
	CurlDomain   string `yaml:"curl_domain"`
}

func init() {
	yamlFile, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		log.Fatalln(err)
		return
	}

	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalln(err)
		return
	}
}

func GetWanIP() (wanIp string) {
	var result bytes.Buffer
	cmd := exec.Command("curl", c.CurlDomain)
	cmd.Stdout = &result
	err := cmd.Run()
	if err != nil {
		log.Fatalln(err)
	}

	wanIp = result.String()[:len(result.String())-1]
	return wanIp
}

func CreateNewAliDns() {
	var request *alidns.AddDomainRecordRequest
	client, err := alidns.NewClientWithAccessKey("cn-hangzhou", c.AccessKeyId, c.AccessSecret)

	request = alidns.CreateAddDomainRecordRequest()
	request.Scheme = "https"
	request.Value = GetWanIP()
	request.Type = "A"
	request.RR = c.DnsDomain
	request.DomainName = c.AliyunDomain

	if client != nil {
		_, err = client.AddDomainRecord(request)
	}

	if err != nil {
		log.Println(err)
		return
	}
}

func GetAliIpAndRecordId() (AliIp, RecordId string) {
	var request *alidns.DescribeSubDomainRecordsRequest
	var response *alidns.DescribeSubDomainRecordsResponse
	client, err := alidns.NewClientWithAccessKey("cn-hangzhou", c.AccessKeyId, c.AccessSecret)

	request = alidns.CreateDescribeSubDomainRecordsRequest()
	request.Scheme = "https"
	domain := c.DnsDomain + "." + c.AliyunDomain
	request.SubDomain = domain

	if client != nil {
		response, err = client.DescribeSubDomainRecords(request)
	}

	if err != nil {
		log.Println(err)
		return "", ""
	}

	if response != nil {
		if response.IsSuccess() {
			for _, v := range response.DomainRecords.Record {
				if v.RR == c.DnsDomain {
					return v.Value, v.RecordId
				}
			}
		}
	}

	return "", ""
}

func UpdateDNS(recordId string) error {
	var request *alidns.UpdateDomainRecordRequest
	var response *alidns.UpdateDomainRecordResponse
	client, err := alidns.NewClientWithAccessKey("cn-hangzhou", c.AccessKeyId, c.AccessSecret)

	request = alidns.CreateUpdateDomainRecordRequest()
	request.Scheme = "https"

	request.RecordId = recordId
	request.RR = c.DnsDomain
	request.Type = "A"
	request.Value = GetWanIP()
	request.Lang = "en"
	request.UserClientIp = GetWanIP()
	request.TTL = "600"
	request.Priority = "1"
	request.Line = "default"

	if client != nil {
		response, err = client.UpdateDomainRecord(request)
		log.Println(response)
	}

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func SetDns() {
	AliIp, RecordId := GetAliIpAndRecordId()
	wanIp := GetWanIP()
	if AliIp != wanIp {
		err := UpdateDNS(RecordId)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func main() {
	AliIp, RecordId := GetAliIpAndRecordId()
	if AliIp == "" && RecordId == "" {
		CreateNewAliDns()
	}

	for {
		go SetDns()
		time.Sleep(time.Hour * 1)
	}
}
