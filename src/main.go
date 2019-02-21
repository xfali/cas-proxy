/**
 * Copyright (C) 2019, Xiongfa Li.
 * All right reserved.
 * @author xiongfa.li
 * @date 2019/2/19
 * @time 16:20
 * @version V1.0
 * Description: 
 */

package main

import (
    "cas-proxy/src/cas"
    "cas-proxy/src/proxy"
    "cas-proxy/src/session"
    "cas-proxy/src/session/memory"
    "flag"
    "fmt"
    "log"
    "net/http"
)

var globalSessions *session.Manager

func init() {
    session.Register("memory", memory.New())
    globalSessions, _ = session.NewSessionManager("memory", "GOSESSIONID", 3600)
    go globalSessions.GC()
}

func main() {
    var err error = nil
    remoteAddr := flag.String("h", "localhost", "代理的后端服务地址")
    port := flag.String("p", "12345", "代理运行端口号")
    casServer := flag.String("s", "localhost", "CAS-Server地址")

    flag.Parse()

    cmd := flag.Arg(0)

    fmt.Printf("action   : %s\n", cmd)
    fmt.Printf("remote addr: %s\n", *remoteAddr)
    fmt.Printf("port: %s\n", *port)
    fmt.Printf("cas-server addr: %s\n", *casServer)

    http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
        check := true
        cookie, err := req.Cookie("JSESSIONID")
        if err != nil {
            check = false
        }
        if check {
            sess := globalSessions.SessionStart(w, req)
            token := sess.Get(cookie.Value)
            if token == nil {
                check = false
            } else {
                //success
                proxy.DoProxy(*remoteAddr, token.(string), w, req)
            }
        }

        if !check && cas.IsAuthentication(w, req, *casServer) {
            http.SetCookie(w, &http.Cookie{
                Name:  "JSESSIONID",
                Value: "fake_id",
            })
            sess := globalSessions.SessionStart(w, req)
            sess.Set("fake_id", "TGC")
            http.Redirect(w, req, cas.ServiceUrl(req), http.StatusFound)
        }
    })
    err = http.ListenAndServe(":"+*port, nil)
    if err != nil {
        log.Fatal("server down!!!")
    }
}
