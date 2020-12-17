package main

import (
	"net"
	"os"

	"github.com/astaxie/beego/logs"

	"proto_test/config"
	"proto_test/protobuf"
	"proto_test/protobuf/protocol"
	"proto_test/session"
	"proto_test/util"
)

func onConnect(sess *session.Session) {
	logs.Info("onConnect", sess)
}

func onMessage(sess *session.Session, buff []byte) {
	logs.Info("onMessage", sess, len(buff))
	t, msg := util.Buff2ProtoMsg(buff)
	switch t {
	case int32(protocol.ProtocolType_C2S_KEEP_ALIVE):
		handleKeepAlive(sess, msg.(*protocol.C2S_KeepAlive))
	case int32(protocol.ProtocolType_C2S_LOGIN):
		handleLogin(sess, msg.(*protocol.C2S_Login))
	case int32(protocol.ProtocolType_C2S_SEND_MSG):
		handleSendMsg(sess, msg.(*protocol.C2S_SendMsg))
	}
}

func onDisconnect(sess *session.Session) {
	logs.Info("onDisconnect", sess)
}

func init() {
	if !protobuf.ParseProto(config.Get().Protocol) {
		logs.Error("parse proto error!")
		os.Exit(0)
	}
}

func main() {
	addr, err := net.ResolveTCPAddr("tcp4", config.Get().Server.Listen)
	if err != nil {
		logs.Error("ResolveTCPAddr", err)
		return
	}
	listen, err := net.ListenTCP("tcp", addr)
	if err != nil {
		logs.Error("ListenTCP", err)
		return
	}
	for {
		conn, err := listen.AcceptTCP()
		if err != nil {
			logs.Error("AcceptTCP", err)
			continue
		}
		sess := session.NewSession(conn, true)
		sess.SetReadCallback(onMessage)
		sess.SetDisconnectCallback(onDisconnect)
		onConnect(sess)
		sess.Run()
	}
}
