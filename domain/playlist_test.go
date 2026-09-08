package domain

import "testing"

func TestParseM3UBasicPaths(t *testing.T) {
	m3u := `#EXTM3U
#EXTINF:-1,Первый трек
C:\Music\song1.mp3
#EXTINF:-1,Второй трек
./music/song2.mp3
/var/media/song3.mp3`

	p := ParseM3U(m3u)

	if len(p.Items) != 3 {
		t.Fatalf("Ожидалось 3 трека, получено %d", len(p.Items))
	}

	// Проверяем имена
	if p.Items[0].Name != "Первый трек" {
		t.Errorf("Неверное имя трека 1: '%s'", p.Items[0].Name)
	}
	if p.Items[1].Name != "Второй трек" {
		t.Errorf("Неверное имя трека 2: '%s'", p.Items[1].Name)
	}

	// Проверяем пути (Path)
	if p.Items[0].Path != `C:\Music\song1.mp3` {
		t.Errorf("Неверный путь 1: '%s'", p.Items[0].Path)
	}
	if p.Items[1].Path != "./music/song2.mp3" {
		t.Errorf("Неверный путь 2: '%s'", p.Items[1].Path)
	}
}
func TestParseM3UWithoutExtInf(t *testing.T) {
	// Треки без EXTINF: имя берётся как имя файла из пути (filepath.Base)
	m3u := `#EXTM3U
C:\NoName1.mp3
./NoName2.mp3`

	p := ParseM3U(m3u)

	if len(p.Items) != 2 {
		t.Fatalf("Ожидалось 2 трека, получено %d", len(p.Items))
	}

	// Теперь имя должно быть только именем файла, а не полным путём
	expectedName1 := "NoName1.mp3"
	if p.Items[0].Name != expectedName1 {
		t.Errorf("Неверное имя трека 1: '%s', ожидалось '%s'", p.Items[0].Name, expectedName1)
	}

	expectedPath1 := "C:\\NoName1.mp3" // Путь остаётся полным
	if p.Items[0].Path != expectedPath1 {
		t.Errorf("Неверный путь трека 1: '%s', ожидалось '%s'", p.Items[0].Path, expectedPath1)
	}

	expectedName2 := "NoName2.mp3"
	if p.Items[1].Name != expectedName2 {
		t.Errorf("Неверное имя трека 2: '%s', ожидалось '%s'", p.Items[1].Name, expectedName2)
	}

	expectedPath2 := "./NoName2.mp3"
	if p.Items[1].Path != expectedPath2 {
		t.Errorf("Неверный путь трека 2: '%s', ожидалось '%s'", p.Items[1].Path, expectedPath2)
	}
}

func TestParseM3UEmptyLinesAndComments(t *testing.T) {
	m3u := "#EXTM3U\n\n# Этот комментарий игнорируется\n\n#EXTINF:120,Трек с паузами\n\nD:\\Songs\\pause.mp3\n\n"

	p := ParseM3U(m3u)

	if len(p.Items) != 1 {
		t.Fatalf("Ожидался 1 трек, получено %d", len(p.Items))
	}
	if p.Items[0].Name != "Трек с паузами" {
		t.Error("Неверное имя при наличии пустых строк и комментариев")
	}
	if p.Items[0].Path != "D:\\Songs\\pause.mp3" {
		t.Error("Неверный путь при наличии пустых строк")
	}
}

func TestPlaylistMethods(t *testing.T) {
	p := &Playlist{}

	// Add
	p.Add("Трек А", "path/a.mp3")
	p.Add("Трек Б", "path/b.mp3")
	if len(p.Items) != 2 || p.Items[1].Name != "Трек Б" {
		t.Error("Add работает неверно")
		return
	}

	// Move
	p.Move(0, 1)
	if p.Items[0].Name != "Трек Б" || p.Items[1].Name != "Трек А" {
		t.Error("Move не поменял треки местами")
	}

	// Remove
	p.Remove(0) // Удаляем "Трек Б"
	if len(p.Items) != 1 || p.Items[0].Name != "Трек А" {
		t.Error("Remove удалил не тот трек")
	}
}

func TestPlaylistToM3U(t *testing.T) {
	p := &Playlist{}
	p.Add("Мой трек", "C:\\MyMusic\\song.mp3")

	result := p.ToM3U()

	expected := "#EXTM3U\nC:\\MyMusic\\song.mp3\n"
	if result != expected {
		t.Errorf("ToM3U вернул неожиданный результат:\n%s\nОжидалось:\n%s", result, expected)
	}
}

func TestParseM3UFileNameFromPath(t *testing.T) {
	m3u := `#EXTM3U
../Треки/Kunteynir/01 - Albums/01 - Эдвард Руки Ножницы Бумага (2004)/08. Всем Хватит (Feat. СевZкваД).mp3
../Треки/Kunteynir/01 - Albums/06 - 5 лет (переиздание 2008) (2013)/08. В Маршрутке Шансон (Feat. Раскольников).mp3`

	p := ParseM3U(m3u)

	if len(p.Items) != 2 {
		t.Fatalf("Ожидалось 2 трека, получено %d", len(p.Items))
	}

	expected1 := "08. Всем Хватит (Feat. СевZкваД).mp3"
	if p.Items[0].Name != expected1 {
		t.Errorf("Имя 1: '%s', ожидалось '%s'", p.Items[0].Name, expected1)
	}

	expected2 := "08. В Маршрутке Шансон (Feat. Раскольников).mp3"
	if p.Items[1].Name != expected2 {
		t.Errorf("Имя 2: '%s', ожидалось '%s'", p.Items[1].Name, expected2)
	}
}
