package game

import (
	"encoding/xml"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type itemPriceDB struct {
	mu     sync.RWMutex
	loaded bool
	prices map[int]int
}

var globalItemPrices = &itemPriceDB{}

type itemNameDB struct {
	mu     sync.RWMutex
	loaded bool
	names  map[string]int
}

var globalItemNames = &itemNameDB{}

var uniqueRanges = []struct {
	Min int
	Max int
}{
	{Min: 100001, Max: 199999},
	{Min: 1000, Max: 1300},
}

var uniqueIDs = []int{}

func LoadItemPrices() map[int]int {
	globalItemPrices.mu.Lock()
	defer globalItemPrices.mu.Unlock()
	if globalItemPrices.loaded {
		return globalItemPrices.prices
	}
	globalItemPrices.prices = make(map[int]int)
	dataRoot := resolveDataRoot()
	_ = loadItemPrices(filepath.Join(dataRoot, "items.xml"), globalItemPrices.prices)
	globalItemPrices.loaded = true
	return globalItemPrices.prices
}

func LoadItemNames() map[string]int {
	globalItemNames.mu.Lock()
	defer globalItemNames.mu.Unlock()
	if globalItemNames.loaded {
		return globalItemNames.names
	}
	globalItemNames.names = make(map[string]int)
	dataRoot := resolveDataRoot()
	_ = loadItemNames(filepath.Join(dataRoot, "items.xml"), globalItemNames.names)
	globalItemNames.loaded = true
	return globalItemNames.names
}

func loadItemPrices(path string, out map[int]int) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := xml.NewDecoder(f)
	for {
		tok, err := dec.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if se.Name.Local != "Item" {
			continue
		}
		id := 0
		price := 0
		for _, attr := range se.Attr {
			switch attr.Name.Local {
			case "ID":
				id, _ = strconv.Atoi(attr.Value)
			case "Price":
				price, _ = strconv.Atoi(attr.Value)
			}
		}
		if id > 0 {
			out[id] = price
		}
	}
	return nil
}

func loadItemNames(path string, out map[string]int) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := xml.NewDecoder(f)
	for {
		tok, err := dec.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if se.Name.Local != "Item" {
			continue
		}
		id := 0
		name := ""
		for _, attr := range se.Attr {
			switch attr.Name.Local {
			case "ID":
				id, _ = strconv.Atoi(attr.Value)
			case "Name":
				name = attr.Value
			}
		}
		if id > 0 && name != "" {
			out[name] = id
		}
	}
	return nil
}

func getItemPrice(itemID int) int {
	prices := LoadItemPrices()
	if prices == nil {
		return 0
	}
	return prices[itemID]
}

func isUniqueItem(itemID int) bool {
	for _, id := range uniqueIDs {
		if itemID == id {
			return true
		}
	}
	for _, r := range uniqueRanges {
		if itemID >= r.Min && itemID <= r.Max {
			return true
		}
	}
	return false
}

func findItemIDByName(name string) int {
	if name == "" {
		return 0
	}
	names := LoadItemNames()
	if id := names[name]; id > 0 {
		return id
	}
	if strings.Contains(name, "精元") && !strings.Contains(name, "的") {
		alt := strings.Replace(name, "精元", "的精元", 1)
		if id := names[alt]; id > 0 {
			return id
		}
	}
	return 0
}
