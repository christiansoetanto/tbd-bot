package util

import (
	"errors"
	"fmt"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/bwmarrin/discordgo"
	"math"
	"net/http"
	"strings"
	"time"
)

func GetOfficeOfReadingsText() (title string, text string, err error) {
	res, err := http.Get("https://ibreviary.com/m2/breviario.php?s=ufficio_delle_letture")
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", "", err
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", "", err
	}

	title = fmt.Sprintf("Office of Readings for %s", time.Now().Format("Monday, 02 January 2006"))
	doc.Find(".rubrica").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title
		if s.Text() == "SECOND READING" {
			parent := s.Parent()
			converter := md.NewConverter("", true, nil)
			markdown := converter.Convert(parent)
			if err != nil {
				return
			}
			//cut Responsory
			responsorySplitted := strings.Split(markdown, "RESPONSORY")
			text = responsorySplitted[0]
			secondReadingSplitted := strings.Split(text, "SECOND READING")
			text = secondReadingSplitted[1]
		}
	})

	if text == "" {
		return "", "", errors.New("no second reading found")
	}
	return title, text, nil
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
