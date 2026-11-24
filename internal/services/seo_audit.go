package services

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
)

// SEOAuditResult contains the scraped metadata and audit analysis for a URL.
type SEOAuditResult struct {
	URL                string   `json:"url"`
	FinalURL           string   `json:"final_url"`
	StatusCode         int      `json:"status_code"`
	Title              string   `json:"title"`
	TitleLength        int      `json:"title_length"`
	TitleOptimal       bool     `json:"title_optimal"`
	Description        string   `json:"description"`
	DescriptionLength  int      `json:"description_length"`
	DescriptionOptimal bool     `json:"description_optimal"`
	H1Tags             []string `json:"h1_tags"`
	H1Count            int      `json:"h1_count"`
	H1Primary          string   `json:"h1_primary"`
	H2Tags             []string `json:"h2_tags"`
	Canonical          string   `json:"canonical"`
	Robots             string   `json:"robots"`
	ImagesTotal        int      `json:"images_total"`
	ImagesWithoutAlt   int      `json:"images_without_alt"`
	Score              int      `json:"score"`
	Issues             []string `json:"issues"`
	Passed             []string `json:"passed"`
}

// SEOAuditService provides URL scraping and SEO audit capabilities.
type SEOAuditService struct {
	client *http.Client
}

// NewSEOAuditService creates a new instance of SEOAuditService.
func NewSEOAuditService() *SEOAuditService {
	return &SEOAuditService{
		client: &http.Client{
			Timeout: 12 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return errors.New("stopped after 10 redirects")
				}
				return nil
			},
		},
	}
}

// AuditURL scrapes the title, meta description, and H1 tags from the provided URL,
// evaluating key SEO factors and returning the structured audit result.
func (s *SEOAuditService) AuditURL(rawURL string) (*SEOAuditResult, error) {
	trimmedURL := strings.TrimSpace(rawURL)
	if trimmedURL == "" {
		return nil, errors.New("empty URL provided")
	}

	if !strings.HasPrefix(trimmedURL, "http://") && !strings.HasPrefix(trimmedURL, "https://") {
		trimmedURL = "http://" + trimmedURL
	}

	parsedURL, err := url.ParseRequestURI(trimmedURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL format: %w", err)
	}

	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; SEOnautBot/1.0; +https://seonaut.org/bot)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	// Read body with a 5MB size limit
	bodyLimit := int64(5 * 1024 * 1024)
	limitedReader := io.LimitReader(resp.Body, bodyLimit)
	bodyBytes, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	doc, err := htmlquery.Parse(bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML document: %w", err)
	}

	result := &SEOAuditResult{
		URL:        parsedURL.String(),
		FinalURL:   resp.Request.URL.String(),
		StatusCode: resp.StatusCode,
		H1Tags:     []string{},
		H2Tags:     []string{},
		Issues:     []string{},
		Passed:     []string{},
	}

	// 1. Scrape Title Tag
	titleNode, err := htmlquery.Query(doc, "//title")
	if err == nil && titleNode != nil {
		result.Title = strings.TrimSpace(htmlquery.InnerText(titleNode))
	}
	result.TitleLength = len([]rune(result.Title))
	result.TitleOptimal = result.TitleLength >= 50 && result.TitleLength <= 60

	// 2. Scrape Meta Description
	descNode, err := htmlquery.Query(doc, "//meta[translate(@name, 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz')='description']/@content")
	if err == nil && descNode != nil {
		result.Description = strings.TrimSpace(htmlquery.SelectAttr(descNode, "content"))
	}
	result.DescriptionLength = len([]rune(result.Description))
	result.DescriptionOptimal = result.DescriptionLength >= 120 && result.DescriptionLength <= 155

	// 3. Scrape All H1 Tags
	h1Nodes, err := htmlquery.QueryAll(doc, "//h1")
	if err == nil {
		for _, node := range h1Nodes {
			text := cleanText(htmlquery.InnerText(node))
			if text != "" {
				result.H1Tags = append(result.H1Tags, text)
			}
		}
	}
	result.H1Count = len(result.H1Tags)
	if result.H1Count > 0 {
		result.H1Primary = result.H1Tags[0]
	}

	// 4. Scrape H2 Tags for secondary structure analysis
	h2Nodes, err := htmlquery.QueryAll(doc, "//h2")
	if err == nil {
		for i, node := range h2Nodes {
			if i >= 10 { // limit to top 10
				break
			}
			text := cleanText(htmlquery.InnerText(node))
			if text != "" {
				result.H2Tags = append(result.H2Tags, text)
			}
		}
	}

	// 5. Scrape Canonical Tag
	canonicalNode, err := htmlquery.Query(doc, "//link[@rel='canonical']/@href")
	if err == nil && canonicalNode != nil {
		result.Canonical = strings.TrimSpace(htmlquery.SelectAttr(canonicalNode, "href"))
	}

	// 6. Scrape Robots Meta
	robotsNode, err := htmlquery.Query(doc, "//meta[translate(@name, 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz')='robots']/@content")
	if err == nil && robotsNode != nil {
		result.Robots = strings.TrimSpace(htmlquery.SelectAttr(robotsNode, "content"))
	}

	// 7. Scrape Images and check Alt attributes
	imgNodes, err := htmlquery.QueryAll(doc, "//img")
	if err == nil {
		result.ImagesTotal = len(imgNodes)
		for _, img := range imgNodes {
			altVal := ""
			for _, attr := range img.Attr {
				if strings.ToLower(attr.Key) == "alt" {
					altVal = strings.TrimSpace(attr.Val)
					break
				}
			}
			if altVal == "" {
				result.ImagesWithoutAlt++
			}
		}
	}

	// 8. Calculate SEO Audit Score & Checklists
	score := 0

	// Title Evaluation
	if result.Title == "" {
		result.Issues = append(result.Issues, "Missing Title Tag: Search engines require a descriptive <title> tag.")
	} else if result.TitleOptimal {
		score += 30
		result.Passed = append(result.Passed, fmt.Sprintf("Title tag length is optimal (%d characters).", result.TitleLength))
	} else if result.TitleLength < 50 {
		score += 15
		result.Issues = append(result.Issues, fmt.Sprintf("Title tag is short (%d characters). Ideal is 50–60 characters.", result.TitleLength))
	} else {
		score += 15
		result.Issues = append(result.Issues, fmt.Sprintf("Title tag is too long (%d characters) and may truncate on SERP.", result.TitleLength))
	}

	// Meta Description Evaluation
	if result.Description == "" {
		result.Issues = append(result.Issues, "Missing Meta Description: A meta description improves click-through rates.")
	} else if result.DescriptionOptimal {
		score += 30
		result.Passed = append(result.Passed, fmt.Sprintf("Meta description length is optimal (%d characters).", result.DescriptionLength))
	} else if result.DescriptionLength < 120 {
		score += 15
		result.Issues = append(result.Issues, fmt.Sprintf("Meta description is brief (%d characters). Ideal is 120–155 characters.", result.DescriptionLength))
	} else {
		score += 15
		result.Issues = append(result.Issues, fmt.Sprintf("Meta description is too long (%d characters) and will truncate on SERP.", result.DescriptionLength))
	}

	// H1 Heading Evaluation
	if result.H1Count == 0 {
		result.Issues = append(result.Issues, "Missing H1 Tag: Each page must have exactly one prominent <h1> heading.")
	} else if result.H1Count == 1 {
		score += 25
		result.Passed = append(result.Passed, fmt.Sprintf("Single primary H1 heading detected: \"%s\".", result.H1Primary))
	} else {
		score += 10
		result.Issues = append(result.Issues, fmt.Sprintf("Multiple H1 tags detected (%d found). Best practice is exactly one <h1> per page.", result.H1Count))
	}

	// Image Alt Text Evaluation
	if result.ImagesTotal > 0 {
		if result.ImagesWithoutAlt == 0 {
			score += 15
			result.Passed = append(result.Passed, fmt.Sprintf("All %d images contain descriptive alt text.", result.ImagesTotal))
		} else {
			score += int(15 * float64(result.ImagesTotal-result.ImagesWithoutAlt) / float64(result.ImagesTotal))
			result.Issues = append(result.Issues, fmt.Sprintf("%d out of %d images are missing alt attributes.", result.ImagesWithoutAlt, result.ImagesTotal))
		}
	} else {
		score += 15
		result.Passed = append(result.Passed, "Clean visual structure without broken image tags.")
	}

	if score > 100 {
		score = 100
	}
	result.Score = score

	return result, nil
}

