package main

import (
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/golang/protobuf/proto"

	"proto_test/protobuf"
	"proto_test/protobuf/protocol"
	"proto_test/session"
	"proto_test/util"
)

const (
	ERROR_HAVE_LOGIN   int32 = 1
	ERROR_NOT_LOGIN    int32 = 2
	ERROR_CONNECT_FAIL int32 = 3
	ERROR_INVALID_JSON int32 = 4
)

type RetMsg struct {
	Ret int32  `json:"ret"`
	Msg string `json:"msg,omitempty"`
}

type ProtoLog struct {
	Ts     string `json:"ts"`
	Dir    string `json:"dir"`
	Id     int32  `json:"id"`
	Msg    string `json:"msg"`
	Detail string `json:"detail"`
}

func NewProtoLog(dir bool, id int32, msg proto.Message) *ProtoLog {
	log := &ProtoLog{
		Id: id,
	}
	ts := time.Now().UnixNano()
	log.Ts = time.Unix(ts/1e9, 0).Format("2006-01-02 15:04:05") + "." + strconv.Itoa(int(ts%1e9/1e3))
	if dir {
		log.Dir = "send"
	} else {
		log.Dir = "recv"
	}
	log.Msg = protobuf.GetMessageNameById(id)
	log.Detail = msg.String()
	return log
}

type RetLog struct {
	RetMsg
	Logs []*ProtoLog `json:"logs"`
}

type UserData struct {
	lock    sync.Mutex
	tick    *time.Ticker
	webTick *time.Ticker
	webReq  bool
	logs    []*ProtoLog
}

func NewUserData() *UserData {
	return &UserData{
		webReq: false,
		logs:   make([]*ProtoLog, 0, 16),
	}
}

func (ud *UserData) PushLog(log *ProtoLog) {
	ud.lock.Lock()
	defer ud.lock.Unlock()
	ud.logs = append(ud.logs, log)
}

func (ud *UserData) PopAllLog() []*ProtoLog {
	ud.lock.Lock()
	defer ud.lock.Unlock()
	ret := ud.logs
	ud.logs = make([]*ProtoLog, 0, 16)
	return ret
}

func onMessage(sess *session.Session, buff []byte) {
	logs.Info("onMessage", sess, len(buff))
	t, msg := util.Buff2ProtoMsg(buff)
	switch t {
	case int32(protocol.ProtocolType_S2C_KEEP_ALIVE):
		handleKeepAlive(sess, msg.(*protocol.S2C_KeepAlive))
	case int32(protocol.ProtocolType_S2C_LOGIN):
		handleLogin(sess, msg.(*protocol.S2C_Login))
	}
	sess.Data.(*UserData).PushLog(NewProtoLog(false, t, msg))
}

func onDisconnect(sess *session.Session) {
	logs.Info("onDisconnect", sess)
	sess.Data.(*UserData).tick.Stop()
	sess.Data.(*UserData).webTick.Stop()
}

func doConnect(server string) *session.Session {
	addr, err := net.ResolveTCPAddr("tcp4", server)
	if err != nil {
		logs.Error("ResolveTCPAddr", err)
		return nil
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		logs.Error("DialTCP", err)
		return nil
	}
	sess := session.NewSession(conn, false)
	sess.SetReadCallback(onMessage)
	sess.SetDisconnectCallback(onDisconnect)
	sess.Run()
	return sess
}

func doLogin(server, loginname string) *RetMsg {
	res := &RetMsg{}
	sess := session.GetManager().GetSession(server + loginname)
	if sess != nil {
		res.Ret = ERROR_HAVE_LOGIN
		return res
	}
	sess = doConnect(server)
	if sess == nil {
		res.Ret = ERROR_CONNECT_FAIL
		return res
	}
	sess.Data = NewUserData()
	go func() {
		sess.Data.(*UserData).tick = time.NewTicker(60 * time.Second)
		for range sess.Data.(*UserData).tick.C {
			msg := &protocol.C2S_KeepAlive{}
			sess.Data.(*UserData).PushLog(NewProtoLog(true, int32(msg.GetType()), msg))
			sess.Send(util.ProtoMsg2Buff(int32(msg.GetType()), msg))
		}
	}()
	go func() {
		sess.Data.(*UserData).webTick = time.NewTicker(3 * time.Second)
		for range sess.Data.(*UserData).webTick.C {
			if !sess.Data.(*UserData).webReq {
				sess.Close()
				break
			}
			sess.Data.(*UserData).webReq = false
		}
	}()
	msg := &protocol.C2S_Login{
		Loginname: proto.String(loginname),
	}
	sess.Data.(*UserData).PushLog(NewProtoLog(true, int32(msg.GetType()), msg))
	sess.Send(util.ProtoMsg2Buff(int32(msg.GetType()), msg))
	return res
}

func doLogout(server, loginname string) *RetMsg {
	res := &RetMsg{}
	sess := session.GetManager().GetSession(server + loginname)
	if sess == nil {
		res.Ret = ERROR_NOT_LOGIN
		return res
	}
	sess.Close()
	return res
}

func doSendProto(server, loginname string, msg_id int32, js string) *RetMsg {
	res := &RetMsg{}
	sess := session.GetManager().GetSession(server + loginname)
	if sess == nil {
		res.Ret = ERROR_NOT_LOGIN
		return res
	}
	msgObj := protobuf.Json2Message(msg_id, js)
	if msgObj == nil {
		res.Ret = ERROR_INVALID_JSON
		return res
	}
	sess.Data.(*UserData).PushLog(NewProtoLog(true, msg_id, msgObj))
	sess.Send(util.ProtoMsg2Buff(msg_id, msgObj))
	return res
}

func getLog(server, loginname string) *RetLog {
	res := &RetLog{}
	sess := session.GetManager().GetSession(server + loginname)
	if sess == nil {
		res.Ret = ERROR_NOT_LOGIN
		return res
	}
	ud := sess.Data.(*UserData)
	ud.webReq = true
	res.Logs = ud.PopAllLog()
	return res
}
