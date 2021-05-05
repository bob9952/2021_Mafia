package main

import (
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"log"
	"sort"
	"strings"
)

var text1 *gtk.TextView
var text2 *gtk.TextView

var buttonGrid *gtk.Grid

func createLoginWindow() *gtk.Window {

	loadCSS("styleNight.css")

	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Window creation failed", err)
	}

	win.SetTitle("Username")
	win.SetPosition(gtk.WIN_POS_CENTER)
	win.Connect("destroy", func() {
		SentMessages <- "/quit\n"
		gtk.MainQuit()
	})

	grid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Grid creation failed", err)
	}
	grid.SetOrientation(gtk.ORIENTATION_VERTICAL)
	grid.SetBorderWidth(30)

	entry, err := gtk.EntryNew()
	if err != nil {
		log.Fatal("Entry creation failed", err)
	}
	entry.SetMarginBottom(10)

	btnEnter, _ := gtk.ButtonNewWithLabel("Enter username")
	btnEnter.Connect("clicked", func() {
		message, _ := entry.GetText()
		message = strings.Trim(message, "\r\n")
		SentMessages <- message + "\n"
	})

	grid.Add(entry)
	grid.Add(btnEnter)

	win.Add(grid)

	return win
}

func resetToMainWindow(win *gtk.Window) {

	widget, _ := win.GetChild()
	win.Remove(widget)
	//win.SetSizeRequest(1280, 720)

	win.SetTitle("Mafia")

	grid, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Grid creation failed", err)
	}
	grid.SetOrientation(gtk.ORIENTATION_VERTICAL)
	grid.SetBorderWidth(30)

	grid1, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Grid creation failed", err)
	}
	grid1.SetOrientation(gtk.ORIENTATION_HORIZONTAL)
	buttonGrid = grid1

	//GRID FOR THE PLAYER
	gridPlayer, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Grid creation failed", err)
	}
	gridPlayer.SetOrientation(gtk.ORIENTATION_VERTICAL)
	gridPlayer.SetBorderWidth(30)

	label, err := gtk.LabelNew(username)
	if err != nil {
		log.Fatal("Label creation failed", err)
	}
	label.SetMarginBottom(10)

	gridPlayer.Add(label)
	pixbuf, _ := gdk.PixbufNewFromFile("awesomeface.png")
	pixbuf, _ = pixbuf.ScaleSimple(100, 100, gdk.INTERP_BILINEAR)
	img, _ := gtk.ImageNew()
	img.SetFromPixbuf(pixbuf)
	gridPlayer.Add(img)

	//GRID FOR THE PLAYER

	grid1.Add(gridPlayer)

	grid2, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Grid creation failed", err)
	}
	grid2.SetOrientation(gtk.ORIENTATION_HORIZONTAL)

	grid_textboxes, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Grid creation failed", err)
	}
	grid_textboxes.SetOrientation(gtk.ORIENTATION_VERTICAL)

	scroll1, err := gtk.ScrolledWindowNew(nil, nil)
	table1, err := gtk.TextTagTableNew()
	buffer1, err := gtk.TextBufferNew(table1)
	textview1, err := gtk.TextViewNewWithBuffer(buffer1)
	textview1.SetMarginBottom(10)
	textview1.SetEditable(false)
	textview1.SetWrapMode(gtk.WRAP_WORD_CHAR)
	scroll1.Add(textview1)
	scroll1.SetSizeRequest(600, 100)
	grid_textboxes.Add(scroll1)

	text1 = textview1

	scroll2, err := gtk.ScrolledWindowNew(nil, nil)
	table2, err := gtk.TextTagTableNew()
	buffer2, err := gtk.TextBufferNew(table2)
	textview2, err := gtk.TextViewNewWithBuffer(buffer2)
	textview2.SetMarginBottom(10)
	textview2.SetEditable(false)
	textview2.SetCanFocus(false)
	textview2.SetWrapMode(gtk.WRAP_WORD_CHAR)
	scroll2.Add(textview2)
	scroll2.SetSizeRequest(600, 300)
	buffer2.CreateMark("end", buffer2.GetEndIter(), false)
	grid_textboxes.Add(scroll2)

	text2 = textview2

	entry, err := gtk.EntryNew()
	if err != nil {
		log.Fatal("Entry creation failed", err)
	}
	entry.SetMarginBottom(10)
	grid_textboxes.Add(entry)

	grid2.Add(grid_textboxes)

	grid.Add(grid1)
	grid.Add(grid2)

	win.Add(grid)
	win.Connect("key_press_event", func(widget *gtk.Window, event *gdk.Event) bool {
		native := gdk.EventKey{event}
		switch native.KeyVal() {
		case gdk.KEY_Return:
			message, _ := entry.GetText()
			if message == "\n" || message == "" {
				return false
			}
			message = strings.Trim(message, "\r\n")
			SentMessages <- message + "\n"
			entry.SetText("")
		default:
			return false
		}
		return false
	})

	win.ShowAll()

}

