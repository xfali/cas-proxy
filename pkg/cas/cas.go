/**
 * Copyright (C) 2019, Xiongfa Li.
 * All right reserved.
 * @author xiongfa.li
 * @date 2019/2/19
 * @time 16:23
 * @version V1.0
 * Description:
 */

package cas

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	LogoutRequest   = "samlp:LogoutRequest"
	SesssionIDStart = "<samlp:SessionIndex>"
	SesssionIDEnd   = "</samlp:SessionIndex>"
)

func CheckLogout(req *http.Request) (string, bool) {
	if req.Method == http.MethodPost {
		body := make([]byte, 0)
		data, err := io.ReadFull(req.Body, body)
		if err != nil {
			return "", false
		}
		dataStr := string(data)
		if strings.Contains(dataStr, LogoutRequest) {
			start := strings.Index(dataStr, SesssionIDStart)
			end := strings.Index(dataStr, SesssionIDEnd)
			ST := dataStr[start+len(SesssionIDStart) : end]
			return ST, true
		}
	}
	return "", false
}

/*
判断当前访问是否已认证
*/
func IsAuthentication(w http.ResponseWriter, r *http.Request, casServerUrl string) bool {
	if !hasTicket(r) {
		redirectToCasServer(w, r, casServerUrl)
		return false
	}

	localUrl := getLocalUrl(r)
	if !validateTicket(localUrl, casServerUrl) {
		redirectToCasServer(w, r, casServerUrl)
		return false
	}
	return true
}

/*
重定向到CAS认证中心
*/
func redirectToCasServer(w http.ResponseWriter, r *http.Request, casServerUrl string) {
	service, ticket := SeparateServiceUrlTicket(r)
	if ticket != "" {
		ticket = "&ticket=" + ticket
	}
	casServerUrl = fmt.Sprintf("%s/login?service=%s%s", casServerUrl, url.QueryEscape(service), ticket)
	http.Redirect(w, r, casServerUrl, http.StatusFound)
}

/*
验证访问路径中的ticket是否有效
*/
func validateTicket(localUrl, casServerUrl string) bool {
	service, ticket := separateTicketParam(localUrl)
	casServerUrl = fmt.Sprintf("%s/serviceValidate?service=%s&ticket=%s", casServerUrl, url.QueryEscape(service), ticket)
	res, err := http.Get(casServerUrl)
	if err != nil {
		return false
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false
	}

	dataStr := string(data)
	if !strings.Contains(dataStr, "cas:authenticationSuccess") {
		return false
	}
	return true
}

/*
从请求中获取访问路径
*/
func SeparateServiceUrlTicket(r *http.Request) (string, string) {
	scheme := "http://"
	if r.TLS != nil {
		scheme = "https://"
	}
	url := strings.Join([]string{scheme, r.Host, r.RequestURI}, "")
	slice := strings.Split(url, "?")
	var ticket string
	if len(slice) > 1 {
		localUrl := slice[0]
		urlParamStr, t := separateTicketParam(slice[1])
		ticket = t
		url = localUrl + "?" + urlParamStr
	}
	return url, ticket
}

/*
处理并确保路径中只有一个ticket参数
*/
func separateTicketParam(urlParams string) (string, string) {
	if len(urlParams) == 0 || !strings.Contains(urlParams, "ticket") {
		return urlParams, ""
	}

	sep := "&"
	params := strings.Split(urlParams, sep)
	var ticket string
	newParams := ""
	for _, value := range params {
		if strings.Contains(value, "ticket") {
			ticket = strings.Split(value, "=")[1]
			continue
		}

		if len(value) == 0 {
			continue
		} else {
			if len(newParams) == 0 {
				newParams = value
			} else {
				newParams = newParams + sep + value
			}
		}
	}
	return newParams, ticket
}

/*
从请求中获取访问路径
*/
func getLocalUrl(r *http.Request) string {
	scheme := "http://"
	if r.TLS != nil {
		scheme = "https://"
	}
	url := strings.Join([]string{scheme, r.Host, r.RequestURI}, "")
	slice := strings.Split(url, "?")
	if len(slice) > 1 {
		localUrl := slice[0]
		urlParamStr := ensureOneTicketParam(slice[1])
		url = localUrl + "?" + urlParamStr
	}
	return url
}

/*
处理并确保路径中只有一个ticket参数
*/
func ensureOneTicketParam(urlParams string) string {
	if len(urlParams) == 0 || !strings.Contains(urlParams, "ticket") {
		return urlParams
	}

	sep := "&"
	params := strings.Split(urlParams, sep)

	newParams := ""
	ticket := ""
	for _, value := range params {
		if strings.Contains(value, "ticket") {
			ticket = value
			continue
		}

		if len(newParams) == 0 {
			newParams = value
		} else {
			newParams = newParams + sep + value
		}

	}
	newParams = newParams + sep + ticket
	return newParams
}

/*
获取ticket
*/
func getTicket(r *http.Request) string {
	return r.FormValue("ticket")
}

/*
判断是否有ticket
*/
func hasTicket(r *http.Request) bool {
	t := getTicket(r)
	return len(t) != 0
}
