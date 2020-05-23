package dns

import (
	"encoding/json"
	"github.com/miekg/dns"
	"log"
	"net"
	"strings"
)

// 获取记录
// 参数domain: 域名 lilu.red
// 参数t: 记录类型 https://pkg.go.dev/github.com/miekg/dns@v1.1.29?tab=doc#TypeNone
// 返回: 文本数组的JSON字符串
func DigShort(domain string, t int) string {
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), uint16(t))
	m.RecursionDesired = true
	r, _, err := c.Exchange(m, net.JoinHostPort("8.8.8.8", "53"))
	if r == nil {
		log.Println("查询DNS记录失败:", err)
		return "[]"
	}
	if r.Rcode != dns.RcodeSuccess {
		log.Println("查询DNS记录失败:", r.Rcode)
		return "[]"
	}

	var ta []string
	for _, a := range r.Answer {
		txt := strings.ReplaceAll(a.String(), a.Header().String(), "")
		ta = append(ta, txt)
	}
	jsonBytes, _ := json.Marshal(ta)
	result := string(jsonBytes)
	if result == "null" {
		result = "[]"
	}
	return result
}