func UpdateText1(msg string) {

	buffer, _ := text1.GetBuffer()
	buffer.SetText(msg)
	text1.SetIndent(10)
	text1.SetPixelsAboveLines(5)
}

func UpdateText2(msg string) {

	indexDots := strings.Index(msg, ":")
	indexBlank := strings.Index(msg, " ")

	buffer, _ := text2.GetBuffer()
	_, end := buffer.GetBounds()
	if indexDots != -1 && indexBlank > indexDots {

		name := msg[0:indexDots]
		buffer.InsertMarkup(end, "\n<span foreground = \""+colors[name]+"\">"+name+":</span>"+msg[indexDots+1:])
	} else {
		buffer.Insert(end, "\n" + msg)
	}
	text2.ScrollToMark(buffer.GetMark("end"), 0.0, true, 0.5, 0.5)
	text2.SetIndent(10)
}

func loadCSS(path string) {

	provider, _ := gtk.CssProviderNew()
	display, _ := gdk.DisplayGetDefault()
	screen, _ := display.GetDefaultScreen()

	gtk.AddProviderForScreen(screen, provider, gtk.STYLE_PROVIDER_PRIORITY_USER)
	provider.LoadFromPath(path)
}

func UpdateJoin(win *gtk.Window, players map[string]bool) {

	children := buttonGrid.GetChildren()
	for child := children; child != nil; child = child.Next() {
		child.Data().(*gtk.Widget).Destroy()
	}

	keys := make([]string, 0)
	for k := range players {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, player := range keys {
		gridPlayer, _ := gtk.GridNew()
		gridPlayer.SetOrientation(gtk.ORIENTATION_VERTICAL)
		gridPlayer.SetBorderWidth(30)

		label, _ := gtk.LabelNew(player)

		pixbuf, _ := gdk.PixbufNewFromFile(pictures[player])
		pixbuf, _ = pixbuf.ScaleSimple(100, 100, gdk.INTERP_BILINEAR)
		img, _ := gtk.ImageNew()
		img.SetFromPixbuf(pixbuf)

		gridPlayer.Add(label)
		gridPlayer.Add(img)

		if player == username && isOwner {

			btnPlayer, _ := gtk.ButtonNewWithLabel("start")
			btnPlayer.SetName(player)

			btnPlayer.Connect("clicked", func() {
				SentMessages <- "/start\n"
			})

			gridPlayer.Add(btnPlayer)
		}
		if player == username {
			context, _ := label.GetStyleContext()
			context.AddClass("labela_username")
		}

		buttonGrid.Add(gridPlayer)

		win.ShowAll()
	}
}

func UpdateNight(win *gtk.Window, players map[string]bool) {

	loadCSS("styleNight.css")

	children := buttonGrid.GetChildren()
	for child := children; child != nil; child = child.Next() {
		child.Data().(*gtk.Widget).Destroy()
	}

	keys := make([]string, 0)
	for k, _ := range players {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, player := range keys {
		alive := players[player]
		gridPlayer, _ := gtk.GridNew()
		gridPlayer.SetOrientation(gtk.ORIENTATION_VERTICAL)
		gridPlayer.SetBorderWidth(30)

		label, _ := gtk.LabelNew(player)

		pixbuf, _ := gdk.PixbufNewFromFile(pictures[player])
		pixbuf, _ = pixbuf.ScaleSimple(100, 100, gdk.INTERP_BILINEAR)
		img, _ := gtk.ImageNew()
		img.SetFromPixbuf(pixbuf)

		gridPlayer.Add(label)
		gridPlayer.Add(img)

		if player != username && role != TOWN && role != JESTER && role != WITCH && alive && players[username] {
			var caption string
			switch role {
			case MAFIA:
				caption = "kill"
			case DOCTOR:
				caption = "protect"
			case INSPECTOR:
				caption = "inspect"
			case AVENGER:
				caption = "pull"
			}
			btnPlayer, _ := gtk.ButtonNewWithLabel(caption)
			btnPlayer.SetName(player)

			btnPlayer.Connect("clicked", func() {
				labelText, _ := label.GetText()
				SentMessages <- "/" + caption + " " + labelText + "\n"
			})

			gridPlayer.Add(btnPlayer)
		}
		if player == username && role == DOCTOR && alive {
			caption := "protect"
			btnPlayer, _ := gtk.ButtonNewWithLabel(caption)
			btnPlayer.SetName(player)

			btnPlayer.Connect("clicked", func() {
				labelText, _ := label.GetText()
				SentMessages <- "/" + caption + " " + labelText + "\n"
			})

			gridPlayer.Add(btnPlayer)
		}
		if role == WITCH && alive && players[username] {
			var captionHeal = "heal"
			btnPlayer1, _ := gtk.ButtonNewWithLabel(captionHeal)
			btnPlayer1.SetName(player)

			btnPlayer1.Connect("clicked", func() {
				labelText, _ := label.GetText()
				SentMessages <- "/" + captionHeal + " " + labelText + "\n"
			})

			gridPlayer.Add(btnPlayer1)

			var captionPoison = "poison"
			btnPlayer2, _ := gtk.ButtonNewWithLabel(captionPoison)
			btnPlayer2.SetName(player)

			btnPlayer2.Connect("clicked", func() {
				labelText, _ := label.GetText()
				SentMessages <- "/" + captionPoison + " " + labelText + "\n"
			})

			gridPlayer.Add(btnPlayer2)
			
			btnPlayer3, _ := gtk.ButtonNewWithLabel("skip")
			btnPlayer3.SetName(player)

			btnPlayer3.Connect("clicked", func() {
				SentMessages <- "/poison skip\n"
			})

			gridPlayer.Add(btnPlayer3)
		}
		if role == AVENGER && alive && players[username] {
			btnPlayer1, _ := gtk.ButtonNewWithLabel("skip")
			btnPlayer1.SetName(player)

			btnPlayer1.Connect("clicked", func() {
				SentMessages <- "/pull skip\n"
			})
			
			gridPlayer.Add(btnPlayer1)
		}

		if player == username {
			context, _ := label.GetStyleContext()
			context.AddClass("labela_username")
		}

		buttonGrid.Add(gridPlayer)

		win.ShowAll()
	}
}

func UpdateDay(win *gtk.Window, players map[string]bool) {

	loadCSS("styleDay.css")

	children := buttonGrid.GetChildren()
	for child := children; child != nil; child = child.Next() {
		child.Data().(*gtk.Widget).Destroy()
	}

	keys := make([]string, 0)
	for k, _ := range players {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, player := range keys {
		gridPlayer, _ := gtk.GridNew()
		gridPlayer.SetOrientation(gtk.ORIENTATION_VERTICAL)
		gridPlayer.SetBorderWidth(30)

		label, _ := gtk.LabelNew(player)

		pixbuf, _ := gdk.PixbufNewFromFile(pictures[player])
		pixbuf, _ = pixbuf.ScaleSimple(100, 100, gdk.INTERP_BILINEAR)
		img, _ := gtk.ImageNew()
		img.SetFromPixbuf(pixbuf)

		gridPlayer.Add(label)
		gridPlayer.Add(img)

		buttonGrid.Add(gridPlayer)

		if player == username {
			context, _ := label.GetStyleContext()
			context.AddClass("labela_username")
		}

		win.ShowAll()
	}
}

func UpdateVote(win *gtk.Window, players map[string]bool) {

	loadCSS("styleVote.css")

	children := buttonGrid.GetChildren()
	for child := children; child != nil; child = child.Next() {
		child.Data().(*gtk.Widget).Destroy()
	}

	keys := make([]string, 0)
	for k, _ := range players {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, player := range keys {
		alive := players[player]
		gridPlayer, _ := gtk.GridNew()
		gridPlayer.SetOrientation(gtk.ORIENTATION_VERTICAL)
		gridPlayer.SetBorderWidth(30)

		label, _ := gtk.LabelNew(player)

		pixbuf, _ := gdk.PixbufNewFromFile(pictures[player])
		pixbuf, _ = pixbuf.ScaleSimple(100, 100, gdk.INTERP_BILINEAR)
		img, _ := gtk.ImageNew()
		img.SetFromPixbuf(pixbuf)

		gridPlayer.Add(label)
		gridPlayer.Add(img)

		if player != username && alive && players[username] {

			btnPlayer, _ := gtk.ButtonNewWithLabel("vote")
			btnPlayer.SetName(player)

			btnPlayer.Connect("clicked", func() {
				labelText, _ := label.GetText()
				SentMessages <- "/" + "vote" + " " + labelText + "\n"
			})

			gridPlayer.Add(btnPlayer)
		} else if player == username && alive && players[username] {
			btnPlayer, _ := gtk.ButtonNewWithLabel("skip")
			btnPlayer.SetName(player)

			btnPlayer.Connect("clicked", func() {
				SentMessages <- "/vote skip\n"
			})

			gridPlayer.Add(btnPlayer)
		}

		if player == username {
			context, _ := label.GetStyleContext()
			context.AddClass("labela_username")
		}

		buttonGrid.Add(gridPlayer)

		win.ShowAll()
	}
}
