// ui/window.go
package ui

import (
	"os"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/redoeyeo/m3u-editor/domain"
)

// Model для таблицы (остается внутри UI, так как привязана к API walk)
type TrackModel struct {
	walk.TableModelBase
	playlist *domain.Playlist
}

func (m *TrackModel) RowCount() int {
	return len(m.playlist.Items)
}

func (m *TrackModel) Value(row, col int) interface{} {
	if row < 0 || row >= len(m.playlist.Items) {
		return ""
	}
	switch col {
	case 0:
		return m.playlist.Items[row].Name
	case 1:
		return m.playlist.Items[row].URL
	}
	return ""
}
func Run(playlist *domain.Playlist) error {
	model := &TrackModel{playlist: playlist}

	var nameEdit, urlEdit *walk.LineEdit
	var tableView *walk.TableView
	var previewEdit *walk.TextEdit

	refresh := func() {
		model.PublishRowsReset()
		previewEdit.SetText(playlist.ToM3U())
	}

	_, err := MainWindow{
		Title:   "M3U Editor",
		MinSize: Size{700, 500},
		Layout:  VBox{},
		Children: []Widget{
			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Label{Text: "Название:"},
					LineEdit{AssignTo: &nameEdit},
					Label{Text: "URL:"},
					LineEdit{AssignTo: &urlEdit},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						Text: "Добавить",
						OnClicked: func() {
							url := urlEdit.Text()
							if url == "" {
								walk.MsgBox(nil, "Ошибка", "URL пустой", walk.MsgBoxIconError)
								return
							}
							name := nameEdit.Text()
							if name == "" {
								name = url
							}
							playlist.Add(name, url)
							nameEdit.SetText("")
							urlEdit.SetText("")
							refresh()
						},
					},
					PushButton{
						Text: "Удалить",
						OnClicked: func() {
							i := tableView.CurrentIndex()
							if i < 0 {
								walk.MsgBox(nil, "Ошибка", "Выберите трек", walk.MsgBoxIconError)
								return
							}
							playlist.Remove(i)
							refresh()
						},
					},
					PushButton{
						Text: "Вверх",
						OnClicked: func() {
							i := tableView.CurrentIndex()
							if i > 0 {
								playlist.Move(i, i-1)
								refresh()
							}
						},
					},
					PushButton{
						Text: "Вниз",
						OnClicked: func() {
							i := tableView.CurrentIndex()
							if i >= 0 && i < len(playlist.Items)-1 {
								playlist.Move(i, i+1)
								refresh()
							}
						},
					},
				},
			},
			HSplitter{
				Children: []Widget{
					TableView{
						AssignTo: &tableView,
						Model:    model,
						Columns: []TableViewColumn{
							{Title: "Название", Width: 250},
							{Title: "URL", Width: 300},
						},
					},
					TextEdit{
						AssignTo: &previewEdit,
						ReadOnly: true,
						Text:     "#EXTM3U ...",
					},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						Text: "Открыть .m3u",
						OnClicked: func() {
							dlg := walk.FileDialog{
								Filter: "M3U Playlist (*.m3u)|*.m3u|All files (*.*)|*.*",
							}
							if ok, _ := dlg.ShowOpen(nil); !ok {
								return
							}
							data, err := os.ReadFile(dlg.FilePath)
							if err != nil {
								walk.MsgBox(nil, "Ошибка", err.Error(), walk.MsgBoxIconError)
								return
							}
							*playlist = *domain.ParseM3U(string(data))
							refresh()
						},
					},
					PushButton{
						Text: "Сохранить .m3u",
						OnClicked: func() {
							dlg := walk.FileDialog{
								Filter: "M3U Playlist (*.m3u)|*.m3u",
							}
							if ok, _ := dlg.ShowSave(nil); !ok {
								return
							}
							path := dlg.FilePath
							if !strings.HasSuffix(path, ".m3u") {
								path += ".m3u"
							}
							err := os.WriteFile(path, []byte(playlist.ToM3U()), 0644)
							if err != nil {
								walk.MsgBox(nil, "Ошибка", err.Error(), walk.MsgBoxIconError)
								return
							}
							walk.MsgBox(nil, "Готово", "Файл сохранён:\n"+path, walk.MsgBoxIconInformation)
						},
					},
				},
			},
		},
	}.Run()

	if err != nil {
		return err
	}

	return nil
}
