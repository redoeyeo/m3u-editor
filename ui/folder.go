package ui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type FoundTrack struct {
	Name string
	Path string
}

type SearchResultModel struct {
	walk.TableModelBase
	items []FoundTrack
}

func (m *SearchResultModel) RowCount() int { return len(m.items) }

func (m *SearchResultModel) Value(row, col int) interface{} {
	if row < 0 || row >= len(m.items) {
		return ""
	}
	switch col {
	case 0:
		return m.items[row].Name
	case 1:
		return m.items[row].Path
	}
	return ""
}

func levenshtein(s1, s2 string) int {
	s1, s2 = strings.ToLower(s1), strings.ToLower(s2)
	m, n := len(s1), len(s2)
	if m == 0 {
		return n
	}
	if n == 0 {
		return m
	}
	d := make([][]int, m+1)
	for i := range d {
		d[i] = make([]int, n+1)
		d[i][0] = i
	}
	for j := 0; j <= n; j++ {
		d[0][j] = j
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}
			a, b, c := d[i-1][j]+1, d[i][j-1]+1, d[i-1][j-1]+cost
			if b < a {
				a = b
			}
			if c < a {
				a = c
			}
			d[i][j] = a
		}
	}
	return d[m][n]
}

func (c *uiCtx) addTrack(name, path string) {
	if c.currentFilePath != "" {
		c.playlist.RelativeAdd(name, path, c.currentFilePath)
	} else {
		c.playlist.Add(name, path)
	}
}

func fuzzyMatch(query, target string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	target = strings.ToLower(target)
	if query == "" {
		return true
	}
	if strings.Contains(target, query) {
		return true
	}
	threshold := len(query) / 4
	if threshold < 2 {
		threshold = 2
	}
	for _, w := range strings.Fields(target) {
		if levenshtein(query, w) <= threshold {
			return true
		}
	}
	return false
}

func (c *uiCtx) addFromFolder() {
	dlg := walk.FileDialog{}
	if ok, _ := dlg.ShowBrowseFolder(nil); !ok {
		return
	}
	folder := dlg.FilePath
	audioExts := map[string]bool{".mp3": true, ".wav": true, ".flac": true, ".m4a": true}
	var found []FoundTrack
	filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if audioExts[ext] {
			name := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
			found = append(found, FoundTrack{Name: name, Path: path})
		}
		return nil
	})

	if len(found) == 0 {
		walk.MsgBox(nil, "Инфо", "В папке нет аудиофайлов", walk.MsgBoxIconInformation)
		return
	}

	c.showSearchDialog(found)
}
func (c *uiCtx) showSearchDialog(found []FoundTrack) {
	var searchEdit *walk.LineEdit
	var resultTable *walk.TableView
	var dlg *walk.Dialog

	model := &SearchResultModel{items: found}
	allItems := found

	_, err := Dialog{
		AssignTo: &dlg,
		Title:    "Добавить из папки",
		MinSize:  Size{600, 400},
		Layout:   VBox{},
		Children: []Widget{
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{Text: "Поиск:"},
					LineEdit{
						AssignTo: &searchEdit,
						OnTextChanged: func() {
							q := searchEdit.Text()
							if q == "" {
								model.items = allItems
							} else {
								var filtered []FoundTrack
								for _, t := range allItems {
									if fuzzyMatch(q, t.Name) {
										filtered = append(filtered, t)
									}
								}
								model.items = filtered
							}
							model.PublishRowsReset()
							resultTable.SetSuspended(true)
							resultTable.SetSuspended(false)
						},
					},
				},
			},
			TableView{
				AssignTo:       &resultTable,
				Model:          model,
				MultiSelection: true,
				Columns: []TableViewColumn{
					{Title: "Название", Width: 250},
					{Title: "Путь", Width: 350},
				},
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						Text: "Добавить выбранные",
						OnClicked: func() {
							indexes := resultTable.SelectedIndexes()
							if len(indexes) == 0 {
								walk.MsgBox(nil, "Внимание", "Выберите хотя бы один трек", walk.MsgBoxIconWarning)
								return
							}
							for _, i := range indexes {
								if i >= 0 && i < len(model.items) {
									c.addTrack(model.items[i].Name, model.items[i].Path)

								}
							}
							c.refresh()
							searchEdit.SetText("")
						},
					},
					PushButton{
						Text: "Добавить все",
						OnClicked: func() {
							for _, t := range model.items {
								c.addTrack(t.Name, t.Path)
							}
							c.refresh()
							searchEdit.SetText("")
						},
					},
					PushButton{
						Text: "Готово",
						OnClicked: func() {
							dlg.Close(walk.DlgCmdOK)
						},
					},
				},
			},
		},
	}.Run(nil)

	if err != nil {
		walk.MsgBox(nil, "Ошибка", err.Error(), walk.MsgBoxIconError)
	}
}
