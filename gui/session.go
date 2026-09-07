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

	rawMu   sync.Mutex
	rawChan chan string // non-nil pendant une session brute (chat/groupe), cf. EnterRaw
}


// initialiser nouvelle session
func NewSession(conn net.Conn) *Session {
	return &Session{conn: conn, pending: make(chan string, 1)}
}


// à chaque ligne reçue du serveur, readServerMessages demande d'abord à Deliver :
// "est-ce que quelqu'un attend cette ligne ?" = awaiting = vrai = pousse dans pending
// qui réveille le Send qui était bloqué dessus.
func (s *Session) Deliver(line string) bool {
	// Le chat et les groupes (CHAT, GROUP JOIN, GROUP INVITE...) ne renvoient
	// jamais de ligne "OK .../ERR .../{...}" : ce sont des messages "[SERVER]
	// ..." lus en continu par une boucle bloquante cote serveur (cf.
	// tcp/server/chat.go). Tant qu'une session brute est active (EnterRaw),
	// on lui route TOUTES les lignes plutot que de les laisser au filtre
	// OK/ERR/{ ci-dessous, qui ne les reconnaitrait jamais.
	s.rawMu.Lock()
	raw := s.rawChan
	s.rawMu.Unlock()
	if raw != nil {
		select {
		case raw <- line:
		default:
			fmt.Println("[GUI] ligne de chat ignorée (canal plein) :", line)
		}
		return true
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

// EnterRaw bascule la session en mode brut : toutes les lignes reçues du
// serveur sont désormais routées vers le channel retourné au lieu de passer
// par Send()/Deliver ou d'être affichées dans le terminal texte. Utilisé
// pour le chat et la gestion de groupe (cf. gui/chat_gui.go), dont le
// protocole reste celui, séquentiel et bloquant, du client texte.
func (s *Session) EnterRaw() <-chan string {
	ch := make(chan string, 64)
	s.rawMu.Lock()
	s.rawChan = ch
	s.rawMu.Unlock()
	return ch
}

// ExitRaw restaure le fonctionnement normal de Send()/Deliver. A appeler
// systematiquement en sortie de chat/groupe (y compris a la fermeture de la
// fenetre), sous peine de laisser toutes les reponses ulterieures du GUI
// partir dans le vide.
func (s *Session) ExitRaw() {
	s.rawMu.Lock()
	s.rawChan = nil
	s.rawMu.Unlock()
}

// SendRaw écrit directement sur la connexion, sans passer par le mécanisme
// d'attente d'une unique réponse OK/ERR/{ de Send() (inutilisable ici,
// cf. EnterRaw). L'appelant est responsable de lire les réponses via le
// channel renvoyé par EnterRaw.
func (s *Session) SendRaw(line string) error {
	_, err := s.conn.Write([]byte(line))
	return err
}


// Send retourne ("", nil) - une reponse vide sans erreur - quand l'appel est
// ignore (action deja en cours, ou trop rapprochee de la precedente). C'est
// le signal que les appelants (boutons du GUI) doivent verifier en premier
// pour ne rien afficher plutot que de traiter "" comme une vraie reponse.
//
// déroulé normal : on verrouille -> on lui dit qu'on attend une réponse
// (awaiting true) -> écris la commande sur le socket -> bloque jusqu'à ce
// qu'une ligne arrive dans pending, soit 5 secs s'écoulent sans réponse
// (serveur a planté, co morte...).
func (s *Session) Send(cmd string) (string, error) {
	// Refuse tout Send() concurrent au lieu de le mettre en attente : avec un
	// simple mutex, un double-clic pendant qu'une commande est deja en vol
	// attendait juste son tour puis renvoyait quand meme la MEME commande
	// une deuxieme fois une fois la premiere terminee - le serveur recevait
	// alors deux "MOVE east" d'affilee et declenchait sa detection de
	// commande repetee (cf. tcp/server/dispatch.go, previous/WARN). Ici, le
	// deuxieme clic est ignore silencieusement, sans jamais toucher le reseau.
	if !s.busy.CompareAndSwap(false, true) {
		return "", nil
	}
	defer s.busy.Store(false)

	// Cooldown : meme une fois le Send() precedent termine, un nouveau clic
	// trop rapproche (double-clic, ou un tap Fyne qui se declenche deux fois
	// pour un seul clic) est lui aussi ignore silencieusement.
	s.cooldownMu.Lock()
	now := time.Now()
	if now.Before(s.nextAllowed) {
		s.cooldownMu.Unlock()
		return "", nil
	}
	s.nextAllowed = now.Add(sendCooldown)
	s.cooldownMu.Unlock()

	// Vide toute valeur perimee laissee par un envoi precedent (protection
	// defensive : ne devrait plus arriver depuis que le serveur ne renvoie
	// plus qu'une seule ligne par commande, mais evite qu'un reliquat ne
	// pollue la reponse de CE Send()).
	select {
	case <-s.pending:
	default:
	}

	s.awaiting.Store(true)
	defer s.awaiting.Store(false)

	if _, err := s.conn.Write([]byte(cmd + "\n")); err != nil { // on envoie la commande sur conn
		return "", err
	}

	// attente = si deliver envoie une ligne dans pending, on la retourne
	select {
	case line := <-s.pending:
		return line, nil
	case <-time.After(5 * time.Second):
		return "", fmt.Errorf("timeout wainting for response to %q", cmd)
	}
}


func (s *Session) Quit() {
	fmt.Fprint(s.conn, "QUIT\n")
}
