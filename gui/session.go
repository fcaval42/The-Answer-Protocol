// Pb de fond: client main go a une seule connexion (readServerMessages) qui lit en boucle
// tout ce qui arrive du serveur et imprime au terminal. Si le GUI envoie "ATTACK" et veut
// connaître la réponse (Ok/ERR) = imprimée dans le terminal comme n'importe quel message
// et GUI voit pas. On va donc la détourner vers le code GUI.
// Ici, session est ce mécanisme : une petite boîte aux lettres partagée entre la boucle
// de lecture (qui reçoit) et le clic du bouton GUI (qui attend).

package gui

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// sendCooldown : apres un Send(), tout nouvel appel dans cette fenetre est
// ignore silencieusement (double-clic, double-declenchement d'un tap Fyne...).
const sendCooldown = 400 * time.Millisecond

type Session struct {
	conn     net.Conn    // connexion avec le serveur
	busy     atomic.Bool // true tant qu'un Send() est en cours (envoi + attente de sa reponse)
	awaiting atomic.Bool // Indique si on attend une réponse à une commande envoyée (bool)
	pending  chan string // canal pour recevoir la réponse du serveur

	cooldownMu  sync.Mutex
	nextAllowed time.Time // prochain instant a partir duquel un nouveau Send() est accepte

	rawMu    sync.Mutex
	rawChan  chan string // non-nil pendant une session brute (chat/groupe), cf. EnterRaw
	rawLock  sync.Mutex  // mutex pour sérialiser EnterRaw (évite chevauchements)

	inviteMu sync.Mutex
	onInvite func(from string) // cf. OnInvite

	questMu     sync.Mutex
	onQuestComplete func(questName, completer string) // cf. OnQuestComplete
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


// initialiser nouvelle session
func NewSession(conn net.Conn) *Session {
	return &Session{
		conn: conn,
		pending: make(chan string, 1),
	}
}


// à chaque ligne reçue du serveur, readServerMessages demande d'abord à Deliver :
// "est-ce que quelqu'un attend cette ligne ?" = awaiting = vrai = pousse dans pending
// qui réveille le Send qui était bloqué dessus.
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
			// Check for quest completion notification
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

	// Check for quest completion notification
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
	// Envoyer QUIT exactement comme le CLI le fait
	fmt.Fprintf(s.conn, "%s", "QUIT")
	fmt.Fprint(s.conn, "[INFO] Fermeture de la connexion...\n")
	// Ne pas fermer la connexion ici - laisser le defer de client_main.go le faire
}

func (s *Session) Close() {
	s.conn.Close()
}
