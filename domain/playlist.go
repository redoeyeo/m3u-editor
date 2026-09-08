package domain

import (
	"path/filepath"
	"strings"
)

type Track struct {
	Name string
	Path string // Было URL, стало Path
}

type Playlist struct {
	Items []Track
}

func (p *Playlist) Add(name, path string) {
	p.Items = append(p.Items, Track{Name: name, Path: path})
}

func (p *Playlist) Remove(i int) {
	if i < 0 || i >= len(p.Items) {
		return
	}
	p.Items = append(p.Items[:i], p.Items[i+1:]...)
}

func (p *Playlist) Move(i, j int) {
	if i < 0 || j < 0 || i >= len(p.Items) || j >= len(p.Items) {
		return
	}
	p.Items[i], p.Items[j] = p.Items[j], p.Items[i]
}

func ParseM3U(text string) *Playlist {
	p := &Playlist{}
	lines := strings.Split(text, "\n")
	var pendingName string

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#EXTINF:") {
			comma := strings.Index(line, ",")
			if comma != -1 {
				pendingName = line[comma+1:]
			} else {
				pendingName = ""
			}
		} else if strings.HasPrefix(line, "#") {
			continue
		} else {
			name := pendingName
			if name == "" {
				name = filepath.Base(line) // ← было: name = line
			}
			p.Add(name, line)
			pendingName = ""
		}
	}
	return p
}

func (p *Playlist) ToM3U() string {
	var b strings.Builder
	b.WriteString("#EXTM3U\n")
	for _, t := range p.Items {
		b.WriteString(t.Path) // Выводим Path
		b.WriteString("\n")
	}
	return b.String()
}
