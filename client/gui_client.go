package main

import (
	"bufio"
	"fmt"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

type roleType int
type colorType int

const (
	MAFIA roleType = iota
	TOWN
	DOCTOR
	INSPECTOR
	AVENGER
	WITCH
	JESTER
)

// Channel for sending messages to the server
var SentMessages = make(chan string)
// Channel for incoming messages from the server
var ReceivedMessages = make(chan string)

var isOwner bool
var players = make(map[string]bool)
var pictures = make(map[string]string)
var colors = make(map[string]string)

var username string

var role roleType

var pictureCounter = 0

func main() {

	var address string
	if len(os.Args) > 1 {
		address = "164.90.171.162:8888"
	} else {
		address = "127.0.0.1:8888"
	}

	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("Could not connect to server. Please try again later.")
		os.Exit(1)
	}

	go func() {
		connbuf := bufio.NewReader(conn)

		for {
			b := make([]byte, 4)
			io.ReadFull(connbuf, b)
			messageLength := int(uint(b[0]) | uint(b[1])<<8 | uint(b[2])<<16 | uint(b[3])<<24)
			tmp := make([]byte, messageLength + 1)
			_, err := io.ReadFull(connbuf, tmp)

			if err != nil {
				fmt.Println("The server is down. Please try again later.")
				os.Exit(1)
				break
			}
			ReceivedMessages <- string(tmp[:messageLength])
		}

	}()

	gtk.Init(nil)

	win := createLoginWindow()

	win.ShowAll()

	go func() {
		for {
			select {
			case poruka1 := <-ReceivedMessages:

				poruka1 = strings.Trim(poruka1, "\r\n")

				args := strings.Split(poruka1, " ")

				cmd := strings.TrimSpace(args[0])

				poruka1 = poruka1[strings.Index(poruka1, " ")+1:]

				poruka2 := poruka1[strings.Index(poruka1, " ")+1:]
				switch cmd {
				case "/open":
					for k, _ := range players {
						delete(players, k)
						delete(pictures, k)
						delete(colors, k)
					}
					pictureCounter = 0
					username = args[1]
					isOwner = false
					glib.IdleAdd(func() {
						resetToMainWindow(win)
						UpdateText1(poruka2)
					})
				case "/firstname":
					glib.IdleAdd(func() {
						dialog := gtk.MessageDialogNew(win, gtk.DIALOG_MODAL, gtk.MESSAGE_INFO, gtk.BUTTONS_OK, "Name taken, try again")
						dialog.SetTitle("Username taken")
						dialog.Run()
						dialog.Destroy()
					})
				case "/firstws":
					glib.IdleAdd(func() {
						dialog := gtk.MessageDialogNew(win, gtk.DIALOG_MODAL, gtk.MESSAGE_INFO, gtk.BUTTONS_OK, "No whitespaces allowed, try again")
						dialog.SetTitle("No whitespaces allowed")
						dialog.Run()
						dialog.Destroy()
					})

				case "/text1":
					glib.IdleAdd(func() { UpdateText1(poruka1) })
				case "/text2":
					glib.IdleAdd(func() { UpdateText2(poruka1) })
				case "/join":
					glib.IdleAdd(func() {
						for k := range players {
							delete(colors, k)
							delete(players, k)
							delete(pictures, k)
						}
						pictureCounter = 0
						if len(args) == 3 {
							isOwner = true
						}
						for _, player := range args[2:] {
							players[player] = true
							colors[player] = selectColor(pictureCounter)
							pictures[player] = selectPicture(pictureCounter)
							pictureCounter++
						}

						UpdateJoin(win, players)
						UpdateText1("Welcome to " + args[1])
					})
				case "/leave":
					glib.IdleAdd(func() {
						delete(colors, args[1])
						delete(players, args[1])
						delete(pictures, args[1])
						pictureCounter--
						UpdateJoin(win, players)
						UpdateText1(args[1] + " left the room!")
					})
				case "/owner":
					if username == args[1] {
						isOwner = true
					}
					glib.IdleAdd(func() { UpdateText2(poruka1) })
				case "/role":
					switch args[1] {
					case "town":
						role = TOWN
					case "mafia":
						role = MAFIA
					case "doctor":
						role = DOCTOR
					case "inspector":
						role = INSPECTOR
					case "avenger":
						role = AVENGER
					case "witch":
						role = WITCH
					case "jester":
						role = JESTER
					}
					glib.IdleAdd(func() { UpdateText2(poruka2) })
				case "/night":

					glib.IdleAdd(func() {
						UpdateNight(win, players)
						UpdateText1(poruka1)
					})
				case "/day":
					glib.IdleAdd(func() {
						UpdateDay(win, players)
						UpdateText1(poruka1)
					})
				case "/vote":
					glib.IdleAdd(func() {
						UpdateVote(win, players)
						UpdateText1(poruka1)
					})
				case "/dead":
					{
						players[args[1]] = false
						glib.IdleAdd(func() { UpdateText2(poruka2) })
					}
				case "/won":
					{
						glib.IdleAdd(func() {
							buffer, _ := text2.GetBuffer()
							buffer.SetText("")
							UpdateText1(poruka1)
						})
					}
				default:
					fmt.Print(string(poruka1))
				}

			case porukaSlanje := <-SentMessages:
				go func() {
					conn.Write([]byte(porukaSlanje))
				}()
			}
		}
	}()

	gtk.Main()
}

func selectPicture(counter int) string {
	path := "picture" + strconv.Itoa(counter) + ".png"
	return path
}

func selectColor(counter int) string {
	var color string
	switch counter {
	case 0:
		color = "blue"
	case 1:
		color = "red"
	case 2:
		color = "yellow"
	case 3:
		color = "magenta"
	case 4:
		color = "black"
	case 5:
		color = "purple"
	case 6:
		color = "orange"
	case 7:
		color = "green"
	case 8:
		color = "cyan"
	case 9:
		color = "white"
	}

	return color
}
