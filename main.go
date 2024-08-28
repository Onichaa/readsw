package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	waLog "go.mau.fi/whatsmeow/util/log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	NewBot("628388024064", func(k string) {
		println(k)
	})

	/* web server */
	port := os.Getenv("PORT")
	if port == "" {
		port = "1337" // Port default jika tidak ada yang disetel
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "readsw Bot Connected")
	})

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
	/* end web server */
}

func registerHandler(client *whatsmeow.Client) func(evt interface{}) {
	return func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			if v.Info.Chat.String() == "status@broadcast" {
				if v.Info.Type != "reaction" {
					sender := v.Info.Sender.String()
					allowedSenders := []string{
						"628114447477@s.whatsapp.net",
						"6285399290529@s.whatsapp.net",
						"6281355923372@s.whatsapp.net",
						"6285211511650@s.whatsapp.net",
						"6282292227486@s.whatsapp.net",
						"6282188338550@s.whatsapp.net",
						"6285240746696@s.whatsapp.net",
						"6281355971448@s.whatsapp.net",
						"6283140891091@s.whatsapp.net",
					}
					if contains(allowedSenders, sender) {
						return
					}

					emojis := []string{"🔥", "✨", "🌟", "🌞", "🎉", "🎊", "😺"}
					rand.Seed(time.Now().UnixNano())
					randomEmoji := emojis[rand.Intn(len(emojis))]

					reaction := client.BuildReaction(v.Info.Chat, v.Info.Sender, v.Info.ID, randomEmoji)
					extras := []whatsmeow.SendRequestExtra{}
					client.MarkRead([]types.MessageID{v.Info.ID}, v.Info.Timestamp, v.Info.Chat, v.Info.Sender)
					client.SendMessage(context.Background(), v.Info.Chat, reaction, extras...)
					fmt.Println("Berhasil melihat status", v.Info.PushName)
				}
			}
			if v.Message.GetConversation() == "Auto Read Story WhatsApp" {
				NewBot(v.Info.Sender.String(), func(k string) {
					client.SendMessage(context.Background(), v.Info.Sender, &waProto.Message{
						ExtendedTextMessage: &waProto.ExtendedTextMessage{
							Text: &k,
						},
					}, whatsmeow.SendRequestExtra{})
				})
			}
		}
	}
}

func NewBot(id string, callback func(string)) *whatsmeow.Client {
	if id == "" {
		callback("Nomor ?")
		return nil
	}
	id = strings.ReplaceAll(id, "admin", "")

	dbLog := waLog.Stdout("Database", "ERROR", true)

	container, err := sqlstore.New("sqlite3", "file:data/"+id+".db?_foreign_keys=on", dbLog)
	if err != nil {
		callback("Kesalahan (error)\n" + fmt.Sprintf("%s", err))
		return nil
	}
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		callback("Kesalahan (error)\n" + fmt.Sprintf("%s", err))
		return nil
	}
	clientLog := waLog.Stdout("Client", "ERROR", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)
	client.AddEventHandler(registerHandler(client))

	err = client.Connect()
	if err != nil {
		callback("Kesalahan (error)\n" + fmt.Sprintf("%s", err))
		return nil
	}

	if client.Store.ID == nil {
		code, _ := client.PairPhone(id, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
		callback("Kode verifikasi anda adalah " + code)
		time.AfterFunc(60*time.Second, func() {
			if client.Store.ID == nil {
				client.Disconnect()
				os.Remove("data/" + id + ".db")
				callback("melebihi 60 detik, memutuskan")
			}
		})

		client.SendPresence(types.PresenceUnavailable)
	} else {
		client.SendPresence(types.PresenceUnavailable)
	}
	return client
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
