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
	"sync"
	"sync/atomic"
	"time"
)

type Session struct {
	conn net.Conn // connexion avec le serveur
	writeMu sync.Mutex // Mutex pour protéger l'écriture sur la co
	awaiting atomic.Bool // Indique si on attend une réponse à une commande envoyée (bool)
	pending chan string // canal pour recevoir la réponse du serveur
}


// initialiser nouvelle session
func New(conn net.Conn) *Session {
	return &Session{conn: conn, pending: make(chan string, 1)}
}


// à chaque ligne reçue du serveur, readServerMessages demande d'abord à Deliver :
// "est-ce que quelqu'un attend cette ligne ?" = awaiting = vrai = pousse dans pending
// qui réveille le Send qui était bloqué dessus.
func (s *Session) Deliver(line string) bool {
	if s.awaiting.Load() {
		s.pending <- line
		return true
	}
	return false
}


// déroulé: on verrouille -> on lui dit qu'on attend une réponse (awaiting true)
// -> écris la commande sur le socket -> bloque jusqu'à ce qu'une ligne arrive dans
// pending, soit 5 secs s'écoulent sans réponse (serveur a planté, co morte...). 
func (s *Session) Send(cmd string) (string, error) {

	s.writeMu.Lock()
	defer s.writeMu.Unlock()

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
