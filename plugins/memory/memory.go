package memory

import (
	"container/list"
	"sync"
	"time"

	"github.com/oshankfriends/session"
)

var pDer = &Provider{sessionList: list.New(), sessions: make(map[string]*list.Element)}

type SessionStore struct {
	sID        string
	accsssTime time.Time
	value      map[interface{}]interface{}
}

func (s *SessionStore) Set(key interface{}, value interface{}) error {
	return nil
}

func (s *SessionStore) Get(key interface{}) interface{} {
	return struct{}{}
}

func (s *SessionStore) Delete(key interface{}) error {
	return nil
}

func (s *SessionStore) SessionID() (sid string) {
	return
}

type Provider struct {
	sync.Mutex
	sessionList *list.List
	sessions    map[string]*list.Element
}

func (p *Provider) SessionInit(sid string) (session.Session, error) {
	p.Lock()
	defer p.Unlock()
	newSession := &SessionStore{sID: sid, accsssTime: time.Now(), value: make(map[interface{}]interface{})}
	e := p.sessionList.PushBack(newSession)
	p.sessions[sid] = e
	return newSession, nil
}

func (p *Provider) SessionRead(sid string) (session.Session, error) {
	if sessionEle, ok := p.sessions[sid]; ok {
		return sessionEle.Value.(session.Session), nil
	}
	return p.SessionInit(sid)
}

func (p *Provider) SessionDestroy(sid string) error {
	if sessionEle, ok := p.sessions[sid]; ok {
		p.sessionList.Remove(sessionEle)
		delete(p.sessions, sid)
	}
	return nil
}

func (p *Provider) SessionUpdate(sid string) error {
	if sessionEle, ok := p.sessions[sid]; ok {
		sessionEle.Value.(*SessionStore).accsssTime = time.Now()
		p.sessionList.MoveToFront(sessionEle)
	}
	return nil
}

func (p *Provider) SessionGC(maxAge time.Duration) {
	p.Lock()
	defer p.Unlock()
	for {
		seesionEle := p.sessionList.Back()
		if seesionEle == nil {
			break
		}
		if seesionEle.Value.(*SessionStore).accsssTime.Unix()+int64(maxAge) < time.Now().Unix() {
			p.sessionList.Remove(seesionEle)
			delete(p.sessions, seesionEle.Value.(SessionStore).sID)
		} else {
			break
		}
	}
}

func init() {
	session.InstallProviderPlugin("memory", pDer)
}
