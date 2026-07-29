// Command backfill fills gaps in the office-of-readings repository.
//
// iBreviary serves the liturgy of a day chosen in the PHP session, so each date
// costs two requests: one to set the date, one to read the office. Requests are
// paced by --delay and run strictly one at a time; there is no concurrency here
// on purpose, because the point is to be gentle with a small volunteer-run site.
//
// Typical staged use, checking the result between runs:
//
//	go run ./cmd/backfill --days 1  --dry-run
//	go run ./cmd/backfill --days 1
//	go run ./cmd/backfill --days 7
//	go run ./cmd/backfill --from 2025-09-20
package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/christiansoetanto/tbd-bot/util"
)

const (
	repoOwner = "christiansoetanto"
	repoName  = "office-of-readings"
)

type config struct {
	from        string
	to          string
	days        int
	delay       time.Duration
	dryRun      bool
	force       bool
	limit       int
	githubToken string
}

func main() {
	var cfg config
	flag.StringVar(&cfg.from, "from", "", "oldest date to backfill, YYYY-MM-DD (mutually exclusive with --days)")
	flag.StringVar(&cfg.to, "to", "", "newest date to backfill, YYYY-MM-DD (default: yesterday)")
	flag.IntVar(&cfg.days, "days", 0, "backfill this many days ending at --to, walking backwards")
	flag.DurationVar(&cfg.delay, "delay", 6*time.Second, "pause between dates; keep this generous, iBreviary is a small site")
	flag.BoolVar(&cfg.dryRun, "dry-run", false, "fetch and parse but write nothing to GitHub")
	flag.BoolVar(&cfg.force, "force", false, "overwrite a date that already has a file (default is to skip)")
	flag.IntVar(&cfg.limit, "limit", 0, "stop after this many dates (0 = no limit)")
	flag.Parse()

	cfg.githubToken = os.Getenv("GITHUBAPITOKEN")
	if cfg.githubToken == "" && !cfg.dryRun {
		log.Fatal("GITHUBAPITOKEN is not set; export it or pass --dry-run")
	}

	dates, err := resolveDates(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if len(dates) == 0 {
		log.Fatal("no dates selected; pass --days or --from")
	}
	if cfg.limit > 0 && len(dates) > cfg.limit {
		dates = dates[:cfg.limit]
	}

	mode := "WRITING to GitHub"
	if cfg.dryRun {
		mode = "DRY RUN, nothing will be written"
	}
	log.Printf("%d date(s), %s -> %s, %v between requests. %s.",
		len(dates),
		dates[0].Format("2006-01-02"),
		dates[len(dates)-1].Format("2006-01-02"),
		cfg.delay, mode)
	log.Printf("estimated wall time: ~%s", (time.Duration(len(dates)) * cfg.delay).Round(time.Second))

	var written, skipped, failed int
	for i, date := range dates {
		if i > 0 {
			time.Sleep(cfg.delay)
		}
		status, err := processDate(cfg, date)
		switch {
		case err != nil:
			failed++
			log.Printf("[%3d/%d] %s  FAILED: %v", i+1, len(dates), date.Format("2006-01-02"), err)
		case status == "skipped":
			skipped++
			log.Printf("[%3d/%d] %s  skipped, already present", i+1, len(dates), date.Format("2006-01-02"))
		default:
			written++
			log.Printf("[%3d/%d] %s  %s", i+1, len(dates), date.Format("2006-01-02"), status)
		}
	}

	log.Printf("done: %d written, %d skipped, %d failed", written, skipped, failed)
	if failed > 0 {
		os.Exit(1)
	}
}

func resolveDates(cfg config) ([]time.Time, error) {
	// Default the newest date to yesterday: the daily cron owns today, and
	// iBreviary may not have rolled over to tomorrow yet in every timezone.
	end := time.Now().UTC().AddDate(0, 0, -1).Truncate(24 * time.Hour)
	if cfg.to != "" {
		parsed, err := time.Parse("2006-01-02", cfg.to)
		if err != nil {
			return nil, fmt.Errorf("--to: %w", err)
		}
		end = parsed
	}

	switch {
	case cfg.days > 0 && cfg.from != "":
		return nil, fmt.Errorf("pass either --days or --from, not both")
	case cfg.days > 0:
		return daysAgoRange(end, cfg.days), nil
	case cfg.from != "":
		start, err := time.Parse("2006-01-02", cfg.from)
		if err != nil {
			return nil, fmt.Errorf("--from: %w", err)
		}
		return dateRange(end, start), nil
	}
	return nil, nil
}

func processDate(cfg config, date time.Time) (string, error) {
	filename := util.OfficeOfReadingsFilename(date)

	if !cfg.force && !cfg.dryRun {
		exists, err := fileExists(cfg, filename)
		if err != nil {
			return "", fmt.Errorf("checking existing file: %w", err)
		}
		if exists {
			return "skipped", nil
		}
	}

	_, text, err := util.GetOfficeOfReadingsTextForDate(date)
	if err != nil {
		return "", err
	}

	if cfg.dryRun {
		preview := text
		if len(preview) > 90 {
			preview = preview[:90]
		}
		return fmt.Sprintf("dry-run OK, %d chars: %s...", len(text), sanitize(preview)), nil
	}

	if err := putFile(cfg, filename, text, date); err != nil {
		return "", err
	}
	return fmt.Sprintf("committed, %d chars", len(text)), nil
}

func sanitize(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if r == '\n' || r == '\r' || r == '\t' {
			r = ' '
		}
		out = append(out, r)
	}
	return string(out)
}

func githubRequest(cfg config, method, path string, body []byte) (int, []byte, error) {
	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/%s", repoOwner, repoName, path)

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, endpoint, reader)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Authorization", "Bearer "+cfg.githubToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer res.Body.Close()

	respBody, err := io.ReadAll(res.Body)
	return res.StatusCode, respBody, err
}

func fileExists(cfg config, filename string) (bool, error) {
	status, body, err := githubRequest(cfg, http.MethodGet, "contents/"+url.PathEscape(filename), nil)
	if err != nil {
		return false, err
	}
	switch status {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	default:
		return false, fmt.Errorf("HTTP %d: %s", status, string(body))
	}
}

// putFile commits straight to the default branch. The daily cron opens and merges
// a pull request per day, which is five API calls; for a backfill of hundreds of
// days that is a lot of churn for no review value, so this writes directly.
func putFile(cfg config, filename, content string, date time.Time) error {
	payload, err := json.Marshal(map[string]string{
		"message": fmt.Sprintf("add office of readings for %s", date.Format("2006-01-02")),
		"content": base64.StdEncoding.EncodeToString([]byte(content)),
	})
	if err != nil {
		return err
	}

	status, body, err := githubRequest(cfg, http.MethodPut, "contents/"+url.PathEscape(filename), payload)
	if err != nil {
		return err
	}
	if status != http.StatusOK && status != http.StatusCreated {
		return fmt.Errorf("HTTP %d: %s", status, string(body))
	}
	return nil
}
