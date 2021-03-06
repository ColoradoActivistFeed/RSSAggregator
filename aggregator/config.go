package aggregator

type Config struct {
	Name                   string
	Description            string
	Link                   string
	AutoCommit             bool
	Organizations          map[string]Organization
	RefreshIntervalSeconds int64
	TemplatePath           string
	OutputPath             string
	StaticAssets           string
	TimeZone               string
}

type Organization struct {
	Description string
	Author      string
	Slug        string
	Link        string
	Sources     []string
}
