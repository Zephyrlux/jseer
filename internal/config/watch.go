package config

import "time"

// Watch reloads config on a fixed interval and invokes onChange.
// interval<=0 disables watching.
func Watch(path string, interval time.Duration, onChange func(*Config)) (stop func()) {
	if interval <= 0 {
		return func() {}
	}
	done := make(chan struct{})
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if cfg, err := Load(path); err == nil {
					if onChange != nil {
						onChange(cfg)
					}
				}
			case <-done:
				return
			}
		}
	}()
	return func() { close(done) }
}
