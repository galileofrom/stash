package config

// Galileo fork: typed accessors for the media request workflow keys.
// Stash uses package-private getters elsewhere; these expose just the
// fork-owned keys to consumers in internal/api without widening the
// general getter surface.

func (i *Config) GetProwlarrEnabled() bool { return i.getBool(ProwlarrEnabled) }
func (i *Config) GetProwlarrURL() string   { return i.getString(ProwlarrURL) }
func (i *Config) GetProwlarrAPIKey() string { return i.getString(ProwlarrAPIKey) }

func (i *Config) GetProwlarrCategories() []int {
	return intSliceFromAny(i.main.Get(ProwlarrCategories))
}

func (i *Config) GetProwlarrIndexerIDs() []int {
	return intSliceFromAny(i.main.Get(ProwlarrIndexerIDs))
}

func (i *Config) GetProwlarrTimeoutSeconds() int { return i.getInt(ProwlarrTimeoutSeconds) }

func (i *Config) GetRequestsLibraryPath() string  { return i.getString(RequestsLibraryPath) }
func (i *Config) GetRequestsRequireApproval() bool { return i.getBool(RequestsRequireApproval) }
func (i *Config) GetRequestsPreferProtocol() string { return i.getString(RequestsPreferProtocol) }
func (i *Config) GetRequestsMinSeeders() int { return i.getInt(RequestsMinSeeders) }
func (i *Config) GetRequestsMaxSizeBytes() int64 { return int64(i.getInt(RequestsMaxSizeBytes)) }

func (i *Config) GetRequestsRankUsenetBoost() float64    { return i.getFloat64(RequestsRankUsenetBoost) }
func (i *Config) GetRequestsRankTorrentBoost() float64   { return i.getFloat64(RequestsRankTorrentBoost) }
func (i *Config) GetRequestsRankSeederWeight() float64   { return i.getFloat64(RequestsRankSeederWeight) }
func (i *Config) GetRequestsRankFreeleechBoost() float64 { return i.getFloat64(RequestsRankFreeleechBoost) }
func (i *Config) GetRequestsRankAgeDecayDays() float64   { return i.getFloat64(RequestsRankAgeDecayDays) }

func (i *Config) GetRequestsQBittorrentURL() string  { return i.getString(RequestsQBittorrentURL) }
func (i *Config) GetRequestsQBittorrentUser() string { return i.getString(RequestsQBittorrentUser) }
func (i *Config) GetRequestsQBittorrentPass() string { return i.getString(RequestsQBittorrentPass) }

// intSliceFromAny normalises common viper representations of an int list:
// []int, []interface{} of numbers, or a comma-separated string.
func intSliceFromAny(v interface{}) []int {
	switch x := v.(type) {
	case nil:
		return nil
	case []int:
		return x
	case []interface{}:
		out := make([]int, 0, len(x))
		for _, e := range x {
			switch n := e.(type) {
			case int:
				out = append(out, n)
			case int64:
				out = append(out, int(n))
			case float64:
				out = append(out, int(n))
			}
		}
		return out
	}
	return nil
}
