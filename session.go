package session

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Manager struct {
	sync.Mutex
	cookieName string
	maxAge     time.Duration
	provider   Provider
}

func NewManager(cookieName, providerName string, maxAge time.Duration) (*Manager, error) {
	provider, ok := providerPlugins[providerName]
	if !ok {
		return nil, fmt.Errorf("%s provider does not exist", providerName)
	}
	return &Manager{cookieName: cookieName, maxAge: maxAge, provider: provider}, nil
}

//SessionId returns a unique session ID for users
func (m *Manager) SessionID() string {
	var b = make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

//StartSession will return an existing session for the user if exist or create a new one
func (m *Manager) StartSession(w http.ResponseWriter, r *http.Request) (session Session, err error) {
	m.Lock()
	defer m.Unlock()
	cookie, err := r.Cookie(m.cookieName)
	if err != nil || cookie.Value == "" {
		sid := m.SessionID()
		session, err = m.provider.SessionInit(url.QueryEscape(sid))
		if err != nil {
			return
		}
		cookie := &http.Cookie{Name: m.cookieName, Value: url.QueryEscape(sid), MaxAge: int(m.maxAge), HttpOnly: true}
		http.SetCookie(w, cookie)
	} else {
		session, err = m.provider.SessionRead(url.QueryEscape(cookie.Value))
		if err != nil {
			return
		}

	}
	return
}

//DestroySession will Delete a session corresponds to a user
func (m *Manager) DestroySession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil {
		return
	}
	m.Lock()
	defer m.Unlock()
	m.provider.SessionDestroy(url.QueryEscape(cookie.Value))
	cookie = &http.Cookie{Name: m.cookieName, Expires: time.Now(), MaxAge: -1, HttpOnly: true}
	http.SetCookie(w, cookie)
}

func (m *Manager) GC() {
	m.Lock()
	defer m.Unlock()
	m.provider.SessionGC(m.maxAge)
	time.AfterFunc(m.maxAge, func() { m.GC() })
}

type Session interface {
	Set(key, value interface{}) error
	Get(key interface{}) interface{}
	Delete(key interface{}) error
	SessionID() string
}

var providerPlugins = make(map[string]Provider)

type Provider interface {
	//SessionInit will Initialize the Session and returns a new Session
	SessionInit(sid string) (Session, error)
	//SessionRead will return a Session by corresponding Session id or return a new one if not exist
	SessionRead(sid string) (Session, error)
	//SessionDestroy will delete a session by given session id
	SessionDestroy(sid string) error
	//SessionGC will delete all the expired Session based on given maxAge
	SessionGC(maxAge time.Duration)
	//SessionUpdate will update an existing session corresponds to given session id
	SessionUpdate(sid string) error
}

func InstallProviderPlugin(name string, provider Provider) {
	if provider == nil {
		return
	}

	if _, ok := providerPlugins[name]; ok {
		log.Printf("%s provider already exist", name)
		return
	}

	providerPlugins[name] = provider
}
