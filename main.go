package main

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/miekg/dns"
)

type Answer struct {
	RName string `json:"rname"`
	RType string `json:"rtype"`
	RData string `json:"rdata"`
}

type DnsResult struct {
	Query   string   `json:"query"`
	Answers []Answer `json:"answers"`
}

func (d *DnsResult) ToString() string {
	out, _ := json.Marshal(d)
	return string(out)
}

func dig(url string) *DnsResult {
	var rmsg *dns.Msg
	var err error

	result := &DnsResult{
		Query:   url,
		Answers: make([]Answer, 0),
	}

	c := dns.Client{
		Timeout: 5 * time.Second,
	}
	m := dns.Msg{}
	m.SetQuestion(dns.Fqdn(url), dns.TypeA)

	attempt := 1
	for {
		rmsg, _, err = c.Exchange(&m, "8.8.8.8:53")
		if err != nil {
			return nil
		}

		// Wait briefly before retrying
		time.Sleep(50 * time.Millisecond)
		attempt++
		if attempt > 3 {
			break
		}
	}

	for _, ans := range rmsg.Answer {
		answer := Answer{}

		switch record := ans.(type) {
		case *dns.A:
			answer.RName = record.Header().Name
			answer.RType = "A"
			answer.RData = record.A.String()
			result.Answers = append(result.Answers, answer)
		case *dns.CNAME:
			answer.RName = record.Header().Name
			answer.RType = "CNAME"
			answer.RData = record.Target
			result.Answers = append(result.Answers, answer)
		}
	}

	return result
}

func main() {
	r := gin.Default()
	r.GET("/dig/:url", func(c *gin.Context) {
		data := c.Param("url")
		c.JSON(200, dig(data))
	})
	r.Run("127.0.0.1:8001")
}
