package matcher

import (
	"fmt"
	"opforjellyfin/internal/logger"
	"opforjellyfin/internal/shared"
	"os"
	"path/filepath"
	"strings"
)

// walks through downloaded files and tries to place them in correct dir
func ProcessTorrentFiles(tmpDir, outDir string, td *shared.TorrentDownload, index *shared.MetadataIndex, files []string) {
	filesChecked := 0
	filesPlaced := 0
	var lastError error

	// collect all paths
	td.PlacementProgress = fmt.Sprintf("🔧 Finding files to place in %s", tmpDir)

	var vidPaths []string

	if len(files) > 0 {
		logger.Log(true, "📋 Using provided file list for import (%d files)", len(files))
		for _, f := range files {
			// Sanitize file path to prevent directory traversal
			cleanPath := filepath.Clean(f)
			if strings.Contains(cleanPath, "..") {
				continue
			}

			fullPath := filepath.Join(tmpDir, cleanPath)
			info, err := os.Stat(fullPath)
			if err != nil {
				// File might be in a subdirectory that we need to find, but typically torrent clients return relative paths
				// If strictly not found, skip
				continue
			}
			if info.IsDir() {
				continue
			}

			ext := strings.ToLower(filepath.Ext(info.Name()))
			if ext == ".mkv" || ext == ".mp4" {
				logger.Log(true, "   ✅ Found video file from list: %s", info.Name())
				vidPaths = append(vidPaths, fullPath)
			}
		}
	} else {
		logger.Log(true, "⚠️ No file list provided, scanning directory: %s", tmpDir)
		err := filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				logger.Log(true, "❌ Failed walking file: %v", err)
				return nil
			}
			if info.IsDir() {
				logger.Log(false, "   📁 Directory: %s", info.Name())
				return nil
			}

			ext := strings.ToLower(filepath.Ext(info.Name()))
			if ext != ".mkv" && ext != ".mp4" {
				logger.Log(false, "   ⏭️  Skipping non-video file: %s", info.Name())
				return nil
			}

			logger.Log(true, "   ✅ Found video file: %s (%.2f MB)", info.Name(), float64(info.Size())/(1024*1024))
			vidPaths = append(vidPaths, path)
			return nil
		})

		if err != nil {
			logger.Log(true, "❌ Error walking tmpDir %s: %v", tmpDir, err)
			td.MarkPlaced(fmt.Sprintf("❌ Error scanning directory: %v", err))
			return
		}
	}

	// Handle case where no video files found
	if len(vidPaths) == 0 {
		logger.Log(true, "⚠️  No video files found in: %s", tmpDir)
		td.MarkPlaced("⚠️ No video files found to place!")
		return
	}

	logger.Log(true, "📊 Found %d video file(s) to process", len(vidPaths))

	for i, path := range vidPaths {
		filesChecked++
		fileName := filepath.Base(path)

		logger.Log(true, "")
		logger.Log(true, "🔄 Processing file %d/%d: %s", i+1, len(vidPaths), fileName)

		// readable src for msg
		readablePath := fileName
		if len(fileName) > 10 {
			readablePath = fileName[10:]
		}

		// upd msg
		td.PlacementProgress = fmt.Sprintf("🔧 Placing ➝ %d/%d - %s", i+1, len(vidPaths), readablePath)
		shared.SaveTorrentDownload(td)

		// match and place
		msg, err := MatchAndPlaceVideo(path, outDir, index, td.ChapterRange)
		if err != nil {
			logger.Log(true, "   ❌ Error placing %s: %v", fileName, err)
			lastError = err
		} else if msg != "" {
			filesPlaced++
			logger.Log(true, "   ✅ Successfully placed file %d/%d", filesPlaced, len(vidPaths))
			//save msg for final summary
			td.PlacementFull = append(td.PlacementFull, msg)
			shared.SaveTorrentDownload(td)
		} else {
			logger.Log(true, "   ⚠️  No message returned for %s - file may not have been placed", fileName)
		}
	}

	// Create appropriate message based on results
	var placedMsg string

	logger.Log(true, "")
	logger.Log(true, "📊 Placement Summary: %d/%d files placed", filesPlaced, filesChecked)

	if filesPlaced == 0 && lastError != nil {
		placedMsg = fmt.Sprintf("❌ Failed to place any files! Last error: %v", lastError)
		logger.Log(true, "❌ %s", placedMsg)
	} else if filesPlaced == 0 {
		placedMsg = "❌ No files could be placed!"
		logger.Log(true, "❌ %s", placedMsg)
	} else if filesPlaced == len(vidPaths) {
		if filesPlaced == 1 {
			placedMsg = "✅ 1 file placed!"
		} else {
			placedMsg = fmt.Sprintf("✅ All %d files placed!", filesPlaced)
		}
		logger.Log(true, "✅ %s", placedMsg)
	} else {
		// Partial success
		placedMsg = fmt.Sprintf("⚠️ %d/%d files placed!", filesPlaced, len(vidPaths))
		logger.Log(true, "⚠️ %s - Some files could not be matched to metadata", placedMsg)
	}

	td.MarkPlaced(placedMsg)
}
