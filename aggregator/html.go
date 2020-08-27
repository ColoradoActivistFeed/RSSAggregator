package aggregator

import (
	"fmt"
	"github.com/gorilla/feeds"
	log "github.com/sirupsen/logrus"
	"html/template"
	"os"
	"strings"

	"github.com/jaytaylor/html2text"

)

func htmlToString(s string) string {
	s, _ = html2text.FromString(s, html2text.Options{OmitLinks: true})

	if len(s) > 200 {
		s = fmt.Sprintf("%s...", s[0:199])
	}

	return s
}



func (a *Aggregator) WriteHTML() error {

	templates := template.New("templates")
	funcs := template.FuncMap{
		"htmlToString": htmlToString,
	}
	templates, err := templates.Funcs(funcs).ParseGlob(fmt.Sprintf("%s/*.html", a.Config.TemplatePath))
	if err != nil {
		return err
	}

	//content, ok := a.Content["All"]
	//if !ok {
	//	return fmt.Errorf("could not find All content category")
	//}
	//
	//path := fmt.Sprintf("%s/index.html", a.Config.OutputPath)
	//f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	//if err != nil {
	//	return fmt.Errorf("failed to open %s, %s", path, err.Error())
	//}
	//defer func() { _ = f.Close() }()
	//
	//err = templates.ExecuteTemplate(f, "index", struct {
	//	Title string
	//	Description string
	//	Content *feeds.Feed
	//	Organizations map[string]Organization
	//}{
	//	Title: a.Config.Name,
	//	Description: a.Config.Description,
	//	Content: content,
	//	Organizations: a.Config.Organizations,
	//})
	//if err != nil {
	//	return err
	//}
	//
	//log.WithField("name", "All").WithField("path", path).Info("wrote rss")

	// write individual feeds
	for name, content := range a.Content {

		path := fmt.Sprintf("%s/index.html", a.Config.OutputPath)
		org := Organization{
			Description: a.Config.Description,
			Author:      "All Sources",
			Slug:        "all",
		}
		if strings.ToLower(name) != "all" {
			orgFromConfig, ok := a.Config.Organizations[name]
			if !ok {
				return fmt.Errorf("could not find organization %s", name)
			}
			path = fmt.Sprintf("%s/%s.html", a.Config.OutputPath, orgFromConfig.Slug)
			org = orgFromConfig
		}

		f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("failed to open %s, %s", path, err.Error())
		}
		defer func() { _ = f.Close() }()

		err = templates.ExecuteTemplate(f, "index", struct {
			Title string
			Description string
			Content *feeds.Feed
			Organizations map[string]Organization
			Org Organization
		}{
			Title: a.Config.Name,
			Description: a.Config.Description,
			Content: content,
			Organizations: a.Config.Organizations,
			Org: org,
		})
		if err != nil {
			return err
		}

		log.WithField("name", name).WithField("path", path).Info("wrote rss")

	}

	for _, templateName := range []string{"sources", "about", "contact"} {
		path := fmt.Sprintf("%s/%s.html", a.Config.OutputPath, templateName)
		f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("failed to open %s, %s", path, err.Error())
		}
		defer func() { _ = f.Close() }()

		err = templates.ExecuteTemplate(f, templateName, struct {
			Title string
			Description string
			Content *feeds.Feed
			Organizations map[string]Organization
			Org Organization
		}{
			Title: a.Config.Name,
			Description: a.Config.Description,
			Organizations: a.Config.Organizations,
		})
		if err != nil {
			return err
		}

		log.WithField("name", templateName).WithField("path", path).Info("wrote rss")
	}

	return nil
}

