package factcheck

import (
	"encoding/json"
	"github.com/bobesa/go-domain-util/domainutil"
	"net/url"
)

type SiteData struct {
	URL         *url.URL
	RawURL      string `json:"url"`
	DisplayName string `json:"display_name"`
	Bias        string `json:"bias"`
	Accuracy    string `json:"accuracy"`
	MBFCURL     string `json:"mbfc_url"`
}

var fcdata = map[string]map[string]*SiteData{}

func LoadData() error {
	// Maybe one day, using Go 1.16, we will be able to embed this file in a nicer way.

	d := []byte(FCData)

	p := struct {
		Sources []*SiteData `json:"sources"`
	}{}

	if err := json.Unmarshal(d, &p); err != nil {
		return err
	}

	for _, s := range p.Sources {
		var err error
		s.URL, err = url.Parse(s.RawURL)

		if err != nil {
			return err
		}

		// Normalize

		if s.URL.Path == "" {
			s.URL.Path = "/"
		}

		s.URL.Host = domainutil.Domain(s.URL.Hostname())

		key := s.URL.Hostname()

		if _, ok := fcdata[key]; !ok {
			fcdata[key] = map[string]*SiteData{}
		}

		fcdata[key][s.URL.Path] = s
	}

	return nil
}
