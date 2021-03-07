package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

type Config struct {
	SacuraLogPath   string `required:"true" split_words:"true"`
	ComponentsPaths string `required:"true" split_words:"true"`
	OutPath         string `required:"true" split_words:"true"`
}

func run() error {
	cfg := &Config{}
	if err := envconfig.Process("", cfg); err != nil {
		return fmt.Errorf("failed to process env var: %w", err)
	}

	paths := strings.Split(cfg.ComponentsPaths, ",")
	finders := make([]HistoryFinder, 0, len(paths))
	for _, p := range paths {
		finder, err := NewSingleFileHistoryFinder(p)
		if err != nil {
			return err
		}
		finders = append(finders, finder)
	}

	parser := SacuraLogParser{HistoryFinder: MultiFileHistoryFinder(finders)}
	history, err := parser.Parse(cfg.SacuraLogPath)
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(history, "", " ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	if err := ioutil.WriteFile(cfg.OutPath, b, os.ModePerm); err != nil {
		return fmt.Errorf("failed to write file %s: %w", cfg.OutPath, err)
	}

	return nil
}

type EventHistory struct {
	ID      string   `json:"id"`
	History []string `json:"history"`
}

type HistoryBySymbol map[string][]EventHistory

type History struct {
	HistoryBySymbol
}

type SacuraLogParser struct {
	HistoryFinder
}

func (p *SacuraLogParser) Parse(path string) (History, error) {
	f, err := os.Open(path)
	if err != nil {
		return History{}, fmt.Errorf("failed to open file: %w", err)
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return History{}, fmt.Errorf("failed to read file: %w", err)
	}

	return p.parse(string(b))
}

func (p *SacuraLogParser) parse(s string) (History, error) {
	bySymbol := make(HistoryBySymbol, 2)
	lines := strings.Split(s, "\n")
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if len(l) == 0 {
			continue
		}
		if l[0] != '+' && l[0] != '-' {
			continue
		}
		parts := strings.Split(l, "\"")
		id := parts[1]

		symbol := string([]rune(l)[0])

		history, err := p.HistoryFinder.Find(id)
		if err != nil {
			return History{}, err
		}

		bySymbol[symbol] = append(bySymbol[symbol], history)
	}

	return History{HistoryBySymbol: bySymbol}, nil
}

type HistoryFinder interface {
	Find(id string) (EventHistory, error)
}

type SingleFileHistoryFinder struct {
	lines []string
}

func NewSingleFileHistoryFinder(path string) (*SingleFileHistoryFinder, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	lines := strings.Split(string(b), "\n")

	// A potential log line with an id contains at least 4 '-'.
	out := make([]string, len(lines))
	for _, l := range lines {
		count := 0
		for c := range []rune(l) {
			if c == '-' {
				count++
			}
			if count >= 4 {
				out = append(out, l)
			}
		}
	}
	return &SingleFileHistoryFinder{lines: out}, nil
}

func (sf *SingleFileHistoryFinder) Find(id string) (EventHistory, error) {
	history := make([]string, 5)
	for _, l := range sf.lines {
		if strings.Contains(l, id) {
			history = append(history, l)
		}
	}
	return EventHistory{ID: id, History: history}, nil
}

type MultiFileHistoryFinder []HistoryFinder

func (mf MultiFileHistoryFinder) Find(id string) (EventHistory, error) {
	history := EventHistory{
		ID:      id,
		History: make([]string, 5*len(mf)),
	}
	for _, f := range mf {
		h, err := f.Find(id)
		if err != nil {
			return EventHistory{}, err
		}
		history.History = append(history.History, h.History...)
	}
	return history, nil
}
