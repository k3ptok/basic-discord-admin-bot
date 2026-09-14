package automod

import (
	"bufio"
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type DomainStore struct {
	mu      sync.RWMutex
	domains map[string]struct{}
	logger  *slog.Logger
}

func NewDomainStore(logger *slog.Logger) *DomainStore {
	return &DomainStore{
		domains: 	make(map[string]struct{}),
		logger:		logger,
	}
}

//check if domain already exists in blocklist
func (ds *DomainStore) Has(domain string) bool {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	_, found := ds.domains[domain]
	return found
}

//returns len of loaded scam domains
func (ds *DomainStore) Count() int {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return len(ds.domains)
}

func (ds *DomainStore) LoadFromReader(r io.Reader) error {
	newMap := make(map[string]struct{})
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		fields := strings.Fields(line)
		domain := fields[len(fields)-1]
		newMap[strings.ToLower(domain)] = struct{}{}

	}
	if err := scanner.Err(); err != nil {
		return err
	}
	//swap pointers
	ds.mu.Lock()
	ds.domains = newMap
	ds.mu.Unlock()

	ds.logger.Info("Updated scam domains in memory", slog.Int("total_domains", len(newMap)))
	return nil
}

//sync from remote feed in the background every in the background
func (ds *DomainStore) StartAutoUpdater(localPath, feedURL string, interval time.Duration) {
	// boot load
	if err := ds.LoadFromFile(localPath); err != nil {
		ds.logger.Warn("Could not load local scam domains backup", slog.Any("error", err))
	} else {
		ds.logger.Info("Loaded initial scam domains from local file", slog.Int("count", ds.Count()))
	}

	//Initial fetch
	ds.fetchAndSave(feedURL, localPath)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			ds.fetchAndSave(feedURL, localPath)
		}
	}()
}

func (ds *DomainStore) LoadFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	ds.LoadFromReader(file)
	return nil
}

func (ds *DomainStore) fetchAndSave(feedURL, localPath string) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(feedURL)
	if err != nil {
		ds.logger.Error("Failed to fetch threats from feed", slog.Any("error", err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ds.logger.Warn("Threat feed returned non-200 status", slog.Int("status", resp.StatusCode))
		return
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		ds.logger.Error("Failed to read remote feed body", slog.Any("error", err))
		return
	}

	if err := ds.LoadFromReader(bytes.NewReader(bodyBytes)); err != nil {
		ds.logger.Error("Failed parsing remote stream", slog.Any("error", err))
		return
	}

	if err := os.WriteFile(localPath, bodyBytes, 0644); err != nil {
		ds.logger.Warn("Failed to update local cache file on disk", slog.Any("error", err))
	}
}