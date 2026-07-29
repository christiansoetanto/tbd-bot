package util

import (
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
)

const (
	breviaryURL = "https://www.ibreviary.com/m2/breviario.php"
	optionsURL  = "https://www.ibreviary.com/m2/opzioni.php"

	// iBreviary picks the language and hymn variant from the same options form
	// that carries the date, so both have to be sent together.
	breviaryLang = "en"
)

// officeOfReadingsTitle renders the human title for a given day. The daily cron
// and the backfill must agree on this, since it also forms part of the filename.
func officeOfReadingsTitle(date time.Time) string {
	return fmt.Sprintf("Office of Readings for %s", date.Format("Monday, 02 January 2006"))
}

// OfficeOfReadingsFilename is the path committed to the office-of-readings repo.
func OfficeOfReadingsFilename(date time.Time) string {
	return fmt.Sprintf("%s-%s.md", date.Format("2006-01-02"), officeOfReadingsTitle(date))
}

// ParseOfficeOfReadings extracts the second reading from an iBreviary page. The
// date is supplied rather than read from the clock so the same parser serves
// both the daily cron and backfills of past days.
func ParseOfficeOfReadings(r io.Reader, date time.Time) (title string, text string, err error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return "", "", err
	}

	title = officeOfReadingsTitle(date)
	doc.Find(".rubrica").Each(func(i int, s *goquery.Selection) {
		// Prefix rather than equality: on Memorials carrying a proper-reading
		// rubric, iBreviary leaves this span open and folds the rubric text into
		// it, so the element reads "SECOND READING<br><br>The proper reading...".
		// Requiring an exact match skipped those days entirely.
		if !strings.HasPrefix(strings.TrimSpace(s.Text()), "SECOND READING") {
			return
		}
		parent := s.Parent()
		converter := md.NewConverter("", true, nil)
		markdown := converter.Convert(parent)

		// Cut the Responsory, which follows the reading and is not wanted.
		text = strings.Split(markdown, "RESPONSORY")[0]
		secondReadingSplit := strings.Split(text, "SECOND READING")
		if len(secondReadingSplit) < 2 {
			text = ""
			return
		}
		text = secondReadingSplit[1]
	})

	if text == "" {
		return "", "", errors.New("no second reading found")
	}
	return title, text, nil
}

// GetOfficeOfReadingsText fetches today's office of readings.
func GetOfficeOfReadingsText() (title string, text string, err error) {
	res, err := http.Get(breviaryURL + "?s=ufficio_delle_letture")
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", "", fmt.Errorf("ibreviary returned HTTP %d", res.StatusCode)
	}
	return ParseOfficeOfReadings(res.Body, time.Now())
}

// GetOfficeOfReadingsTextForDate fetches the office of readings for an arbitrary
// day. iBreviary has no date query parameter — the chosen day lives in the PHP
// session — so this posts the date to the options form first, then reads the
// office back over the same cookie jar.
func GetOfficeOfReadingsTextForDate(date time.Time) (title string, text string, err error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return "", "", err
	}
	client := &http.Client{Jar: jar, Timeout: 30 * time.Second}

	// Establish a session so there is a PHPSESSID to attach the date to.
	initRes, err := client.Get(breviaryURL)
	if err != nil {
		return "", "", fmt.Errorf("session init: %w", err)
	}
	initRes.Body.Close()

	form := url.Values{
		"lang":   {breviaryLang},
		"giorno": {fmt.Sprintf("%d", date.Day())},
		"mese":   {fmt.Sprintf("%d", int(date.Month()))},
		"anno":   {fmt.Sprintf("%d", date.Year())},
		"ok":     {"ok"},
	}
	optRes, err := client.PostForm(optionsURL, form)
	if err != nil {
		return "", "", fmt.Errorf("set date %s: %w", date.Format("2006-01-02"), err)
	}
	optRes.Body.Close()
	if optRes.StatusCode != 200 {
		return "", "", fmt.Errorf("set date %s: HTTP %d", date.Format("2006-01-02"), optRes.StatusCode)
	}

	res, err := client.Get(breviaryURL + "?s=ufficio_delle_letture")
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", "", fmt.Errorf("ibreviary returned HTTP %d for %s", res.StatusCode, date.Format("2006-01-02"))
	}

	return ParseOfficeOfReadings(res.Body, date)
}

func GenerateOfficeOfReadingsEmbeds() ([]*discordgo.MessageEmbed, error) {
	title, text, err := GetOfficeOfReadingsText()
	if err != nil {
		return nil, err
	}
	var embeds []*discordgo.MessageEmbed
	const lengthLimit = 3000
	isContinueFromBefore := false
	for len(text) > 0 {
		min := int(math.Min(float64(len(text)), lengthLimit))
		chunk := text[0:min]
		currentTitle := title
		if isContinueFromBefore {
			currentTitle = currentTitle + " (continued)"
			chunk = "..." + chunk
		}
		if len(text) > lengthLimit {
			chunk = chunk + "... (continued)"
			isContinueFromBefore = true
		} else {
			isContinueFromBefore = false
		}

		embed := EmbedBuilder(currentTitle, chunk)
		embeds = append(embeds, embed)
		text = text[min:]
	}
	return embeds, nil
}
