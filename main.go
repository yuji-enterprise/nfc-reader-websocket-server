package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/peterhellberg/acr122u"
)

func main() {
	upgrader := &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	clientSlice := []struct{}{}
	nfcChan := make(chan string)
	go func() {
		ctx, err := acr122u.EstablishContext()
		if err != nil {
			log.Fatal("Error : Can not detect acr122U reader. Please make sure reader already plugin on you computer then restart.")
		} else {
			log.Println("Connect acr123u reader")
			ctx.ServeFunc(func(card acr122u.Card) {
				log.Printf("Reader Scan: %x\n", card.UID())

				// If has WS client, put uid to channel.
				if len(clientSlice) > 0 {
					nfcChan <- fmt.Sprintf("%x", card.UID())
				}
			})
		}
	}()

	http.HandleFunc("/card-reader", func(w http.ResponseWriter, r *http.Request) {
		log.Println("WS Client connected.")
		if conn, err := upgrader.Upgrade(w, r, nil); err != nil {
			log.Println("upgrade:", err)
			return
		} else {
			clientSlice = append(clientSlice, struct{}{})

			// Note : about Error 'websocket: close sent'
			// https://stackoverflow.com/questions/67030955/get-websocket-error-close-sent-when-i-navigate-to-other-page
			closeStatusChan := make(chan struct{})

			go reader(conn, closeStatusChan)
			go writer(conn, closeStatusChan, nfcChan, &clientSlice)
		}
	})

	log.Println("WS server start at :8899")
	log.Fatal(http.ListenAndServe(":8899", nil))
}

func reader(conn *websocket.Conn, closeStatusChan chan struct{}) {
	defer conn.Close()
	defer close(closeStatusChan)

	for {
		if mtype, msg, err := conn.ReadMessage(); err != nil {
			log.Println("WS read:", err)
			break
		} else {
			log.Printf("WS receive: %s\n", msg)
			if err := conn.WriteMessage(mtype, append(msg, []byte(" from WS server")...)); err != nil {
				log.Println("WS write:", err)
				break
			}
		}
	}
}

func writer(conn *websocket.Conn, closeStatusChan chan struct{}, nfcChan chan string, clientSlice *[]struct{}) {
	defer conn.Close()

	for {
		select {
		case <-closeStatusChan:
			log.Println("WS Client disconnected.")
			(*clientSlice) = (*clientSlice)[1:]
			// the reader is done, so return
			return
		case uid := <-nfcChan:
			log.Printf("WS Send: %s\n", uid)
			if err := conn.WriteMessage(websocket.TextMessage, []byte(uid)); err != nil {
				log.Println("WS write:", err)
				return
			}
		}
	}
}
