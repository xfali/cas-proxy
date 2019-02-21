/**
 * Copyright (C) 2019, Xiongfa Li.
 * All right reserved.
 * @author xiongfa.li
 * @date 2019/2/19
 * @time 16:28
 * @version V1.0
 * Description: 
 */

package proxy

import (
    "io"
    "net/http"
    "strings"
)

func DoProxy(remote_addr string, token string, w http.ResponseWriter, req *http.Request) {
    cli := &http.Client{}
    body := make([]byte, 0)
    n, err := io.ReadFull(req.Body, body)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        io.WriteString(w, "Request Data Error")
        return
    }
    reqUrl := "http://" + remote_addr + req.URL.Path

    req2, err := http.NewRequest(req.Method, reqUrl, strings.NewReader(string(body)))
    if err != nil {
        io.WriteString(w, "Request Error")
        return
    }
    // set request content type
    for k, v := range req.Header {
        req2.Header[k] = v
    }
    //del cookie
    req2.Header.Del("Cookie")
    //set token
    req2.Header.Set("X-AUTH-TOKEN", token)
    //contentType := req.Header.Get("Content-Type")
    //req2.Header.Set("Content-Type", contentType)
    // request
    rep2, err := cli.Do(req2)
    if err != nil {
        w.WriteHeader(http.StatusBadGateway)
        io.WriteString(w, "Not Found!")
        return
    }
    defer rep2.Body.Close()
    n, err = io.ReadFull(rep2.Body, body)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        io.WriteString(w, "Request Error")
        return
    }
    // set response header
    for k, v := range rep2.Header {
        w.Header().Set(k, v[0])
    }

    io.WriteString(w, string(body[:n]))
}
