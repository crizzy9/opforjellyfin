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
	dstPathNoSuffix := findMetadataMatch(fileName, index, ogcr)

	if dstPathNoSuffix == "" {
		logger.Log(true, "   ❌ No metadata match found for: %s", fileName)
		return "", fmt.Errorf("no metadata match found for file: %s (chapter range: %s)", fileName, ogcr)
	}

	logger.Log(false, "   📍 Target path (no ext): %s", dstPathNoSuffix)

	// extract suffix from original file
	ext := filepath.Ext(fileName)
	finalPath := dstPathNoSuffix + ext

	// SafeMoveFile now handles all locking internally
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

// returns directory to place file, without suffix
// Returns empty string if no match found
func findMetadataMatch(fileName string, index *shared.MetadataIndex, ogcr string) string {

	cfg := shared.LoadConfig()
	baseDir := cfg.TargetDir

	// finds season containing chapterRange, returns the seasonFolderName and seasonIndex
	// uses ogcr to find correct season even if its a bundle
	seasonFolderName, seasonIndex := findSeasonForChapter(ogcr, index)
	if seasonFolderName == "" {
		logger.Log(true, "   ❌ findMetaDataMatch: failed to find Season-folder for range %s", ogcr)
		return ""
	}
	logger.Log(false, "   ✓ Season found: %s for range %s", seasonFolderName, ogcr)

	// searches the seasonIndex for matching title for chapterRange, tries ogcr first for single-episode seasons
	newFileName := findTitleForChapter(ogcr, seasonIndex)
	if newFileName == "" {
		// if first fails, extract specific chapterRange from fileName
		chapterRange := shared.ExtractChapterRangeFromTitle(fileName)
		if chapterRange == "" {
			logger.Log(false, "   → Trying rough extraction for: %s", fileName)
			// use ogcr + file regex
			// if this extraction fails, try rougher methods
			seasonZ := shared.ExtractSeasonNumber(seasonFolderName)
			seasonNum := fmt.Sprintf("%02s", seasonZ)

			// rough extract can find chapterRange or rough chapter(in relation to season) if lucky.
			chapterNum, isRange := shared.RoughExtractChapterFromTitle(fileName)
			logger.Log(false, "   → Rough extracted chapterNum: %s", chapterNum)

			if isRange {
				newFileName = findTitleForChapter(chapterNum, seasonIndex)
			} else {
				// build a matching string from season and rough chapter, eg: seasonNum = 3 and chapternum = 05 => S03E05
				epKey := fmt.Sprintf("S%sE%s", seasonNum, chapterNum)
				newFileName = findTitleRough(epKey, seasonIndex)
			}
		} else {
			// if extraction succeeded, find title from chapterRange
			newFileName = findTitleForChapter(chapterRange, seasonIndex)
		}
	} else {
		logger.Log(false, "   ✓ Title match found: ChapterKey: %s - EpisodeTitle: %s", ogcr, newFileName)
	}

	if newFileName == "" {
		logger.Log(true, "   ❌ Could not determine episode title for file: %s", fileName)
		return ""
	}

	seasonDir := filepath.Join(baseDir, seasonFolderName)
	fullPathNoSuffix := filepath.Join(seasonDir, newFileName)

	logger.Log(false, "   → Target: %s", fullPathNoSuffix)
	return fullPathNoSuffix
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
