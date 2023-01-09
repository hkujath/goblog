package page

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/russross/blackfriday/v2"
	"html/template"
	"os"
	"path/filepath"
	"time"
)

// Page represents a website page.
type Page struct {
	Title      string
	LastChange time.Time
	Content    template.HTML
	Comments   []Comment
}

type Pages []Page
type Comment struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// LoadPage creates a Page object from data of a given file path.
func LoadPage(fPath string) (Page, error) {
	var p Page
	fi, err := os.Stat(fPath)
	if err != nil {
		return p, fmt.Errorf("LoadPage: %w", err)
	}

	p.Title = fi.Name()
	p.LastChange = fi.ModTime()
	p.Comments, err = LoadComments(p.Title)
	if err != nil {
		return p, fmt.Errorf("loadPage.loadComments: %w", err)
	}

	b, err := os.ReadFile(fPath)
	if err != nil {
		return p, fmt.Errorf("LoadPage.ReadFile: %w", err)
	}
	p.Content = template.HTML(blackfriday.Run(b, blackfriday.WithNoExtensions()))

	return p, nil
}

// LoadPages loads multiple pages from a give source folder
func LoadPages(src string) (Pages, error) {
	var ps Pages
	fs, err := os.ReadDir(src)
	if err != nil {
		return ps, fmt.Errorf("LoadPages.ReadDir: %w", err)
	}

	for _, f := range fs {
		if f.IsDir() {
			continue
		}

		fPath := filepath.Join(src, f.Name())
		p, err := LoadPage(fPath)
		if err != nil {
			return ps, fmt.Errorf("LoadPages.loadPage: %w", err)
		}
		ps = append(ps, p)
	}
	return ps, nil
}

func ParseFiles(content, tmplFolder string) (*template.Template, error) {
	wd, _ := os.Getwd()
	fmt.Printf("Current WD: [%s]\n", wd)

	return template.ParseFiles(
		filepath.Join(wd, tmplFolder, "base.tmpl.html"),
		filepath.Join(wd, tmplFolder, "header.tmpl.html"),
		filepath.Join(wd, tmplFolder, "footer.tmpl.html"),
		filepath.Join(wd, tmplFolder, "comment.tmpl.html"),
		filepath.Join(wd, tmplFolder, content),
	)
}

// SaveComments saves the given comments into a json file.
func SaveComments(title string, cs []Comment) error {
	wd, _ := os.Getwd()
	fPath := filepath.Join(wd, "comments", title+".json")
	f, err := os.OpenFile(fPath, os.O_CREATE|os.O_WRONLY, 0555)

	if err != nil {
		return fmt.Errorf("saveComments: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	return enc.Encode(cs)
}

// LoadComments loads the comments for a given page name
func LoadComments(title string) ([]Comment, error) {
	var cs []Comment
	wd, _ := os.Getwd()
	fPath := filepath.Join(wd, "comments", title+".json")
	f, err := os.Open(fPath)

	if errors.Is(err, os.ErrNotExist) {
		return cs, nil
	}

	if err != nil {
		return cs, fmt.Errorf("loadCommants: %w", err)
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	err = dec.Decode(&cs)
	return cs, err
}
