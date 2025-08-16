package scrapper

func NewWebScraper() WebScraper {
	return NewCollyScraper()
}

func NewWebScraperWithDebug() WebScraper {
	return NewCollyScraperWithDebug()
}
