package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"tap-server/src/model"
)

const maxStoredRecords = 5000

type JSONStore struct {
	mu       sync.RWMutex
	flushMu  sync.Mutex
	dataDir  string
	players  map[string]*model.Player
	chats    []model.InteractionRecord
	events   []model.InteractionRecord
	version  uint64
	dirty    bool
	stop     chan struct{}
	done     chan struct{}
	closeOne sync.Once
}

func NewJSONStore(dataDir string) (*JSONStore, error) {
	store := &JSONStore{
		dataDir: dataDir,
		players: make(map[string]*model.Player),
		chats:   make([]model.InteractionRecord, 0),
		events:  make([]model.InteractionRecord, 0),
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}
	if err := store.load(); err != nil {
		return nil, err
	}
	go store.periodicFlush()
	return store, nil
}

func (s *JSONStore) load() error {
	interactionsPath := filepath.Join(s.dataDir, "interactions.json")
	rawInteractions, err := os.ReadFile(interactionsPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read interactions file: %w", err)
	}

	if len(rawInteractions) > 0 && bytes.HasPrefix(bytes.TrimSpace(rawInteractions), []byte("{")) {
		var legacy model.InteractionDatabase
		if err := json.Unmarshal(rawInteractions, &legacy); err != nil {
			return fmt.Errorf("parse legacy interaction database: %w", err)
		}
		if legacy.Players != nil {
			for username, player := range legacy.Players {
				if player != nil {
					s.players[username] = clonePlayer(player)
				}
			}
		}
		for _, record := range legacy.Interactions {
			s.appendRecord(record)
		}
		s.version++
		s.dirty = true
		return nil
	}

	if len(rawInteractions) > 0 {
		if err := json.Unmarshal(rawInteractions, &s.events); err != nil {
			return fmt.Errorf("parse interactions file: %w", err)
		}
		s.events = recordsWithoutChat(s.events)
	}

	if err := readJSONFile(filepath.Join(s.dataDir, "chat.json"), &s.chats); err != nil {
		return fmt.Errorf("load chat file: %w", err)
	}
	s.chats = recordsOfType(s.chats, model.InteractionChat)

	if err := readJSONFile(filepath.Join(s.dataDir, "players.json"), &s.players); err != nil {
		return fmt.Errorf("load players file: %w", err)
	}
	if s.players == nil {
		s.players = make(map[string]*model.Player)
	}
	for username, player := range s.players {
		if player != nil {
			s.players[username] = clonePlayer(player)
		}
	}

	return nil
}

func readJSONFile(path string, target any) error {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return nil
}

func (s *JSONStore) RecordInteraction(level string, interactionType model.InteractionType, actor, scope, target, payload string) {
	s.RecordEvent(model.InteractionRecord{
		Level:   level,
		Type:    interactionType,
		Actor:   actor,
		Scope:   scope,
		Target:  target,
		Payload: payload,
	})
}

func (s *JSONStore) RecordEvent(record model.InteractionRecord) {
	record.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	record.Timestamp = time.Now().UTC()
	if record.Level == "" {
		record.Level = "INFO"
	}
	if record.Type == "" {
		record.Type = model.InteractionSystem
	}
	PrintPrettyLog(record)

	s.mu.Lock()
	s.appendRecord(record)
	s.version++
	s.dirty = true
	s.mu.Unlock()
}

func (s *JSONStore) appendRecord(record model.InteractionRecord) {
	if record.Type == model.InteractionChat {
		s.chats = append(s.chats, record)
		if len(s.chats) > maxStoredRecords {
			s.chats = append([]model.InteractionRecord(nil), s.chats[len(s.chats)-maxStoredRecords:]...)
		}
		return
	}
	s.events = append(s.events, record)
	if len(s.events) > maxStoredRecords {
		s.events = append([]model.InteractionRecord(nil), s.events[len(s.events)-maxStoredRecords:]...)
	}
}

func (s *JSONStore) SavePlayerSnapshot(player *model.Player) {
	if player == nil {
		return
	}
	s.mu.Lock()
	s.players[player.Username] = clonePlayer(player)
	s.version++
	s.dirty = true
	s.mu.Unlock()
}

func (s *JSONStore) GetSavedPlayer(username string) (*model.Player, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	player, ok := s.players[username]
	if !ok || player == nil {
		return nil, false
	}
	return clonePlayer(player), true
}

func (s *JSONStore) periodicFlush() {
	defer close(s.done)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := s.Flush(); err != nil {
				s.writeErrorLog(err)
			}
		case <-s.stop:
			return
		}
	}
}

func (s *JSONStore) Flush() error {
	s.flushMu.Lock()
	defer s.flushMu.Unlock()

	s.mu.RLock()
	if !s.dirty {
		s.mu.RUnlock()
		return nil
	}
	players := make(map[string]*model.Player, len(s.players))
	for username, player := range s.players {
		players[username] = clonePlayer(player)
	}
	chats := append([]model.InteractionRecord(nil), s.chats...)
	events := append([]model.InteractionRecord(nil), s.events...)
	version := s.version
	s.mu.RUnlock()

	files := []struct {
		name  string
		value any
	}{
		{name: "players.json", value: players},
		{name: "chat.json", value: chats},
		{name: "interactions.json", value: events},
	}
	for _, file := range files {
		if err := s.writeJSON(file.name, file.value); err != nil {
			return err
		}
	}

	s.mu.Lock()
	if s.version == version {
		s.dirty = false
	}
	s.mu.Unlock()
	return nil
}

func (s *JSONStore) writeJSON(name string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", name, err)
	}
	path := filepath.Join(s.dataDir, name)
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0600); err != nil {
		return fmt.Errorf("write %s: %w", name, err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("replace %s: %w", name, err)
	}
	return nil
}

func (s *JSONStore) Close() error {
	s.closeOne.Do(func() {
		close(s.stop)
	})
	<-s.done
	return s.Flush()
}

func (s *JSONStore) writeErrorLog(err error) {
	record := model.InteractionRecord{
		Timestamp: time.Now().UTC(),
		Level:     "ERROR",
		Type:      model.InteractionSystem,
		Event:     "store_flush_failed",
		Payload:   err.Error(),
	}
	PrintPrettyLog(record)
}

func recordsWithoutChat(records []model.InteractionRecord) []model.InteractionRecord {
	filtered := make([]model.InteractionRecord, 0, len(records))
	for _, record := range records {
		if record.Type != model.InteractionChat {
			filtered = append(filtered, record)
		}
	}
	return filtered
}

func recordsOfType(records []model.InteractionRecord, interactionType model.InteractionType) []model.InteractionRecord {
	filtered := make([]model.InteractionRecord, 0, len(records))
	for _, record := range records {
		if record.Type == interactionType {
			filtered = append(filtered, record)
		}
	}
	return filtered
}

func clonePlayer(player *model.Player) *model.Player {
	copyPlayer := *player
	copyPlayer.Inventory = append([]string(nil), player.Inventory...)
	copyPlayer.Quests = make(map[string]*model.PlayerQuest, len(player.Quests))
	for key, quest := range player.Quests {
		if quest != nil {
			copyQuest := *quest
			copyPlayer.Quests[key] = &copyQuest
		}
	}
	return &copyPlayer
}
