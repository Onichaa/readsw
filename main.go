package main

import (
    "context"
    "fmt"
    "math/rand"
    "os"
    "strings"
    "time"
    "os/signal"
    "syscall"

    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/store/sqlstore"
    "go.mau.fi/whatsmeow/types"
    "go.mau.fi/whatsmeow/types/events"
    waLog "go.mau.fi/whatsmeow/util/log"

    _ "github.com/mattn/go-sqlite3"
    "github.com/mdp/qrterminal"
)

func main() {
    NewBot("6285796103714", func(k string) { //masukkan nomor kamu yang ingin di pasangkan auto read story wa
        println(k)
    }) 
}

func registerHandler(client *whatsmeow.Client) func(evt interface{}) {
    return func(evt interface{}) {
        switch v := evt.(type) {
        case *events.Message:
            if v.Info.Chat.String() == "status@broadcast" {
                if v.Info.Type != "reaction" {
                    sender := v.Info.Sender.String()
                    allowedSenders := []string{ //disini isi nomer yang ingin agar bot tidak otomatis read sw dari list nomor dibawah 
                        "6281447477366@s.whatsapp.net",
                        "6281457229553@s.whatsapp.net",
                    }
                    if contains(allowedSenders, sender) {
                        return
                    }

                    time.Sleep(2 * time.Minute) //ini agar 2 menit kedepan akan otomatis dibaca tiap sw
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

    container, err := sqlstore.New("sqlite3", "file:"+id+".db?_foreign_keys=on", dbLog)
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

        switch int(questLogin()) {
            
            case 1:
                if err := client.Connect(); 
                    err != nil {
                    fmt.Println(err)
                }

                code, err := client.PairPhone(id, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
                if err != nil {
                    fmt.Println(err)
                }

                fmt.Println("Kode verifikasi anda adalah: " + code)
                break
            
            case 2:
                qrChan, _ := client.GetQRChannel(context.Background())
                if err := client.Connect(); 
                    err != nil {
                    fmt.Println(err)
                }
                for evt := range qrChan {
                    switch string(evt.Event) {
                    case "code":
                        qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
                        fmt.Println("Scan Qrnya!!")
                        break
                    }
                }
                break
            
            default:
                fmt.Println("Pilih apa?")
            }
    } else {
        fmt.Println("Connected to readsw!!")
    }

    // Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
    c := make(chan os.Signal)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    <-c

    client.Disconnect()
    
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

func questLogin() int {
    fmt.Println("Silahkan Pilih Opsi Login:")
    fmt.Println("1. Pairing Code")
    fmt.Println("2. Qr")
    fmt.Print("Pilih : ")
    var input int
    _, err := fmt.Scanln(&input)
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        return 0
    }

    return input
}
