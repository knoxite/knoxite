// Server-UDP
// Startet mit `go run ./server.go <port>`
// Wartet auf "UDP-Connections" und gibt eine Antwort (z.B. Registrierung erfolgreich)
// UDP-Verbindungen müssen eine Nachricht mitliefern, die dem knoxite server die Information
// zum Client gibt (Registrierungsnummer, Authentifizierung, etc.)
// Speicherung: Wie speichern wir was wo? (sqlite-DB, wo und wie werden die config Dateien
// abgespeichert (verschlüsselt)? Wie werden sie deployed (auf dem client Geräten)?
// Wozu HTTP-Server? Storage Backend oder Admin-Interface oder beides?
// Server als subcommand?

package main

import (
        "fmt"
        "math/rand"
        "net"
        "os"
        "strconv"
        "strings"
        "time"
)

func random(min, max int) int {
        return rand.Intn(max-min) + min
}

func main() {
        arguments := os.Args
        if len(arguments) == 1 {
                fmt.Println("Please provide a port number!")
                return
        }
        PORT := ":" + arguments[1]

        s, err := net.ResolveUDPAddr("udp4", PORT)
        if err != nil {
                fmt.Println(err)
                return
        }

        connection, err := net.ListenUDP("udp4", s)
        if err != nil {
                fmt.Println(err)
                return
        }

        defer connection.Close()
        buffer := make([]byte, 1024)
        rand.Seed(time.Now().Unix())

        for {
                n, addr, err := connection.ReadFromUDP(buffer)
                fmt.Print("-> ", string(buffer[0:n-1]))

                if strings.TrimSpace(string(buffer[0:n])) == "STOP" {
                        fmt.Println("Exiting UDP server!")
                        return
                }

                data := []byte(strconv.Itoa(random(1, 1001)))
                fmt.Printf("data: %s\n", string(data))
                _, err = connection.WriteToUDP(data, addr)
                if err != nil {
                        fmt.Println(err)
                        return
                }
        }
}
