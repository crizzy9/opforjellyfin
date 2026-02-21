# Things

## Prompt
There are some major issues in the opforjellyfin repository that need to be addressed
- Import of the video after download is not happening correctly. it takes the same video and imports it for all episodes. Each episodes must be imported separately making sure the right episode is being imported
- Currently in the UI there is no option for sub and dub based downloads. Lets add a new settings option for this at the global level and make sure we always adhere to it. When doing a search it should show if its a sub or dub based download and allow the user to filter based on it, along with the quality of the video 720p 1080p etc


Looks like there is still a bunch of problems here
- import is still broken. it downloads the right episode but doesnt import the right video in its place in the actual library. It took a video i had in the download folder earlier and linked that to the first episode and then all the other episodes were also linked to the same episode not the actual one that was downloaded. Can you please do a deep dive on this. The whole thing needs to be revamped. Please show me exactly how this import works, how it differs for different scenarios like download arc vs download a single episode vs download multiple episode. How do you retrieve the save path location for the hardlinking of the episode from the download directory to the One Pace season library directory
- A log of things broken after gotempl modal not working. needs to show if sub/dub and resolution in the modal popup. Also change the icons for auto search and interactive search like sonarr
- a similar setting for resolution preference should be added (or just default to highest resolution/1080p)
- Main arc page not working sometimes.
- Activity needs to show number done and a history with logs of where it copied and what all it did. esentiially allowing you to edit the metadata as you please if required (this is for future)
- should allow editing failed imports as well
- improve logging. add explicit logging for all main tasks. like imported, downloaded, torrent link, save path, hardlinked etc all the logs i need in docker

In addition to this i need to have a way to deploy this using my flake.nix on a nixos machine. Allow it to be configurable in a nix homelab environment. Additionally lets add a template to wrap this service with a gluetun vpn tunnel and reverse proxy it with traefik in the readme for clarity of implementation


## TODO

- [-] Add options for Sub and Dub w/o sub for downloads as settings for search
- [-] seasons are not getting imported correctly (single episodes are but not the entire season) it says imported but it did not actually get imported
- [-] items from activity list should be removed once the import is complete and a default seed ratio of 0.6 should be assigned to downloads so theyll be gone from the downloads folder once it is reached
- [-] episode search is not working correctly (not finding items even when its there it should look for One Pace and episode numbers like `One Pace 304-306`)
- [-] Search all should automatically search for all and queue them all. Currently it doesnt show up on most seasons. if the entire season is not available it should search for all episodes and queue them
- [-] num files placed seems wrong in activity section
- [-] proper logging so we see when an import or download or something has failed. Full logging should be done
- [-] remove stray videos logic and fail the import if no matches

- [ ] metadata for each season and episode must be inserted when downloading only
- [ ] some episodes still showing as One Pace something something in jellyfin repeated ones. They need to be removed somehow
- [ ] remove download and allow setting a prefered resolution and then just do search and auto search

- [ ] UI overhaul
  - [x] Proper sidebar like sonarr (implementation didnt work)
  - [x] Loading spinner
  - [x] Background color
  - [x] Font change
  - [x] Show clickable items with an underscore like a href
  - [x] Show Season number in the arcs list and sort by season
  - [x] fix double download status in activity
  - [x] show importing after done dont just say ready to organize (statusing is not proper)
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

## Notes
  jellyfin theme > Dashboard > General > Custom CSS code > `@import url('https://cdn.jsdelivr.net/gh/stpnwf/ZestyTheme@latest/theme.css');`

