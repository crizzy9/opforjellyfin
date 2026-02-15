package matcher

import (
	"opforjellyfin/internal/shared"
	"testing"
)

func TestDetermineEpisodeTitle(t *testing.T) {
	// Mock Index
	index := &shared.MetadataIndex{
		Seasons: map[string]shared.SeasonIndex{
			"Season 1": {
				Range:        "1-3",
				Name:         "East Blue",
				SeasonNumber: 1,
				EpisodeRange: map[string]shared.EpisodeData{
					"1": {Title: "One Piece - 01 - I'm Luffy!"},
					"2": {Title: "One Piece - 02 - Zoro!"},
					"3": {Title: "One Piece - 03 - Nami!"},
				},
			},
		},
	}

	tests := []struct {
		name           string
		fileName       string
		ogcr           string
		expectedTitle  string
		expectedSeason string
	}{
		{
			name:           "Single Episode Torrent",
			fileName:       "[One Pace][1] Romace Dawn 01 [1080p].mkv",
			ogcr:           "1",
			expectedTitle:  "One Piece - 01 - I'm Luffy!",
			expectedSeason: "Season 1",
		},
		{
			name:           "Batch Torrent File 1",
			fileName:       "One Pace - 01 - Romance Dawn 01.mkv",
			ogcr:           "1-3",
			expectedTitle:  "One Piece - 01 - I'm Luffy!",
			expectedSeason: "Season 1",
		},
		{
			name:           "Batch Torrent File 2",
			fileName:       "One Pace - 02 - Romance Dawn 02.mkv",
			ogcr:           "1-3", // The torrent is for the whole arc
			expectedTitle:  "One Piece - 02 - Zoro!",
			expectedSeason: "Season 1",
		},
		{
			name:           "Batch Torrent File 3",
			fileName:       "[One Pace] Chapter 3 [1080p].mkv", // Simulating a filename that extract logic handles
			ogcr:           "1-3",
			expectedTitle:  "One Piece - 03 - Nami!",
			expectedSeason: "Season 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, season := DetermineEpisodeTitle(tt.fileName, index, tt.ogcr)
			if title != tt.expectedTitle {
				t.Errorf("expected title %q, got %q", tt.expectedTitle, title)
			}
			if season != tt.expectedSeason {
				t.Errorf("expected season %q, got %q", tt.expectedSeason, season)
			}
		})
	}
}
