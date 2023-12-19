package platform

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// WebflowMembers parses the CSV file generated by the OnePassword Team page.
type WebflowMembers struct{}

func (p *WebflowMembers) Description() ProcessorDescription {
	return ProcessorDescription{
		Kind: "webflow",
		Name: "Webflow Site Permissions",
		Steps: []string{
			"Open https://webflow.com/dashboard/sites/<site>/members",
			"Save this page (Complete)",
			"Collect resulting .html file for analysis (the other files are not necessary)",
			"Execute 'yacls --kind={{.Kind}} --input={{.Path}}'",
		},
		MatchingFilename: regexp.MustCompile(`webflow.*html$`),
	}
}

func (p *WebflowMembers) Process(c Config) (*Artifact, error) {
	src, err := NewSourceFromConfig(c, p)
	if err != nil {
		return nil, fmt.Errorf("source: %w", err)
	}

	a := &Artifact{Metadata: src}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(src.content))
	if err != nil {
		return nil, fmt.Errorf("document: %w", err)
	}

	table := doc.Find("table[data-automation-id=Workspaces__membersTable]").First()

	// Find the members
	table.Find("tr").Each(func(i int, tr *goquery.Selection) {
		name := ""
		account := ""
		role := ""

		tr.Find("p").Each(func(i int, p *goquery.Selection) {
			attr, _ := p.Attr("data-automation-id")
			log.Printf("p=%s, attr=%s", p.Text(), attr)

			if strings.HasPrefix(attr, "memberCellDisplayName") {
				log.Printf("member cell=%s", p.Text())
				name, _, _ = strings.Cut(p.Text(), "(")
				name = strings.TrimSpace(name)
				log.Printf("name=%s", name)
			} else {
				account, _, _ = strings.Cut(p.Text(), "(")
				account = strings.TrimSpace(account)
				log.Printf("account=%s", account)
			}
		})

		tr.Find("span").Each(func(i int, s *goquery.Selection) {
			class, _ := s.Attr("class")
			if class == "" {
				role, _, _ = strings.Cut(s.Text(), "(")
				role = strings.TrimSpace(role)
			}
		})

		if account != "" {
			a.Users = append(a.Users, User{Account: account, Name: name, Role: role})
		}
	})

	return a, nil
}
