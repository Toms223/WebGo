package app

import "strings"

type TwitterCard struct {
	Card string

	Site string

	Creator string

	Title string

	Description string

	Image string

	ImageAlt string
}

func (c TwitterCard) toMap() map[string]string {
	m := make(map[string]string)

	if c.Card != "" {
		m["twitter:card"] = c.Card
	}

	if c.Site != "" {
		if !strings.HasPrefix(c.Site, "@") {
			c.Site = "@" + c.Site
		}
		m["twitter:site"] = c.Site
	}

	if c.Creator != "" {
		if !strings.HasPrefix(c.Creator, "@") {
			c.Creator = "@" + c.Creator
		}
		m["twitter:creator"] = c.Creator
	}

	if c.Title != "" {
		m["twitter:title"] = c.Title
	}

	if c.Description != "" {
		m["twitter:description"] = c.Description
	}

	if c.Image != "" {
		m["twitter:image"] = c.Image
	}

	if c.ImageAlt != "" {
		m["twitter:image:alt"] = c.ImageAlt
	}

	return m
}
