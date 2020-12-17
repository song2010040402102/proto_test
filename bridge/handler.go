package main

import (
	"github.com/astaxie/beego/logs"

	"proto_test/protobuf/protocol"
	"proto_test/session"
)

func handleKeepAlive(sess *session.Session, res *protocol.S2C_KeepAlive) {
	logs.Info("handleKeepAlive", res)
}

func handleLogin(sess *session.Session, res *protocol.S2C_Login) {
	logs.Info("handleLogin", res)
	if res.GetRet() == 0 {
		sess.Loginname = res.GetLoginname()
		session.GetManager().AddSession(sess)
	} else {
		sess.Close()
	}
}
