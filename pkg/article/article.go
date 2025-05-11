package article

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Ignore article chunks if they are shorter than this threshold.
const keepChunkCharThreshold int = 300

var manyNewlinesRe = regexp.MustCompile(`\n{3,}`)

// ArticleReader is a reader of articles on the web.
type ArticleReader struct {
	client http.Client
}

// New creates a new ArticleReader
func NewReader(timeout time.Duration) ArticleReader {
	return ArticleReader{
		client: http.Client{
			Timeout: timeout,
		},
	}
}

// Read reads an article from the web
func (ar ArticleReader) Read(articleUrl string) (string, error) {
	// TODO: handle a PDF link
	validUrl, err := url.Parse(articleUrl)
	if err != nil {
		return "", fmt.Errorf("invalid url %s", articleUrl)
	}
	resp, err := ar.client.Get(validUrl.String())
	if err != nil {
		return "", fmt.Errorf("failed fetching content from url %s: %w", articleUrl, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("got non-ok status code from article request: %d", resp.StatusCode)
	}

	reader, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed creating goquery document reader from response body: %w", err)
	}

	var content string
	reader.Find("article").Each(func(i int, s *goquery.Selection) {
		textChunk := trimAndCleanContent(s.Text())
		if len(textChunk) > keepChunkCharThreshold {
			content += textChunk
		}
	})
	return content, nil
}

// trimAndCleanContent cleans content of whitespace and newlines
func trimAndCleanContent(content string) string {
	content = strings.TrimSpace(content)
	// remove excessively large amounts of newlines
	return strings.TrimSpace(manyNewlinesRe.ReplaceAllString(content, "\n\n"))
}
