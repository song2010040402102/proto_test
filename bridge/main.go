package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/astaxie/beego/logs"

	"proto_test/config"
	"proto_test/protobuf"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		ret := doLogin(r.Form.Get("server"), r.Form.Get("loginname"))
		data, _ := json.Marshal(ret)
		fmt.Fprint(w, string(data))
	} else {
		fmt.Fprint(w, r.Method, "not support!")
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		ret := doLogout(r.Form.Get("server"), r.Form.Get("loginname"))
		data, _ := json.Marshal(ret)
		fmt.Fprint(w, string(data))
	} else {
		fmt.Fprint(w, r.Method, "not support!")
	}
}

func protoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" {
		fmt.Fprint(w, r.Method, "not support!")
		return
	}
	var ret string
	r.ParseForm()
	msgId, _ := strconv.Atoi(r.Form.Get("msg_id"))
	if r.Method == "GET" {
		if msgId == 0 {
			type IdName struct {
				Id   int32  `json:"id"`
				Name string `json:"name"`
			}
			var idNames []*IdName
			all := protobuf.GetAllMsgType()
			for k, v := range all {
				if v == "C2S_KEEP_ALIVE" || v == "C2S_LOGIN" {
					continue
				}
				s := strings.ToUpper(v)
				if s[:3] == "CS_" || s[:4] == "C_S_" || s[:4] == "C2S_" || s[:6] == "C_2_S_" {
					idNames = append(idNames, &IdName{Id: k, Name: v})
				}
			}
			data, _ := json.Marshal(idNames)
			ret = string(data)

		} else {
			ret = protobuf.GetJsonByMsgId(int32(msgId))
		}
	} else {
		data, _ := json.Marshal(doSendProto(r.Form.Get("server"), r.Form.Get("loginname"), int32(msgId), r.Form.Get("js")))
		ret = string(data)
	}
	fmt.Fprint(w, ret)
}

func logHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		r.ParseForm()
		data, _ := json.Marshal(getLog(r.Form.Get("server"), r.Form.Get("loginname")))
		fmt.Fprint(w, string(data))
	} else {
		fmt.Fprint(w, r.Method, "not support!")
	}
}

func init() {
	if !protobuf.ParseProto(config.Get().Protocol) {
		logs.Error("parse proto error!")
		os.Exit(0)
	}
}

func main() {
	logs.Notice("server start...")
	http.Handle("/", http.FileServer(http.Dir(config.Get().Bridge.Web)))
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/proto", protoHandler)
	http.HandleFunc("/log", logHandler)
	go http.ListenAndServe(config.Get().Bridge.Listen, nil)
	var sig os.Signal
	c := make(chan os.Signal, 1)
	for {
		signal.Notify(c)
		sig = <-c
		if sig != syscall.SIGPIPE && sig != syscall.SIGCHLD && sig != syscall.SIGURG {
			break
		}
	}
	logs.Notice("server terminate with sig:", sig)
}