// ScrapeHTMLContent parses raw HTML directly (useful for testing or embedded scraping).
func (s *SEOAuditService) ScrapeHTMLContent(targetURL string, htmlContent string) (*SEOAuditResult, error) {
	doc, err := htmlquery.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	result := &SEOAuditResult{
		URL:        targetURL,
		FinalURL:   targetURL,
		StatusCode: 200,
		H1Tags:     []string{},
		H2Tags:     []string{},
		Issues:     []string{},
		Passed:     []string{},
	}

	titleNode, _ := htmlquery.Query(doc, "//title")
	if titleNode != nil {
		result.Title = strings.TrimSpace(htmlquery.InnerText(titleNode))
	}
	result.TitleLength = len([]rune(result.Title))
	result.TitleOptimal = result.TitleLength >= 50 && result.TitleLength <= 60

	descNode, _ := htmlquery.Query(doc, "//meta[translate(@name, 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz')='description']/@content")
	if descNode != nil {
		result.Description = strings.TrimSpace(htmlquery.SelectAttr(descNode, "content"))
	}
	result.DescriptionLength = len([]rune(result.Description))
	result.DescriptionOptimal = result.DescriptionLength >= 120 && result.DescriptionLength <= 155

	h1Nodes, _ := htmlquery.QueryAll(doc, "//h1")
	for _, node := range h1Nodes {
		text := cleanText(htmlquery.InnerText(node))
		if text != "" {
			result.H1Tags = append(result.H1Tags, text)
		}
	}
	result.H1Count = len(result.H1Tags)
	if result.H1Count > 0 {
		result.H1Primary = result.H1Tags[0]
	}

	return result, nil
}

func cleanText(s string) string {
	fields := strings.Fields(s)
	return strings.Join(fields, " ")
}
