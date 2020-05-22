package dns

import (
	"encoding/json"
	"github.com/miekg/dns"
	"log"
	"net"
	"strings"
)

// 获取域名的TXT记录(n个)的JSON字符串
func GetTxt(domain string) string {
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeTXT)
	m.RecursionDesired = true
	r, _, err := c.Exchange(m, net.JoinHostPort("223.5.5.5", "53"))
	if r == nil {
		log.Println(err)
		return ""
	}
	if r.Rcode != dns.RcodeSuccess {
		log.Println("获取结果失败", r.Rcode)
		return ""
	}

	resultMap := make(map[string]string)
	for _, a := range r.Answer {
		txt := strings.ReplaceAll(a.String(), a.Header().String(), "")
		txt = txt[1 : len(txt)-1]
		kv := strings.Split(txt, "=")
		resultMap[kv[0]] = kv[1]
	}
	jsonBytes, _ := json.Marshal(resultMap)
	return string(jsonBytes)
}
