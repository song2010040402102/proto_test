package session

import (
	"net"
	"sync"
	"time"

	"github.com/astaxie/beego/logs"
)

const (
	MAX_WRITE_CACHE int = 32
	READ_BUFF_SIZE  int = 1024
	WRITE_BUFF_SIZE int = 1024
	READ_TIMEOUT    int = 120
	WRITE_TIMEOUT   int = 60
)

type Session struct {
	conn               *net.TCPConn
	serv               bool
	Loginname          string
	Data 		   		interface{}
	readCallback       func(*Session, []byte)
	disconnectCallback func(*Session)
	writeCache         chan []byte
}

func NewSession(conn *net.TCPConn, serv bool) *Session {
	conn.SetReadBuffer(READ_BUFF_SIZE)
	conn.SetWriteBuffer(WRITE_BUFF_SIZE)
	conn.SetNoDelay(false)
	return &Session{
		conn:       conn,
		serv:       serv,
		writeCache: make(chan []byte, MAX_WRITE_CACHE),
	}
}

func (cs *Session) String() string {
	return "addr: " + cs.conn.RemoteAddr().String() + " Loginname: " + cs.Loginname
}

func (cs *Session) Addr() string {
	return cs.conn.RemoteAddr().String()
}

func (cs *Session) SetReadCallback(cb func(*Session, []byte)) {
	cs.readCallback = cb
}

func (cs *Session) SetDisconnectCallback(cb func(*Session)) {
	cs.disconnectCallback = cb
}

func (cs *Session) Run() {
	go cs.read()
	go cs.write()
}

func (cs *Session) Send(buff []byte) {
	cs.writeCache <- buff
}

func (cs *Session) Close() {
	cs.conn.Close()
	cs.disconnectCallback(cs)
	GetManager().RemoveSession(cs)
}

func (cs *Session) read() {
	buff := make([]byte, 0, READ_BUFF_SIZE)
	sum, size := uint32(0), uint32(0)
	for {
		var data []byte
		b := make([]byte, READ_BUFF_SIZE)
		cs.conn.SetReadDeadline(time.Now().Add(time.Duration(READ_TIMEOUT) * time.Second))
		n, err := cs.conn.Read(b)
		if err != nil {
			logs.Error("session read", n, err)
			cs.Close()
			return
		}
		sum += uint32(n)
		if size == 0 {
			size = uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
		}
		buff = append(buff, b...)
		if sum == size+4 {
			data = buff[4 : size+4]
			sum, size = 0, 0
			buff = buff[:0]
			cs.readCallback(cs, data)
		} else {
			for sum >= size+8 {
				data = buff[4 : size+4]
				sum -= size + 4
				size = uint32(buff[size+4]) | uint32(buff[size+5])<<8 | uint32(buff[size+6])<<16 | uint32(buff[size+7])<<24
				buff = buff[size+4:]
				cs.readCallback(cs, data)
			}
		}
	}
}

func (cs *Session) write() {
	for {
		buff := make([]byte, 4, WRITE_BUFF_SIZE)
		for {
			b := <-cs.writeCache
			if len(b) == 0 {
				continue
			}
			size := uint32(len(b))
			buff[0] = byte(size & 0xff)
			buff[1] = byte(size >> 8 & 0xff)
			buff[2] = byte(size >> 16 & 0xff)
			buff[3] = byte(size >> 24 & 0xff)
			buff = append(buff, b...)
			if len(cs.writeCache) == 0 || len(buff) >= WRITE_BUFF_SIZE {
				break
			}
		}
		for i := 0; i < len(buff); {
			j := i + WRITE_BUFF_SIZE
			if j > len(buff) {
				j = len(buff)
			}
			cs.conn.SetWriteDeadline(time.Now().Add(time.Duration(WRITE_TIMEOUT) * time.Second))
			n, err := cs.conn.Write(buff[i:j])
			if err != nil {
				logs.Error("session write", n, err)
				cs.Close()
				return
			}
			i += n
		}
	}
}

type SessionManager struct {
	lock       sync.RWMutex
	mapSession map[string]*Session
}

func (sm *SessionManager) GetSession(key string) *Session {
	sm.lock.RLock()
	defer sm.lock.RUnlock()
	return sm.mapSession[key]
}

func (sm *SessionManager) AddSession(sess *Session) {
	if sess == nil {
		return
	}
	sm.lock.Lock()
	defer sm.lock.Unlock()
	key := sess.Loginname
	if !sess.serv {
		key = sess.Addr() + key
	}
	if _, ok := sm.mapSession[key]; ok {
		return
	}
	sm.mapSession[key] = sess
}

func (sm *SessionManager) RemoveSession(sess *Session) {
	if sess == nil {
		return
	}
	sm.lock.Lock()
	defer sm.lock.Unlock()
	key := sess.Loginname
	if !sess.serv {
		key = sess.Addr() + key
	}
	delete(sm.mapSession, key)
}

func GetManager() *SessionManager {
	return g_sessionMan
}

func init() {
	g_sessionMan = &SessionManager{
		mapSession: make(map[string]*Session),
	}
}

var g_sessionMan *SessionManager
