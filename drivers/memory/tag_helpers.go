package memory

// removeKeyTags removes tag associations for a key.
// Caller must hold the lock.
func (d *Driver) removeKeyTags(key string) {
	if tags, ok := d.keyTags[key]; ok {
		for _, tag := range tags {
			if keys, ok := d.tags[tag]; ok {
				delete(keys, key)
				if len(keys) == 0 {
					delete(d.tags, tag)
				}
			}
		}
		delete(d.keyTags, key)
	}
}

// addKeyTags adds tag associations for a key.
// Caller must hold the lock.
func (d *Driver) addKeyTags(key string, tags []string) {
	if len(tags) == 0 {
		return
	}

	// Remove old tags if any (simplifies logic for overwrites)
	d.removeKeyTags(key)

	d.keyTags[key] = tags
	for _, tag := range tags {
		if _, ok := d.tags[tag]; !ok {
			d.tags[tag] = make(map[string]struct{})
		}
		d.tags[tag][key] = struct{}{}
	}
}
