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
}


// initialiser nouvelle session
func NewSession(conn net.Conn) *Session {
	return &Session{conn: conn, pending: make(chan string, 1)}
}


// à chaque ligne reçue du serveur, readServerMessages demande d'abord à Deliver :
// "est-ce que quelqu'un attend cette ligne ?" = awaiting = vrai = pousse dans pending
// qui réveille le Send qui était bloqué dessus.
func (s *Session) Deliver(line string) bool {
	if !s.awaiting.Load() {
		return false
	}
	// Une vraie réponse commence toujours par "OK ", "ERR", ou
	// (cas de MOVE, qui renvoie la Location brute en JSON) "{" ; tout le reste
	// n'est pas ce qu'on attend, on le laisse suivre son chemin normal
	// (affichage terminal) et on continue d'attendre.
	if !strings.HasPrefix(line, "OK ") && !strings.HasPrefix(line, "ERR") && !strings.HasPrefix(line, "{") {
		return false
	}
	// Envoi non-bloquant : pending a une capacite de 1, donc si une ligne y
	// est deja en attente (ne devrait plus arriver depuis que chaque
	// commande ne renvoie plus qu'une seule ligne, cf. tcp/server/dispatch.go),
	// on ne bloque jamais la boucle de lecture unique du client entier pour
	// autant - on log et on laisse tomber cette ligne en trop plutot que de
	// geler toute reception reseau en attendant qu'un Send() vienne la lire.
	select {
	case s.pending <- line:
	default:
		fmt.Println("[GUI] réponse en trop ignorée (pending déjà plein) :", line)
	}
	return true
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
