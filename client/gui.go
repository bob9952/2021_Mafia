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

var button_grid *gtk.Grid

func createLoginWindow() *gtk.Window {

	myCSSNight()

	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		log.Fatal("Window creation failed", err)
	}

	win.SetTitle("Username")
	win.SetPosition(gtk.WIN_POS_CENTER)
	win.Connect("destroy", func() {
		PoslatePoruke <- "/quit\n"
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

	btn_enter, _ := gtk.ButtonNewWithLabel("Enter username")
	btn_enter.Connect("clicked", func() {
		message, _ := entry.GetText()
		message = strings.Trim(message, "\r\n")
		PoslatePoruke <- message + "\n"
	})

	grid.Add(entry)
	grid.Add(btn_enter)

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
	button_grid = grid1

	//GRID FOR THE PLAYER
	grid_player, err := gtk.GridNew()
	if err != nil {
		log.Fatal("Grid creation failed", err)
	}
	grid_player.SetOrientation(gtk.ORIENTATION_VERTICAL)
	grid_player.SetBorderWidth(30)

	label, err := gtk.LabelNew(username)
	if err != nil {
		log.Fatal("Label creation failed", err)
	}
	label.SetMarginBottom(10)

	grid_player.Add(label)
	pixbuf, _ := gdk.PixbufNewFromFile("awesomeface.png")
	pixbuf, _ = pixbuf.ScaleSimple(100, 100, gdk.INTERP_BILINEAR)
	img, _ := gtk.ImageNew()
	img.SetFromPixbuf(pixbuf)
	grid_player.Add(img)

	//GRID FOR THE PLAYER

	grid1.Add(grid_player)

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
			PoslatePoruke <- message + "\n"
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

	index_dots := strings.Index(msg, ":")
	index_blank := strings.Index(msg, " ")

	buffer, _ := text2.GetBuffer()
	_, end := buffer.GetBounds()
	if index_dots != -1 && index_blank > index_dots {

		name := msg[0:index_dots]
		buffer.InsertMarkup(end, "\n<span foreground = \""+colors[name]+"\">"+name+":</span>"+msg[index_dots+1:])
	} else {
		buffer.Insert(end, "\n"+msg)
	}
	text2.ScrollToMark(buffer.GetMark("end"), 0.0, true, 0.5, 0.5)
	text2.SetIndent(10)
}

func myCSSNight() {

	provider, _ := gtk.CssProviderNew()
	display, _ := gdk.DisplayGetDefault()
	screen, _ := display.GetDefaultScreen()

	gtk.AddProviderForScreen(screen, provider, gtk.STYLE_PROVIDER_PRIORITY_USER)
	provider.LoadFromPath("styleNight.css")
}

func myCSSDay() {

	provider, _ := gtk.CssProviderNew()
	display, _ := gdk.DisplayGetDefault()
	screen, _ := display.GetDefaultScreen()

	gtk.AddProviderForScreen(screen, provider, gtk.STYLE_PROVIDER_PRIORITY_USER)
	provider.LoadFromPath("styleDay.css")
}

func myCSSVote() {

	provider, _ := gtk.CssProviderNew()
	display, _ := gdk.DisplayGetDefault()
	screen, _ := display.GetDefaultScreen()

	gtk.AddProviderForScreen(screen, provider, gtk.STYLE_PROVIDER_PRIORITY_USER)
	provider.LoadFromPath("styleVote.css")
}

func UpdateJoin(win *gtk.Window, players map[string]bool) {

	children := button_grid.GetChildren()
	for child := children; child != nil; child = child.Next() {
		child.Data().(*gtk.Widget).Destroy()
		//button_grid.Remove(child.First())
	}

	keys := make([]string, 0)
	for k, _ := range players {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, player := range keys {
		grid_player, _ := gtk.GridNew()
		grid_player.SetOrientation(gtk.ORIENTATION_VERTICAL)
		grid_player.SetBorderWidth(30)

		label, _ := gtk.LabelNew(player)

		pixbuf, _ := gdk.PixbufNewFromFile(pictures[player])
		pixbuf, _ = pixbuf.ScaleSimple(100, 100, gdk.INTERP_BILINEAR)
		img, _ := gtk.ImageNew()
		img.SetFromPixbuf(pixbuf)

		grid_player.Add(label)
		grid_player.Add(img)

		if player == username && isOwner {

			btn_player, _ := gtk.ButtonNewWithLabel("start")
			btn_player.SetName(player)

			btn_player.Connect("clicked", func() {
				PoslatePoruke <- "/start\n"
			})

			grid_player.Add(btn_player)
		}
		if player == username {
			context, _ := label.GetStyleContext()
			context.AddClass("labela_username")
		}

		button_grid.Add(grid_player)

		win.ShowAll()
	}
}

func UpdateNight(win *gtk.Window, players map[string]bool) {

	myCSSNight()

	children := button_grid.GetChildren()
	for child := children; child != nil; child = child.Next() {
		child.Data().(*gtk.Widget).Destroy()
		//button_grid.Remove(child.First())
	}

	keys := make([]string, 0)
	for k, _ := range players {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, player := range keys {
		alive := players[player]
		grid_player, _ := gtk.GridNew()
		grid_player.SetOrientation(gtk.ORIENTATION_VERTICAL)
		grid_player.SetBorderWidth(30)

		label, _ := gtk.LabelNew(player)

		pixbuf, _ := gdk.PixbufNewFromFile(pictures[player])
		pixbuf, _ = pixbuf.ScaleSimple(100, 100, gdk.INTERP_BILINEAR)
		img, _ := gtk.ImageNew()
		img.SetFromPixbuf(pixbuf)

		grid_player.Add(label)
		grid_player.Add(img)

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
			btn_player, _ := gtk.ButtonNewWithLabel(caption)
			btn_player.SetName(player)

			btn_player.Connect("clicked", func() {
				labelText, _ := label.GetText()
				PoslatePoruke <- "/" + caption + " " + labelText + "\n"
			})

			grid_player.Add(btn_player)
		}
		if player == username && role == DOCTOR && alive {
			caption := "protect"
			btn_player, _ := gtk.ButtonNewWithLabel(caption)
			btn_player.SetName(player)

			btn_player.Connect("clicked", func() {
				labelText, _ := label.GetText()
				PoslatePoruke <- "/" + caption + " " + labelText + "\n"
			})

			grid_player.Add(btn_player)
		}
		if role == WITCH && alive && players[username] {
			var caption_heal = "heal"
			btn_player1, _ := gtk.ButtonNewWithLabel(caption_heal)
			btn_player1.SetName(player)

			btn_player1.Connect("clicked", func() {
				labelText, _ := label.GetText()
				PoslatePoruke <- "/" + caption_heal + " " + labelText + "\n"
			})

			grid_player.Add(btn_player1)

			var caption_poison = "poison"
			btn_player2, _ := gtk.ButtonNewWithLabel(caption_poison)
			btn_player2.SetName(player)

			btn_player2.Connect("clicked", func() {
				labelText, _ := label.GetText()
				PoslatePoruke <- "/" + caption_poison + " " + labelText + "\n"
			})

			grid_player.Add(btn_player2)
			
			btn_player3, _ := gtk.ButtonNewWithLabel("skip")
			btn_player3.SetName(player)

			btn_player3.Connect("clicked", func() {
				PoslatePoruke <- "/poison skip\n"
			})

			grid_player.Add(btn_player3)
		}
		if role == AVENGER && alive && players[username] {
			btn_player1, _ := gtk.ButtonNewWithLabel("skip")
			btn_player1.SetName(player)

			btn_player1.Connect("clicked", func() {
				PoslatePoruke <- "/" + "pull skip" + "\n"
			})
			
			grid_player.Add(btn_player1)
		}

		if player == username {
			context, _ := label.GetStyleContext()
			context.AddClass("labela_username")
		}

		button_grid.Add(grid_player)

		win.ShowAll()
	}
}

func UpdateDay(win *gtk.Window, players map[string]bool) {

	myCSSDay()

	children := button_grid.GetChildren()
	for child := children; child != nil; child = child.Next() {
		child.Data().(*gtk.Widget).Destroy()
		//button_grid.Remove(child.First())
	}

	keys := make([]string, 0)
	for k, _ := range players {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, player := range keys {
		grid_player, _ := gtk.GridNew()
		grid_player.SetOrientation(gtk.ORIENTATION_VERTICAL)
		grid_player.SetBorderWidth(30)

		label, _ := gtk.LabelNew(player)

		pixbuf, _ := gdk.PixbufNewFromFile(pictures[player])
		pixbuf, _ = pixbuf.ScaleSimple(100, 100, gdk.INTERP_BILINEAR)
		img, _ := gtk.ImageNew()
		img.SetFromPixbuf(pixbuf)

		grid_player.Add(label)
		grid_player.Add(img)

		button_grid.Add(grid_player)

		if player == username {
			context, _ := label.GetStyleContext()
			context.AddClass("labela_username")
		}

		win.ShowAll()
	}
}

func UpdateVote(win *gtk.Window, players map[string]bool) {

	myCSSVote()

	children := button_grid.GetChildren()
	for child := children; child != nil; child = child.Next() {
		child.Data().(*gtk.Widget).Destroy()
		//button_grid.Remove(child.First())
	}

	keys := make([]string, 0)
	for k, _ := range players {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, player := range keys {
		alive := players[player]
		grid_player, _ := gtk.GridNew()
		grid_player.SetOrientation(gtk.ORIENTATION_VERTICAL)
		grid_player.SetBorderWidth(30)

		label, _ := gtk.LabelNew(player)

		pixbuf, _ := gdk.PixbufNewFromFile(pictures[player])
		pixbuf, _ = pixbuf.ScaleSimple(100, 100, gdk.INTERP_BILINEAR)
		img, _ := gtk.ImageNew()
		img.SetFromPixbuf(pixbuf)

		grid_player.Add(label)
		grid_player.Add(img)

		if player != username && alive && players[username] {

			btn_player, _ := gtk.ButtonNewWithLabel("vote")
			btn_player.SetName(player)

			btn_player.Connect("clicked", func() {
				labelText, _ := label.GetText()
				PoslatePoruke <- "/" + "vote" + " " + labelText + "\n"
			})

			grid_player.Add(btn_player)
		} else if player == username && alive && players[username] {
			btn_player, _ := gtk.ButtonNewWithLabel("skip")
			btn_player.SetName(player)

			btn_player.Connect("clicked", func() {
				PoslatePoruke <- "/" + "vote skip" + "\n"
			})

			grid_player.Add(btn_player)
		}

		if player == username {
			context, _ := label.GetStyleContext()
			context.AddClass("labela_username")
		}

		button_grid.Add(grid_player)

		win.ShowAll()
	}
}
