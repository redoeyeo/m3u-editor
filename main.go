// main.go
package main

import (
	"github.com/redoeyeo/m3u-editor/domain"
	"github.com/redoeyeo/m3u-editor/ui"
)

func main() {
	// Создаем пустую плейлист-структуру
	pl := &domain.Playlist{}

	// Запускаем GUI, передавая ему нашу структуру
	if err := ui.Run(pl); err != nil {
		panic(err)
	}
}
