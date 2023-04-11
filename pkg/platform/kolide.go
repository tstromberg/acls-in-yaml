package platform

import (
	"fmt"
	"strings"

	"github.com/gocarina/gocsv"
)

// KolideUsers parses the CSV file generated by the Kolide Users page.
type KolideUsers struct{}

func (p *KolideUsers) Description() ProcessorDescription {
	return ProcessorDescription{
		Kind: "kolide",
		Name: "Kolide Users",
		Steps: []string{
			"Open https://k2.kolide.com/3361/settings/admin/users",
			"Click CSV",
			"Download resulting CSV file for analysis",
			"Execute 'yacls --kind={{.Kind}} --input={{.Path}}'",
		},
	}
}

type kolideMemberRecord struct {
	Name        string `csv:"Name"`
	Email       string `csv:"Email"`
	Permissions string `csv:"Permissions"`
}

func (p *KolideUsers) Process(c Config) (*Artifact, error) {
	src, err := NewSourceFromConfig(c, p)
	if err != nil {
		return nil, fmt.Errorf("source: %w", err)
	}
	a := &Artifact{Metadata: src}

	records := []kolideMemberRecord{}
	if err := gocsv.UnmarshalBytes(src.content, &records); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	for _, r := range records {
		u := User{
			Account: r.Email,
			Name:    strings.TrimSpace(r.Name),
			Role:    r.Permissions,
		}

		a.Users = append(a.Users, u)
	}

	return a, nil
}
