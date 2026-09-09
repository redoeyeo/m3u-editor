// ui/window.go
package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/redoeyeo/m3u-editor/domain"
)

// ── Model ──────────────────────────────────────────

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
		return m.playlist.Items[row].Path
	}
	return ""
}

// ── Context ───────────────────────────────────────

type uiCtx struct {
	playlist        *domain.Playlist
	model           *TrackModel
	tableView       *walk.TableView
	preview         *walk.TextEdit
	currentFilePath string // ← новое
}

func (c *uiCtx) refresh() {
	c.model.PublishRowsReset()
	text := strings.ReplaceAll(c.playlist.ToM3U(), "\n", "\r\n")
	text = strings.ReplaceAll(text, "\x00", "")
	c.preview.SetText(text)
}

// ── Обработчики кнопок ─────────────────────────────
func (c *uiCtx) addFromFiles() {
	dlg := walk.FileDialog{
		// Фильтр оставляем в формате "Описание|расширения"
		Filter: "Audio files|*.mp3;*.wav;*.flac;*.m4a|All files|*.*",
	}

	// ВАЖНО: Используем метод ShowOpenMultiple вместо ShowOpen
	// Он специально создан для выбора нескольких файлов
	if ok, err := dlg.ShowOpenMultiple(nil); !ok {
		if err != nil {
			walk.MsgBox(nil, "Ошибка", err.Error(), walk.MsgBoxIconError)
		}
		return
	}

	addedCount := 0
	for _, path := range dlg.FilePaths {
		name := filepath.Base(path)
		c.playlist.Add(name, path)
		addedCount++
	}

	if addedCount == 0 {
		walk.MsgBox(nil, "Инфо", "Файлы не выбраны", walk.MsgBoxIconInformation)
		return
	}

	c.refresh()
	walk.MsgBox(nil, "Готово", fmt.Sprintf("Добавлено файлов: %d", addedCount), walk.MsgBoxIconInformation)
}

func (c *uiCtx) removeTrack() {
	i := c.tableView.CurrentIndex()
	if i < 0 {
		walk.MsgBox(nil, "Ошибка", "Выберите трек для удаления", walk.MsgBoxIconError)
		return
	}
	c.playlist.Remove(i)
	c.refresh()
}

func (c *uiCtx) openFile() {
	dlg := walk.FileDialog{
		Filter: "M3U Playlist (*.m3u)|*.m3u|All files (*.*)|*.*",
	}
	if ok, _ := dlg.ShowOpen(nil); !ok {
		return
	}
	data, err := os.ReadFile(dlg.FilePath)
	if err != nil {
		walk.MsgBox(nil, "Ошибка чтения", err.Error(), walk.MsgBoxIconError)
		return
	}
	c.currentFilePath = dlg.FilePath // ← новое
	*c.playlist = *domain.ParseM3U(string(data))
	c.refresh()
}

func (c *uiCtx) saveFile() {
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
	c.currentFilePath = path // ← новое
	err := os.WriteFile(path, []byte(c.playlist.ToM3U()), 0644)
	if err != nil {
		walk.MsgBox(nil, "Ошибка записи", err.Error(), walk.MsgBoxIconError)
		return
	}
	walk.MsgBox(nil, "Готово", "Файл сохранён:\n"+path, walk.MsgBoxIconInformation)
}

func (c *uiCtx) refreshFile() {
	if c.currentFilePath == "" {
		walk.MsgBox(nil, "Ошибка", "Сначала откройте или сохраните файл", walk.MsgBoxIconError)
		return
	}
	err := os.WriteFile(c.currentFilePath, []byte(c.playlist.ToM3U()), 0644)
	if err != nil {
		walk.MsgBox(nil, "Ошибка записи", err.Error(), walk.MsgBoxIconError)
		return
	}
	walk.MsgBox(nil, "Готово", "Файл обновлён:\n"+c.currentFilePath, walk.MsgBoxIconInformation)
}

// ── Главное окно ────────────────────────────────────

func Run(playlist *domain.Playlist) error {
	ctx := &uiCtx{
		playlist: playlist,
		model:    &TrackModel{playlist: playlist},
	}

	_, err := MainWindow{
		Title:   "M3U Editor",
		MinSize: Size{700, 500},
		Layout:  VBox{},
		Children: []Widget{
			// Кнопки управления
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{Text: "Добавить файлы", OnClicked: ctx.addFromFiles},
					PushButton{Text: "Удалить выбранный", OnClicked: ctx.removeTrack},
					PushButton{Text: "Обновить файл", OnClicked: ctx.refreshFile}, // ← новое
				},
			},

			// Таблица + предпросмотр
			HSplitter{
				Children: []Widget{
					TableView{
						AssignTo: &ctx.tableView,
						Model:    ctx.model,
						Columns: []TableViewColumn{
							{Title: "Название", Width: 300},
							{Title: "Путь", Width: 400},
						},
					},
					TextEdit{
						AssignTo: &ctx.preview,
						ReadOnly: true,
						Text:     "#EXTM3U ...",
					},
				},
			},

			// Открыть / Сохранить
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{Text: "Открыть .m3u", OnClicked: ctx.openFile},
					PushButton{Text: "Сохранить .m3u", OnClicked: ctx.saveFile},
				},
			},
		},
	}.Run()

	return err
}
