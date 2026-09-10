package gui

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const sendCooldown = 400 * time.Millisecond

type Session struct {
	conn     net.Conn
	busy     atomic.Bool
	awaiting atomic.Bool
	pending  chan string

	cooldownMu  sync.Mutex
	nextAllowed time.Time

	rawMu   sync.Mutex
	rawChan chan string
	rawLock sync.Mutex

	inviteMu sync.Mutex
	onInvite func(from string)

	questMu         sync.Mutex
	onQuestComplete func(questName, completer string)

	disconnectMu sync.Mutex
	onDisconnect func()

	roomBroadcastMu sync.Mutex
	onRoomBroadcast func(roomJSON string)
}

const inviteMarker = " vous a invité à rejoindre un groupe. ACCEPT/DECLINE"

const questCompleteMarker = " a terminé la quête "

func parseInvite(line string) (string, bool) {
	const prefix = "[EVT] "
	if !strings.HasPrefix(line, prefix) {
		return "", false
	}
	rest := strings.TrimPrefix(line, prefix)
	idx := strings.Index(rest, inviteMarker)
	if idx <= 0 {
		return "", false
	}
	return rest[:idx], true
}

func parseQuestComplete(line string) (questName, completer string, ok bool) {
	const prefix = "[EVT] "
	if !strings.HasPrefix(line, prefix) {
		return "", "", false
	}
	rest := strings.TrimPrefix(line, prefix)
	idx := strings.Index(rest, questCompleteMarker)
	if idx <= 0 {
		return "", "", false
	}
	completer = strings.TrimSpace(rest[:idx])
	questName = strings.TrimSpace(rest[idx+len(questCompleteMarker):])
	questName = strings.TrimSuffix(questName, ".")
	questName = strings.TrimSpace(questName)
	return completer, questName, true
}

func (s *Session) OnInvite(f func(from string)) {
	s.inviteMu.Lock()
	s.onInvite = f
	s.inviteMu.Unlock()
}

func (s *Session) OnQuestComplete(f func(questName, completer string)) {
	s.questMu.Lock()
	s.onQuestComplete = f
	s.questMu.Unlock()
}

func (s *Session) OnDisconnect(f func()) {
	s.disconnectMu.Lock()
	s.onDisconnect = f
	s.disconnectMu.Unlock()
}

func (s *Session) NotifyDisconnect() {
	s.disconnectMu.Lock()
	h := s.onDisconnect
	s.disconnectMu.Unlock()
	if h != nil {
		h()
	}
}

func (s *Session) OnRoomBroadcast(f func(roomJSON string)) {
	s.roomBroadcastMu.Lock()
	s.onRoomBroadcast = f
	s.roomBroadcastMu.Unlock()
}

func (s *Session) NotifyRoomBroadcast(roomJSON string) {
	s.roomBroadcastMu.Lock()
	h := s.onRoomBroadcast
	s.roomBroadcastMu.Unlock()
	if h != nil {
		h(roomJSON)
	}
}

func NewSession(conn net.Conn) *Session {
	return &Session{
		conn:    conn,
		pending: make(chan string, 1),
	}
}

func (s *Session) Deliver(line string) bool {
	s.rawMu.Lock()
	raw := s.rawChan
	s.rawMu.Unlock()
	if raw != nil {
		select {
		case raw <- line:
		default:
			if from, ok := parseInvite(line); ok {
				s.inviteMu.Lock()
				h := s.onInvite
				s.inviteMu.Unlock()
				if h != nil {
					h(from)
					return true
				}
			}
			if completer, questName, ok := parseQuestComplete(line); ok {
				s.questMu.Lock()
				h := s.onQuestComplete
				s.questMu.Unlock()
				if h != nil {
					h(questName, completer)
					return true
				}
			}
			fmt.Println("[GUI] ligne de chat ignorée (canal plein ou fermé) :", line)
		}
		return true
	}

	if from, ok := parseInvite(line); ok {
		s.inviteMu.Lock()
		h := s.onInvite
		s.inviteMu.Unlock()
		if h != nil {
			h(from)
			return true
		}
	}

	if completer, questName, ok := parseQuestComplete(line); ok {
		s.questMu.Lock()
		h := s.onQuestComplete
		s.questMu.Unlock()
		if h != nil {
			h(questName, completer)
			return true
		}
	}

	if !s.awaiting.Load() {
		if strings.HasPrefix(line, "{") {
			s.NotifyRoomBroadcast(line)
		}
		return false
	}
	if !strings.HasPrefix(line, "OK ") && !strings.HasPrefix(line, "ERR") && !strings.HasPrefix(line, "{") {
		return false
	}
	select {
	case s.pending <- line:
	default:
		fmt.Println("[GUI] réponse en trop ignorée (pending déjà plein) :", line)
	}
	return true
}

func (s *Session) EnterRaw() <-chan string {
	s.rawLock.Lock()

	s.rawMu.Lock()
	if s.rawChan != nil {
		close(s.rawChan)
	}
	ch := make(chan string, 64)
	s.rawChan = ch
	s.rawMu.Unlock()

	return ch
}

func (s *Session) ExitRaw() {
	s.rawMu.Lock()
	if s.rawChan != nil {
		close(s.rawChan)
		s.rawChan = nil
	}
	s.rawMu.Unlock()

	s.rawLock.Unlock()
}

func (s *Session) SendRaw(line string) error {
	_, err := s.conn.Write([]byte(line))
	return err
}

func (s *Session) Send(cmd string) (string, error) {
	if !s.busy.CompareAndSwap(false, true) {
		return "", nil
	}
	defer s.busy.Store(false)

	s.cooldownMu.Lock()
	now := time.Now()
	if now.Before(s.nextAllowed) {
		s.cooldownMu.Unlock()
		return "", nil
	}
	s.nextAllowed = now.Add(sendCooldown)
	s.cooldownMu.Unlock()

	select {
	case <-s.pending:
	default:
	}

	s.awaiting.Store(true)
	defer s.awaiting.Store(false)

	if _, err := s.conn.Write([]byte(cmd + "\n")); err != nil {
		return "", err
	}

	select {
	case line := <-s.pending:
		return line, nil
	case <-time.After(5 * time.Second):
		return "", fmt.Errorf("timeout wainting for response to %q", cmd)
	}
}

func (s *Session) Quit() {
	fmt.Fprintf(s.conn, "%s", "QUIT")
	fmt.Fprint(s.conn, "[INFO] Fermeture de la connexion...\n")
}
