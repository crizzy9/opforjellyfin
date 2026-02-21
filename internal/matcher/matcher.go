// matcher/matcher.go
package matcher

import (
	"fmt"
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/shared"
	"opforjellyfin/internal/ui"
	"os"
	"path/filepath"
	"strings"
)

// Matches video-file to metadata, then places it
// No mutex needed here - shared.SafeMoveFile handles all locking
func MatchAndPlaceVideo(videoPath, defaultDir string, index *shared.MetadataIndex, ogcr string) (string, error) {

	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		logger.Log(true, "   ❌ Video file does not exist: %s", videoPath)
		return "", nil
	}

	fileName := filepath.Base(videoPath)
	logger.Log(false, "   🔍 Attempting to match: %s (chapter range: %s)", fileName, ogcr)

	// strict
	dstPathNoSuffix := findMetadataMatch(fileName, index, ogcr, defaultDir)

	if dstPathNoSuffix == "" {
		logger.Log(true, "   ❌ No metadata match found for: %s", fileName)
		return "", fmt.Errorf("no metadata match found for file: %s (chapter range: %s)", fileName, ogcr)
	}

	logger.Log(false, "   📍 Target path (no ext): %s", dstPathNoSuffix)

	// extract suffix from original file
	ext := filepath.Ext(fileName)
	finalPath := dstPathNoSuffix + ext

	// SafeMoveFile now handles all locking internally
	logger.Log(true, "   🔄 Hardlinking: %s -> %s", videoPath, finalPath)
	if err := shared.SafeMoveFile(videoPath, finalPath); err != nil {
		logger.Log(true, "   ❌ Failed to place file to target location: %s", err)
		return "", fmt.Errorf("failed to place %s to %s: %w", fileName, finalPath, err)
	}

	//relative path for logs
	relPath, _ := filepath.Rel(defaultDir, finalPath)
	//debug
	logger.Log(false, "%s", fmt.Sprintf("placed: %s → %s", fileName, relPath))

	// some formatting
	fileNameNoPrefix := fileName
	if len(fileName) > 10 {
		fileNameNoPrefix = fileName[10:]
	}
	relPathNoPrefix := filepath.Base(relPath)
	if len(relPathNoPrefix) > 10 {
		relPathNoPrefix = relPathNoPrefix[10:]
	}
	outFileName := ui.AnsiPadRight(fileNameNoPrefix, 26, "..")
	outRelPath := ui.AnsiPadRight(".."+relPathNoPrefix, 36, "..")
	msg := fmt.Sprintf("🎞️  Placed: %s → %s", outFileName, outRelPath)

	return msg, nil
}

// findMetadataMatch determines the best match for a given filename
func findMetadataMatch(fileName string, index *shared.MetadataIndex, ogcr string, baseDir string) string {
	// findMetadataMatch now delegates to DetermineEpisodeTitle
	newFileName, seasonFolderName := DetermineEpisodeTitle(fileName, index, ogcr)

	if newFileName == "" || seasonFolderName == "" {
		logger.Log(true, "   ❌ Could not determine episode title or season for file: %s", fileName)
		return ""
	}

	seasonDir := filepath.Join(baseDir, seasonFolderName)
	fullPathNoSuffix := filepath.Join(seasonDir, newFileName)

	logger.Log(false, "   → Target: %s", fullPathNoSuffix)
	return fullPathNoSuffix
}

// DetermineEpisodeTitle is a pure function that determines the target episode title and season folder
// based on the filename, metadata index, and original global chapter range (ogcr).
func DetermineEpisodeTitle(fileName string, index *shared.MetadataIndex, ogcr string) (string, string) {
	// 1. Find the season containing the OGCR
	seasonFolderName, seasonIndex := findSeasonForChapter(ogcr, index)
	if seasonFolderName == "" {
		logger.Log(true, "   ❌ DetermineEpisodeTitle: failed to find Season-folder for range %s", ogcr)
		return "", ""
	}
	logger.Log(false, "   ✓ Season found: %s for range %s", seasonFolderName, ogcr)

	// 2. Try to extract specific chapter info from the FILENAME first
	// This fixes the bug where batch torrents all get mapped to the same OGCR episode

	// Try extracting exact chapter/range from filename
	fileChapterRange := shared.ExtractChapterRangeFromTitle(fileName)
	var newFileName string

	if fileChapterRange != "" {
		// Check if this extracted range exists in the current season
		newFileName = findTitleForChapter(fileChapterRange, seasonIndex)
		if newFileName != "" {
			logger.Log(false, "   ✓ Filename match found: %s -> %s", fileName, newFileName)
			return newFileName, seasonFolderName
		}
	}

	// 3. If filename extraction failed or didn't match an episode in the season,
	// try rough extraction from filename

	// Rough extraction logic
	_, isRange := shared.RoughExtractChapterFromTitle(fileName)
	if !isRange {
		// Try to construct a key like S01E01 if we found a rough number
		seasonZ := shared.ExtractSeasonNumber(seasonFolderName)
		seasonNum := fmt.Sprintf("%02s", seasonZ)
		chapterNum, _ := shared.RoughExtractChapterFromTitle(fileName)

		if chapterNum != "" {
			epKey := fmt.Sprintf("S%sE%s", seasonNum, chapterNum)
			newFileName = findTitleRough(epKey, seasonIndex)
			if newFileName != "" {
				logger.Log(false, "   ✓ Rough filename match found: %s -> %s", fileName, newFileName)
				return newFileName, seasonFolderName
			}
		}
	}

	// 4. Fallback: Use the OGCR (Global Torrent Range)
	// This should only happen for single-episode torrents where the filename might not be descriptive enough,
	// or if we truly failed to parse the filename.
	newFileName = findTitleForChapter(ogcr, seasonIndex)
	if newFileName != "" {
		logger.Log(false, "   ✓ OGCR match found: %s -> %s", ogcr, newFileName)
		return newFileName, seasonFolderName
	}

	return "", ""
}

// exact match, returns title from metadataindex using chapterKey.
func findTitleForChapter(chapterKey string, sindex shared.SeasonIndex) string {
	normKey := shared.NormalizeDash(chapterKey)

	logger.Log(false, "findEpisodeKeyForChapter: chapterKey: %s - normKey: %s ", chapterKey, normKey)

	for epRange, ep := range sindex.EpisodeRange {
		if shared.NormalizeDash(epRange) == normKey {
			return ep.Title
		}
	}

	// no title found based on ChapterKey,
	return ""
}

// finds the season a ChapterKey belongs to. returns the season name as a string, also returns the whole SeasonIndex struct
func findSeasonForChapter(chapterKey string, index *shared.MetadataIndex) (string, shared.SeasonIndex) {
	chStart, chEnd := shared.ParseRange(chapterKey)

	for seasonName, season := range index.Seasons {
		seasonStart, seasonEnd := shared.ParseRange(season.Range)

		if chStart >= seasonStart && chEnd <= seasonEnd {
			return seasonName, season
		}
	}

	return "", shared.SeasonIndex{}

}

// rough finder
func findTitleRough(epKey string, sindex shared.SeasonIndex) string {

	for _, ep := range sindex.EpisodeRange {
		if strings.Contains(ep.Title, epKey) {
			logger.Log(false, "roughFindTitle match found: %s > %s", epKey, ep.Title)
			return ep.Title
		}
	}

	logger.Log(false, "roughFindTitle did not find a match. for %s", epKey)
	return ""
}
