package main

import (
	"github.com/astaxie/beego/logs"
	"github.com/golang/protobuf/proto"

	"proto_test/protobuf/protocol"
	"proto_test/session"
	"proto_test/util"
)

func handleKeepAlive(sess *session.Session, req *protocol.C2S_KeepAlive) {
	logs.Info("handleKeepAlive", req)
	res := &protocol.S2C_KeepAlive{}
	defer func() { sess.Send(util.ProtoMsg2Buff(int32(res.GetType()), res)) }()
}

func handleLogin(sess *session.Session, req *protocol.C2S_Login) {
	logs.Info("handleLogin", req)
	res := &protocol.S2C_Login{}
	defer func() { sess.Send(util.ProtoMsg2Buff(int32(res.GetType()), res)) }()

	if session.GetManager().GetSession(req.GetLoginname()) != nil {
		res.Ret = proto.Int32(int32(protocol.ProtocolError_HAVE_LOGIN))
		return
	}
	sess.Loginname = req.GetLoginname()
	session.GetManager().AddSession(sess)
	res.Loginname = proto.String(req.GetLoginname())
}

func handleSendMsg(sess *session.Session, req *protocol.C2S_SendMsg) {
	logs.Info("handleSendMsg", req)
	res := &protocol.S2C_SendMsg{}
	defer func() { sess.Send(util.ProtoMsg2Buff(int32(res.GetType()), res)) }()

	res.Ret = proto.Int32(0)
}
