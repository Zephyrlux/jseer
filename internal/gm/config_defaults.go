package gm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"

	"jseer/internal/storage"
)

type configSeed struct {
	Key   string
	Value map[string]any
}

func defaultConfigSeeds() []configSeed {
	return []configSeed{
		{
			Key: "role_attributes",
			Value: map[string]any{
				"level_cap":    100,
				"exp_rate":     1,
				"base_hp":      100,
				"base_attack":  20,
				"base_defence": 15,
				"base_speed":   10,
				"energy_cap":   100,
			},
		},
		{
			Key: "items_equipment",
			Value: map[string]any{
				"items":      []any{},
				"equipments": []any{},
			},
		},
		{
			Key: "dungeons",
			Value: map[string]any{
				"dungeons": []any{},
			},
		},
		{
			Key: "shop",
			Value: map[string]any{
				"products": []any{},
			},
		},
		{
			Key: "events",
			Value: map[string]any{
				"events": []any{},
			},
		},
		{
			Key: "battle",
			Value: map[string]any{
				"crit_rate":       0.1,
				"crit_multiplier": 1.5,
				"type_adv":        1.2,
				"type_resist":     0.8,
				"status_hit":      0.35,
			},
		},
		{
			Key: "economy",
			Value: map[string]any{
				"coin_supply_rate":  1,
				"gold_supply_rate":  1,
				"tax_rate":          0.02,
				"recycle_rate":      0.6,
				"daily_reward_coin": 1000,
				"daily_reward_gold": 10,
			},
		},
	}
}

func loadConfigSeeds() []configSeed {
	dir := filepath.Join(".", "data", "config")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return defaultConfigSeeds()
	}
	seeds := make([]configSeed, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".json") {
			continue
		}
		raw, readErr := os.ReadFile(filepath.Join(dir, name))
		if readErr != nil {
			continue
		}
		var val map[string]any
		if err := json.Unmarshal(raw, &val); err != nil {
			continue
		}
		key := strings.TrimSuffix(name, filepath.Ext(name))
		key = strings.ReplaceAll(key, "-", "_")
		seeds = append(seeds, configSeed{Key: key, Value: val})
	}
	if len(seeds) == 0 {
		return defaultConfigSeeds()
	}
	return seeds
}

func seedDefaultConfigs(ctx context.Context, store storage.Store, logger *zap.Logger) {
	keys, err := store.ListConfigKeys(ctx)
	if err != nil {
		logger.Warn("gm config seed list failed", zap.Error(err))
		return
	}
	existing := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		existing[k] = struct{}{}
	}

	for _, seed := range loadConfigSeeds() {
		if _, ok := existing[seed.Key]; ok {
			continue
		}
		raw, err := json.Marshal(seed.Value)
		if err != nil {
			logger.Warn("gm config seed marshal failed", zap.String("key", seed.Key), zap.Error(err))
			continue
		}
		sum := sha256.Sum256(raw)
		_, err = store.SaveConfig(ctx, &storage.ConfigEntry{
			Key:      seed.Key,
			Value:    raw,
			Checksum: hex.EncodeToString(sum[:]),
		}, "bootstrap")
		if err != nil {
			logger.Warn("gm config seed save failed", zap.String("key", seed.Key), zap.Error(err))
		}
	}
}
