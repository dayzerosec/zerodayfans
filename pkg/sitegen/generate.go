package sitegen

import (
	"encoding/json"
	"errors"
	"github.com/dayzerosec/zerodayfans/pkg/config"
	"github.com/dayzerosec/zerodayfans/pkg/enrichment"
	"html/template"
	"log"
	"os"
	"path/filepath"
)

func Generate() error {
	log.Printf("Generating %s", config.Cfg.Output.Title)
	entries, err := loadRaw()
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		return errors.New("no entries for website")
	}

	_, err = checkDirectory(config.Cfg.Output.Webroot, false)
	if err != nil {
		return err
	}

	tmplDir, err := checkDirectory(config.Cfg.TemplatesDir, true)
	if err != nil {
		return err
	}

	data := SiteData{
		Output:  config.Cfg.Output,
		Entries: entries,
	}

	for _, key := range config.Cfg.Sidebar.Order {
		if content, ok := config.Cfg.Sidebar.Items[key]; ok {
			data.Sidebar = append(data.Sidebar, content)
		} else {
			log.Printf("sidebar item not found: %s", key)
		}
	}

	tmpl := template.New("---").Funcs(tmplFuncs)
	if tmpl, err = tmpl.ParseGlob(tmplDir + "/*.html"); err != nil {
		return err
	}
	if tmpl, err = tmpl.ParseGlob(tmplDir + "/*.tmpl"); err != nil {
		return err
	}

	// Find all the *.html files from the template directory
	var htmlFiles []string
	err = filepath.Walk(tmplDir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".html" {
			htmlFiles = append(htmlFiles, filepath.Base(path))
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Execute the found templates
	for _, file := range htmlFiles {
		log.Println("Generating file:", file)
		if err = executeTemplate(tmpl, file, data); err != nil {
			return err
		}
	}

	if err = createFeedList(); err != nil {
		log.Printf("Error creating feed list: %v", err)
	}

	if err = copyStaticFiles(); err != nil {
		log.Printf("Error copying static files: %v", err)

	}

	return nil
}

type SiteData struct {
	Output  config.OutputConfig
	Entries []enrichment.EnrichedData
	Sidebar []config.SidebarContentConfig
}

func executeTemplate(tmpl *template.Template, name string, data SiteData) error {
	if fp, err := os.Create(filepath.Join(config.Cfg.Output.Webroot, name)); err == nil {
		defer func() { _ = fp.Close() }()
		err = tmpl.ExecuteTemplate(fp, name, data)
		if err != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}

func checkDirectory(dir string, mustExist bool) (string, error) {
	var err error

	dir, err = filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	stat, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) && !mustExist {
			if err = os.MkdirAll(dir, 0755); err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	} else if !stat.IsDir() {
		return "", errors.New("directory is not a directory")
	}

	return dir, nil
}

func loadRaw() ([]enrichment.EnrichedData, error) {
	fp, err := os.Open(config.Cfg.Output.RawFile)
	if err != nil {
		return nil, err
	}
	defer func() { _ = fp.Close() }()

	var rawFeed []enrichment.EnrichedData
	if err = json.NewDecoder(fp).Decode(&rawFeed); err != nil {
		return nil, err
	}

	return rawFeed, nil
}
