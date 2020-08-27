package aggregator

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

func (a *Aggregator) Fetch() error {

	a.Content = map[string]*feeds.Feed{}
	fp := gofeed.NewParser()
	for name, f := range a.Config.Organizations {

		loc, err := time.LoadLocation("America/Denver")
		if err != nil {
			return err
		}

		now := time.Now().In(loc)
		content := &feeds.Feed{
			Title:       name,
			Link:        &feeds.Link{Href: f.Link},
			Description: f.Description,
			Author:      &feeds.Author{Name: f.Author, Email: ""},
			Items:       []*feeds.Item{},
			Created:     now,
		}

		for _, url := range f.Sources {
			remoteRSS, err := fp.ParseURL(url)
			if err != nil {
				return fmt.Errorf("failed to download content from %s, %s", url, err.Error())
			}
			for _, i := range remoteRSS.Items {
				f := &feeds.Item{
					Title:       i.Title,
					Link:        &feeds.Link{Href: i.Link},
					Author:      &feeds.Author{Name: f.Author, Email: ""},
					Description: i.Description,
					Id:          i.GUID,
					Content:     i.Content,
					Created:     now,
				}
				if i.UpdatedParsed != nil {
					f.Updated = i.UpdatedParsed.In(loc)
				}
				if i.PublishedParsed != nil {
					f.Created = i.PublishedParsed.In(loc)
				}
				content.Items = append(content.Items, f)
			}
			log.WithField("url", url).Info("downloaded feed")
		}
		content.Sort(func(a, b *feeds.Item) bool {
			return a.Created.After(b.Created)
		})

		a.Content[name] = content
	}

	// generate all feed
	content := &feeds.Feed{
		Title:       a.Config.Name,
		Description: a.Config.Description,
		Created:     time.Now(),
		Items:       []*feeds.Item{},
		Author:      &feeds.Author{Name: "", Email: ""},
		Link:        &feeds.Link{Href: ""},
	}
	for _, i := range a.Content {
		for _, c := range i.Items {
			content.Items = append(content.Items, c)
		}
	}
	content.Sort(func(a, b *feeds.Item) bool {
		return a.Created.After(b.Created)
	})
	a.Content["All"] = content

	return nil
}

func (a *Aggregator) WriteRSS() error {

	for name, content := range a.Content {

		slug := strings.ToLower(name)
		if org, ok := a.Config.Organizations[name]; ok {
			slug = org.Slug
		}

		for _, format := range []string{"rss", "atom", "json"} {

			path := fmt.Sprintf("%s/%s/%s.xml", a.Config.OutputPath, format, slug)
			if format == "json" {
				path = fmt.Sprintf("%s/%s/%s.json", a.Config.OutputPath, format, slug)
			}
			f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
			if err != nil {
				return fmt.Errorf("failed to open %s, %s", path, err.Error())
			}

			switch format {
			case "rss":
				err = content.WriteRss(f)
			case "atom":
				err = content.WriteAtom(f)
			case "json":
				err = content.WriteJSON(f)
			}
			_ = f.Close()
			if err != nil {
				return fmt.Errorf("failed to write %s, %s", path, err.Error())
			}

			log.WithField("name", name).WithField("path", path).Info("wrote rss")
		}

	}

	return nil
}
