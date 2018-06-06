// http://mattn.github.io/go-gtk/
// https://github.com/mattn/go-gtk
// apt-get install libgtk2.0-dev libglib2.0-dev libgtksourceview2.0-dev
// go get  github.com/mattn/go-gtk/gdkpixbuf
// go get github.com/mattn/go-pointer
// go get https://github.com/mattn/go-gtk

package lib

import (
	"github.com/mattn/go-gtk/glib"
	"github.com/mattn/go-gtk/gtk"
)

func Run(db Database) (string, error) {
	gtk.Init(nil)
	window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
	window.SetPosition(gtk.WIN_POS_CENTER)
	window.SetResizable(false)
	window.SetTitle("GoPass")
	window.Connect("destroy", func(ctx *glib.CallbackContext) {
		println("got destroy!", ctx.Data().(string))
		gtk.MainQuit()
	}, "foo")

	//--------------------------------------------------------
	// GtkVBox
	//--------------------------------------------------------
	vbox := gtk.NewVBox(false, 1)

	//--------------------------------------------------------
	// GtkMenuBar
	//--------------------------------------------------------
	menubar := gtk.NewMenuBar()
	vbox.PackStart(menubar, false, false, 0)

	//--------------------------------------------------------
	// GtkVPaned
	//--------------------------------------------------------
	vpaned := gtk.NewVPaned()
	vbox.Add(vpaned)

	//--------------------------------------------------------
	// GtkFrame
	//--------------------------------------------------------
	frame1 := gtk.NewFrame("Store")
	framebox1 := gtk.NewVBox(false, 1)
	frame1.Add(framebox1)

	vpaned.Pack1(frame1, false, false)

	//--------------------------------------------------------
	// GtkEntry
	//--------------------------------------------------------
	key_row := gtk.NewHBox(false, 1)
	key_label := gtk.NewLabel("Key            ")
	key_row.Add(key_label)
	key_entry := gtk.NewEntry()
	// key_entry.SetText("key")
	key_row.Add(key_entry)

	pass_row := gtk.NewHBox(false, 1)
	pass_label := gtk.NewLabel("Passphrase")
	pass_row.Add(pass_label)
	pass_entry := gtk.NewEntry()
	// pass_entry.SetText("")
	pass_row.Add(pass_entry)

	val_row := gtk.NewHBox(false, 1)
	val_label := gtk.NewLabel("Value         ")
	val_row.Add(val_label)
	val_entry := gtk.NewEntry()
	// val_entry.SetText("")
	val_row.Add(val_entry)

	framebox1.Add(key_row)
	framebox1.Add(pass_row)
	framebox1.Add(val_row)

	//--------------------------------------------------------
	// GtkHBox
	//--------------------------------------------------------
	buttons := gtk.NewHBox(false, 1)

	get_button := gtk.NewButtonWithLabel("GET")
	get_button.Clicked(func() {
		value, err := db.Get("store", key_entry.GetText(), pass_entry.GetText())
		println(value, err)
		if nil != err {
			val_entry.SetText(err.Error())
		} else {
			val_entry.SetText(value)
		}
	})
	buttons.Add(get_button)

	set_button := gtk.NewButtonWithLabel("SET")
	set_button.Clicked(func() {
		db.Set("store", key_entry.GetText(), val_entry.GetText(), pass_entry.GetText())
	})
	buttons.Add(set_button)

	framebox1.PackStart(buttons, false, false, 0)

	//--------------------------------------------------------
	// GtkStatusbar
	//--------------------------------------------------------
	// statusbar := gtk.NewStatusbar()
	// context_id := statusbar.GetContextId("go-gtk")
	// statusbar.Push(context_id, "")
	// framebox1.PackStart(statusbar, false, false, 0)

	//--------------------------------------------------------
	// Event
	//--------------------------------------------------------
	window.Add(vbox)
	window.SetSizeRequest(300, 200)
	window.ShowAll()
	gtk.Main()

	return "", nil
}
