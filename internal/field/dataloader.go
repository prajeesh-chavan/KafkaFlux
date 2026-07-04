package field

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
)

const dataDirKey = "__data"

type DataLoader struct {
	mu          sync.RWMutex
	baseDir     string
	FirstNames  []string
	LastNames   []string
	StreetNames []string
	Cities      map[string][]string
	States      map[string][]string
	Companies   []string
	JobTitles   []string
	UserAgents  []string
	LoremWords  []string
	CountryCodes map[string]string
	currencies  []string
	languages   []string
	countryCodesList []string
}

var globalLoader *DataLoader
var loadOnce sync.Once

func InitDataLoader(dataDir string) (*DataLoader, error) {
	var initErr error
	loadOnce.Do(func() {
		globalLoader, initErr = newDataLoader(dataDir)
	})
	if initErr != nil {
		return nil, initErr
	}
	return globalLoader, nil
}

func GetDataLoader() *DataLoader {
	return globalLoader
}

func newDataLoader(dataDir string) (*DataLoader, error) {
	dl := &DataLoader{
		baseDir:     dataDir,
		Cities:      make(map[string][]string),
		States:      make(map[string][]string),
		CountryCodes: make(map[string]string),
	}

	if err := dl.loadJSON("first_names.json", &dl.FirstNames); err != nil {
		return nil, err
	}
	if err := dl.loadJSON("last_names.json", &dl.LastNames); err != nil {
		return nil, err
	}
	if err := dl.loadJSON("street_names.json", &dl.StreetNames); err != nil {
		return nil, err
	}
	if err := dl.loadJSON("cities.json", &dl.Cities); err != nil {
		return nil, err
	}
	if err := dl.loadJSON("states.json", &dl.States); err != nil {
		return nil, err
	}
	if err := dl.loadJSON("companies.json", &dl.Companies); err != nil {
		return nil, err
	}
	if err := dl.loadJSON("job_titles.json", &dl.JobTitles); err != nil {
		return nil, err
	}
	if err := dl.loadJSON("user_agents.json", &dl.UserAgents); err != nil {
		return nil, err
	}
	if err := dl.loadJSON("lorem_words.json", &dl.LoremWords); err != nil {
		return nil, err
	}
	if err := dl.loadJSON("country_codes.json", &dl.CountryCodes); err != nil {
		return nil, err
	}

	dl.currencies = []string{"USD", "EUR", "GBP", "INR", "JPY", "CAD", "AUD", "CNY", "BRL", "CHF"}
	dl.languages = []string{"en", "es", "fr", "de", "zh", "ja", "pt", "ru", "ar", "hi"}
	dl.countryCodesList = []string{"US", "IN", "GB", "DE", "FR", "JP", "CN", "BR", "CA", "AU"}

	return dl, nil
}

func (dl *DataLoader) loadJSON(filename string, target interface{}) error {
	path := filepath.Join(dl.baseDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to load %s: %w", filename, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to parse %s: %w", filename, err)
	}
	return nil
}

func getLoader(state map[string]interface{}) *DataLoader {
	if state == nil {
		if globalLoader != nil {
			return globalLoader
		}
		return emptyLoader
	}
	if v, ok := state[dataDirKey]; ok {
		if dl, ok2 := v.(*DataLoader); ok2 {
			return dl
		}
	}
	if globalLoader != nil {
		return globalLoader
	}
	return emptyLoader
}

var emptyLoader = &DataLoader{
	FirstNames:   []string{"John", "Jane"},
	LastNames:    []string{"Doe", "Smith"},
	StreetNames:  []string{"Main St"},
	Cities:       map[string][]string{"US": {"New York"}},
	States:       map[string][]string{"US": {"NY"}},
	Companies:    []string{"Acme"},
	JobTitles:    []string{"Engineer"},
	UserAgents:   []string{"Mozilla/5.0"},
	LoremWords:   []string{"lorem", "ipsum"},
	CountryCodes: map[string]string{"US": "United States", "IN": "India"},
	currencies:   []string{"USD"},
	languages:    []string{"en"},
	countryCodesList: []string{"US"},
}

func (dl *DataLoader) RandomString(r *rand.Rand, items []string) string {
	if len(items) == 0 {
		return ""
	}
	return items[r.Intn(len(items))]
}

func (dl *DataLoader) RandomCity(r *rand.Rand, countryCode string) string {
	dl.mu.RLock()
	defer dl.mu.RUnlock()
	list, ok := dl.Cities[countryCode]
	if !ok || len(list) == 0 {
		list = dl.Cities["US"]
	}
	if len(list) == 0 {
		return ""
	}
	return list[r.Intn(len(list))]
}

func (dl *DataLoader) RandomState(r *rand.Rand, countryCode string) string {
	dl.mu.RLock()
	defer dl.mu.RUnlock()
	list, ok := dl.States[countryCode]
	if !ok || len(list) == 0 {
		list = dl.States["US"]
	}
	if len(list) == 0 {
		return ""
	}
	return list[r.Intn(len(list))]
}

func (dl *DataLoader) CountryName(code string) string {
	dl.mu.RLock()
	defer dl.mu.RUnlock()
	if name, ok := dl.CountryCodes[code]; ok {
		return name
	}
	return "United States"
}

func (dl *DataLoader) Currencies() []string { return dl.currencies }
func (dl *DataLoader) Languages() []string  { return dl.languages }
func (dl *DataLoader) CountryCodesList() []string { return dl.countryCodesList }
