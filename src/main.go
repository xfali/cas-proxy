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
    "sync"
)

var globalSessions *session.Manager
var sessionidMap map[string]string
var globalLock sync.Mutex

func init() {
    session.Register("memory", memory.New())
    globalSessions, _ = session.NewSessionManager("memory", "JSESSIONID", 3600)
    sessionidMap = map[string]string{}
    go globalSessions.GC()
}

func main() {
    var err error = nil
    remoteAddr := flag.String("h", "localhost", "代理的后端服务地址")
    port := flag.String("p", "12345", "代理运行端口号")
    casServer := flag.String("s", "localhost", "CAS-Server地址")
    path := flag.String("e", "/", "入口地址")

    flag.Parse()

    cmd := flag.Arg(0)

    fmt.Printf("action   : %s\n", cmd)
    fmt.Printf("remote addr: %s\n", *remoteAddr)
    fmt.Printf("port: %s\n", *port)
    fmt.Printf("cas-server addr: %s\n", *casServer)

    http.HandleFunc(*path, func(w http.ResponseWriter, req *http.Request) {
        session := globalSessions.TryGetSession(w, req)
        if session != nil {
            token := session.Get("token")
            if token != nil {
                //success
                proxy.DoProxy(*remoteAddr, token.(string), w, req)
                return
            }
        }

        //check logout request
        if st, logout := cas.CheckLogout(req); logout {
            globalLock.Lock()
            if value, ok := sessionidMap[st]; ok {
                globalSessions.TryDestroySession(value)
                delete(sessionidMap, st)
            }
            globalLock.Unlock()
            w.WriteHeader(http.StatusOK)
            return
        }

        if cas.IsAuthentication(w, req, *casServer) {
            sess := globalSessions.SessionStart(w, req)
            url, ticket := cas.SeparateServiceUrlTicket(req)
            globalLock.Lock()
            sessionidMap[ticket] = sess.SessionID()
            globalLock.Unlock()
            sess.Set("token", "xx")
            http.Redirect(w, req, url, http.StatusFound)
        }
    })
    err = http.ListenAndServe(":"+*port, nil)
    if err != nil {
        log.Fatal("server down!!!")
    }
}
