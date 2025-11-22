# Things
## TODO
- [x] Make it selfhostable
- [x] build a UI
- [x] Allow downloading via a torrent client
- [x] Testing
- [x] Download changes in the UI
- [x] List changes in the UI
- [x] Download directory setup via torrent client
- [x] Download testing (pass)
- [x] Browser caching or database
- [x] Torrent connectivity
  - [x] Qbittorrent
  - [ ] Transmission (Untested)
  - [ ] Deluge (Untested)
- [x] Activity tab not auto polling % not working after refresh
- [x] Hardlinking like sonarr (import working)
- [x] hardlinking confirmation like sonarr (not actually working, once deleted from qbittorrent with also delete content files it doesnt delete it from the downloads folder and says permission denied and when tried manually, should auto delete once seed ratio is reached as well)

- [-] seasons are not getting imported correctly (single episodes are but not the entire season) it says imported but it did not actually get imported
- [-] items from activity list should be removed once the import is complete and a default seed ratio of 0.6 should be assigned to downloads so theyll be gone from the downloads folder once it is reached
- [-] episode search is not working correctly (not finding items even when its there it should look for One Pace and episode numbers like `One Pace 304-306`)
- [-] Search all should automatically search for all and queue them all. Currently it doesnt show up on most seasons. if the entire season is not available it should search for all episodes and queue them
- [-] num files placed seems wrong in activity section
- [-] proper logging so we see when an import or download or something has failed. Full logging should be done
- [-] remove stray videos logic and fail the import if no matches

- [ ] metadata for each season and episode must be inserted when downloading only
- [ ] some episodes still showing as [One Pace something something in jellyfin repeated ones. They need to be removed somehow
- [ ] remove download and allow setting a prefered resolution and then just do search and auto search

- [ ] UI overhaul
  - [x] Proper sidebar like sonarr (implementation didnt work)
  - [x] Loading spinner
  - [x] Background color
  - [x] Font change
  - [x] Show clickable items with an underscore like a href
  - [x] Show Season number in the arcs list and sort by season
  - [ ] fix double download status in activity
  - [ ] show importing after done dont just say ready to organize (statusing is not proper)
  - [ ] better feedback while clicking things
  - [ ] update toast timeout to 15s and make them dismissable
  - [ ] Log Viewing
  - [ ] Images for Arcs
  - [ ] Settings and System overhaul to be similar to sonarr
  - [ ] Mobile UI (collapsible sidebar)
  - [ ] Name change
  - [ ] Logo change
  - [ ] Icon and notation changes

- [ ] use go templ (check htmx-go-templ project)
- [ ] Strip the CLI functionality
- [ ] Unit Tests
- [ ] Nixos based deployment
- [ ] Complementry updates based on other changes and user testing (final testing)
- [ ] Create a proper readme
- [ ] Finalize
  - [ ] docker image
  - [ ] nixpkgs image
  - [ ] reddit post
  - [ ] github settings
  - [ ] build pipeline
- [ ] later
  - [ ] Add other indexers than nyaa
  - [ ] allow adding a custom seed ratio (default is 0.6)
  - [ ] clear all functionality to start from top


## Notes
  jellyfin theme > Dashboard > General > Custom CSS code > `@import url('https://cdn.jsdelivr.net/gh/stpnwf/ZestyTheme@latest/theme.css');`

