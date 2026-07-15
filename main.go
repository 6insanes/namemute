// NameMute installer: installs the NameMute UI mod into Civilization V.
//
// The primary mechanism injects the mod's files into Assets/DLC/Expansion2
// (Brave New World), the same technique used by JdH's CiV MP Mod Manager, a
// known-working community tool for using mods in Civ5 multiplayer. A
// standalone top-level DLC folder (the trick older EUI docs describe) turned
// out not to work on the current game build — Civ5 keeps a fixed list of
// recognized Assets/DLC subfolder names, and an invented one is silently
// ignored. Expansion2 itself is always scanned, so files placed inside it
// (in the same relative layout the base game already uses) get picked up.
// Files that already live under Expansion2 get overwritten in place (with a
// backup); files that only exist in the base game layer are added as new
// files nested under Expansion2 instead, so the originals are never touched.
//
// As a fallback, -direct-replace overwrites the 17 base game Lua files in
// their original location instead (with backups). Since NameMute is a pure
// UI mod (AffectsSavedGames=0, no EntryPoints/DB changes), either approach
// is safe.
package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

//go:embed assets
var modFiles embed.FS

const modAssetsRoot = "assets"

// directReplaceTargets maps each mod file to the base game file it overrides
// when running in -direct-replace mode.
var directReplaceTargets = map[string][]string{
	"GameplayUtilities.lua": {
		"Assets/Gameplay/Lua/GameplayUtilities.lua",
		"steamassets/assets/gameplay/lua/gameplayutilities.lua",
	},
	"DiscussionDialog.lua": {
		"Assets/DLC/Expansion2/UI/InGame/LeaderHead/DiscussionDialog.lua",
		"steamassets/assets/dlc/expansion2/ui/ingame/leaderhead/discussiondialog.lua",
	},
	"DiploCorner.lua": {
		"Assets/DLC/Expansion2/UI/InGame/WorldView/DiploCorner.lua",
		"steamassets/assets/dlc/expansion2/ui/ingame/worldview/diplocorner.lua",
	},
	"MPTurnPanel.lua": {
		"Assets/UI/InGame/WorldView/MPTurnPanel.lua",
		"steamassets/assets/ui/ingame/worldview/mpturnpanel.lua",
	},
	"MPList.lua": {
		"Assets/UI/InGame/WorldView/MPList.lua",
		"steamassets/assets/ui/ingame/worldview/mplist.lua",
	},
	"TradeLogic.lua": {
		"Assets/DLC/Expansion2/UI/InGame/WorldView/TradeLogic.lua",
		"steamassets/assets/dlc/expansion2/ui/ingame/worldview/tradelogic.lua",
	},
	"DiploList.lua": {
		"Assets/DLC/Expansion2/UI/InGame/DiploList.lua",
		"steamassets/assets/dlc/expansion2/ui/ingame/diplolist.lua",
	},
	"DiploCurrentDeals.lua": {
		"Assets/UI/InGame/Popups/DiploCurrentDeals.lua",
		"steamassets/assets/ui/ingame/popups/diplocurrentdeals.lua",
	},
	"DiploVotePopup.lua": {
		"Assets/DLC/Expansion2/UI/InGame/Popups/DiploVotePopup.lua",
		"steamassets/assets/dlc/expansion2/ui/ingame/popups/diplovotepopup.lua",
	},
	"DiploRelationships.lua": {
		"Assets/DLC/Expansion2/UI/InGame/Popups/DiploRelationships.lua",
		"steamassets/assets/dlc/expansion2/ui/ingame/popups/diplorelationships.lua",
	},
	"DiploGlobalRelationships.lua": {
		"Assets/DLC/Expansion2/UI/InGame/Popups/DiploGlobalRelationships.lua",
		"steamassets/assets/dlc/expansion2/ui/ingame/popups/diploglobalrelationships.lua",
	},
	"VoteResultsPopup.lua": {
		"Assets/DLC/Expansion2/UI/InGame/Popups/VoteResultsPopup.lua",
		"steamassets/assets/dlc/expansion2/ui/ingame/popups/voteresultspopup.lua",
	},
	"EspionageOverview.lua": {
		"Assets/DLC/Expansion2/UI/InGame/Popups/EspionageOverview.lua",
		"steamassets/assets/dlc/expansion2/ui/ingame/popups/espionageoverview.lua",
	},
	"VictoryProgress.lua": {
		"Assets/DLC/Expansion2/UI/InGame/Popups/VictoryProgress.lua",
		"steamassets/assets/dlc/expansion2/ui/ingame/popups/victoryprogress.lua",
	},
	"PlotMouseOverInclude.lua": {
		"Assets/DLC/Expansion2/UI/InGame/PlotMouseoverInclude.lua",
		"steamassets/assets/dlc/expansion2/ui/ingame/plotmouseoverinclude.lua",
	},
	"UnitFlagManager.lua": {
		"Assets/DLC/Expansion2/UI/InGame/UnitFlagManager.lua",
		"steamassets/assets/dlc/expansion2/ui/ingame/unitflagmanager.lua",
	},
	"GameMenu.lua": {
		"Assets/UI/InGame/Menus/GameMenu.lua",
		"steamassets/assets/ui/ingame/menus/gamemenu.lua",
	},
}

// directReplaceOrder keeps output deterministic (map iteration order is not).
var directReplaceOrder = []string{
	"GameplayUtilities.lua", "DiscussionDialog.lua", "DiploCorner.lua", "MPTurnPanel.lua",
	"MPList.lua", "TradeLogic.lua", "DiploList.lua", "DiploCurrentDeals.lua",
	"DiploVotePopup.lua", "DiploRelationships.lua", "DiploGlobalRelationships.lua",
	"VoteResultsPopup.lua", "EspionageOverview.lua", "VictoryProgress.lua",
	"PlotMouseOverInclude.lua", "UnitFlagManager.lua", "GameMenu.lua",
}

// expansion2Roots are candidate locations of the game's Expansion2 (Brave
// New World) DLC folder, which the engine always scans (unlike a made-up
// top-level DLC folder name, which turned out to be ignored — Civ5 keeps a
// fixed list of recognized Assets/DLC subfolders). This mirrors the
// technique used by JdH's CiV MP Mod Manager, a known-working community tool
// for using mods in Civ5 multiplayer: it copies mod files directly into
// "CiV Install Dir\assets\dlc\expansion2" rather than a new DLC folder.
var expansion2Roots = []string{"Assets/DLC/Expansion2", "steamassets/assets/dlc/expansion2"}

// expansion2RelPath returns modFileName's path relative to the Expansion2
// root. Files that already natively live under Expansion2 keep that exact
// path (so installing there overwrites the original, same as -direct-replace
// would). Files that only exist in the base game layer get the analogous
// subpath nested under Expansion2 instead — a new file, nothing overwritten.
func expansion2RelPath(modFileName string) string {
	base := directReplaceTargets[modFileName][0] // e.g. "Assets/DLC/Expansion2/UI/.../X.lua" or "Assets/UI/.../X.lua"
	rel := strings.TrimPrefix(base, "Assets/DLC/Expansion2/")
	if rel != base {
		return rel
	}
	return strings.TrimPrefix(base, "Assets/")
}

type candidate struct {
	label string
	path  string
}

func main() {
	uninstall := flag.Bool("uninstall", false, "remove the mod and restore original game files")
	directReplaceMode := flag.Bool("direct-replace", false, "overwrite base game files in their original location instead of injecting into Assets/DLC/Expansion2")
	yes := flag.Bool("yes", false, "don't ask for confirmation")
	gameDirOverride := flag.String("game-dir", "", "manual path to the Civilization V install directory (contains Assets/)")
	flag.Parse()

	fmt.Println("NameMute installer")
	fmt.Println("===================")

	gameDirs := detectGameDirs(*gameDirOverride)

	if len(gameDirs) == 0 {
		fmt.Println("\nНе удалось найти папку установки Civilization V автоматически.")
		fmt.Println("Укажите её вручную: -game-dir \"путь\"")
		exitPause(1)
	}

	fmt.Println("\nНайденные папки установки игры:")
	for _, c := range gameDirs {
		fmt.Printf("  [%s] %s\n", c.label, c.path)
	}

	if *uninstall {
		if !*yes && !confirm("\nУдалить NameMute и восстановить оригинальные файлы игры?") {
			fmt.Println("Отменено.")
			exitPause(0)
		}
		doUninstall(gameDirs)
		exitPause(0)
	}

	mode := "внутрь Assets/DLC/Expansion2"
	if *directReplaceMode {
		mode = "прямая замена файлов"
	}
	if !*yes && !confirm(fmt.Sprintf("\nУстановить NameMute? Режим: %s.", mode)) {
		fmt.Println("Отменено.")
		exitPause(0)
	}
	doInstall(gameDirs, *directReplaceMode)
	exitPause(0)
}

func confirm(prompt string) bool {
	fmt.Print(prompt + " [y/N]: ")
	var answer string
	fmt.Scanln(&answer)
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes" || answer == "д" || answer == "да"
}

func exitPause(code int) {
	if runtime.GOOS == "windows" {
		fmt.Print("\nНажмите Enter для выхода...")
		var s string
		fmt.Scanln(&s)
	}
	os.Exit(code)
}

// --- detection -------------------------------------------------------------

func detectGameDirs(override string) []candidate {
	if override != "" {
		return []candidate{{"вручную", override}}
	}
	seen := map[string]bool{}
	var out []candidate
	for _, root := range steamLibraryRoots() {
		p := filepath.Join(root, "steamapps", "common", "Sid Meier's Civilization V")
		if dirExists(filepath.Join(p, "Assets")) && !seen[p] {
			seen[p] = true
			out = append(out, candidate{"Steam", p})
		}
	}
	return out
}

// steamLibraryRoots returns every Steam library folder it can find: the
// default install location(s) plus anything listed in libraryfolders.vdf.
func steamLibraryRoots() []string {
	var bases []string
	home, _ := os.UserHomeDir()

	if runtime.GOOS == "windows" {
		for _, pf := range []string{os.Getenv("ProgramFiles(x86)"), os.Getenv("ProgramFiles")} {
			if pf != "" {
				bases = append(bases, filepath.Join(pf, "Steam"))
			}
		}
		// Common alternate drive locations.
		for _, drive := range []string{"C:", "D:", "E:", "F:", "G:"} {
			bases = append(bases, filepath.Join(drive+string(os.PathSeparator), "Steam"))
			bases = append(bases, filepath.Join(drive+string(os.PathSeparator), "SteamLibrary"))
		}
	} else {
		bases = append(bases,
			filepath.Join(home, ".steam", "steam"),
			filepath.Join(home, ".steam", "root"),
			filepath.Join(home, ".local", "share", "Steam"),
		)
		// Common alternate mount locations for secondary libraries.
		matches, _ := filepath.Glob("/mnt/*/Steam")
		bases = append(bases, matches...)
		matches, _ = filepath.Glob("/media/*/*/Steam")
		bases = append(bases, matches...)
	}

	seen := map[string]bool{}
	var roots []string
	pushRoot := func(p string) {
		p = filepath.Clean(p)
		if !seen[p] {
			seen[p] = true
			roots = append(roots, p)
		}
	}

	for _, b := range bases {
		if !dirExists(b) {
			continue
		}
		pushRoot(b)
		for _, extra := range parseLibraryFolders(filepath.Join(b, "steamapps", "libraryfolders.vdf")) {
			pushRoot(extra)
		}
	}
	return roots
}

var vdfPathRe = regexp.MustCompile(`"path"\s*"((?:[^"\\]|\\.)*)"`)

func parseLibraryFolders(vdfPath string) []string {
	data, err := os.ReadFile(vdfPath)
	if err != nil {
		return nil
	}
	var out []string
	for _, m := range vdfPathRe.FindAllStringSubmatch(string(data), -1) {
		p := strings.ReplaceAll(m[1], `\\`, `\`)
		out = append(out, p)
	}
	return out
}

func dirExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}

// --- install / uninstall ----------------------------------------------------

func doInstall(gameDirs []candidate, directReplaceMode bool) {
	if directReplaceMode {
		fmt.Println("\n--- Прямая замена файлов игры ---")
		for _, g := range gameDirs {
			for _, name := range directReplaceOrder {
				p := findExisting(g.path, directReplaceTargets[name])
				if p == "" {
					fmt.Printf("  пропущено (файл не найден): %s\n", directReplaceTargets[name][0])
					continue
				}
				status, err := directReplace(p, name)
				if err != nil {
					fmt.Printf("  ОШИБКА (%s): %v\n", p, err)
					continue
				}
				fmt.Printf("  %s: %s\n", status, p)
			}
		}
		fmt.Println("\nГотово. Полностью закройте игру (если она запущена) и запустите заново.")
		return
	}

	fmt.Println("\n--- Установка внутрь Assets/DLC/Expansion2 ---")
	for _, g := range gameDirs {
		root := findExpansion2Root(g.path)
		for _, name := range directReplaceOrder {
			target := filepath.Join(root, filepath.FromSlash(expansion2RelPath(name)))
			status, err := expansion2Install(target, name)
			if err != nil {
				fmt.Printf("  ОШИБКА (%s): %v\n", target, err)
				continue
			}
			fmt.Printf("  %s: %s\n", status, target)
		}
	}
	fmt.Println("\nГотово. Полностью закройте игру (если она запущена) и запустите заново.")
	fmt.Println("Ничего включать в меню Mods не нужно — мод подхватывается автоматически вместе с Expansion2.")
}

func doUninstall(gameDirs []candidate) {
	fmt.Println("\n--- Удаление файлов из Assets/DLC/Expansion2 ---")
	for _, g := range gameDirs {
		root := findExpansion2Root(g.path)
		for _, name := range directReplaceOrder {
			target := filepath.Join(root, filepath.FromSlash(expansion2RelPath(name)))
			expansion2Uninstall(target)
		}
	}

	fmt.Println("\n--- Восстановление файлов (после -direct-replace, если применялось) ---")
	for _, g := range gameDirs {
		for _, name := range directReplaceOrder {
			restoreOne(findExisting(g.path, directReplaceTargets[name]))
		}
	}
	fmt.Println("\nГотово.")
}

func restoreOne(p string) {
	if p == "" {
		return
	}
	backup := p + ".namemute_backup"
	if !fileExists(backup) {
		return
	}
	data, err := os.ReadFile(backup)
	if err != nil {
		fmt.Printf("  ОШИБКА чтения бэкапа (%s): %v\n", backup, err)
		return
	}
	if err := os.WriteFile(p, data, 0644); err != nil {
		fmt.Printf("  ОШИБКА восстановления (%s): %v\n", p, err)
		return
	}
	fmt.Printf("  восстановлено: %s\n", p)
}

func findExisting(gameDir string, relAlternatives []string) string {
	for _, rel := range relAlternatives {
		p := filepath.Join(gameDir, filepath.FromSlash(rel))
		if fileExists(p) {
			return p
		}
	}
	return ""
}

// findExpansion2Root returns the game's Assets/DLC/Expansion2 folder,
// preferring whichever casing variant already exists on disk.
func findExpansion2Root(gameDir string) string {
	for _, rel := range expansion2Roots {
		p := filepath.Join(gameDir, filepath.FromSlash(rel))
		if dirExists(p) {
			return p
		}
	}
	return filepath.Join(gameDir, filepath.FromSlash(expansion2Roots[0]))
}

// expansion2Install writes modFileName's content to target. If target
// already exists (we're overwriting a genuine Expansion2 file), the
// original is backed up first, same as -direct-replace. If target doesn't
// exist yet, this is a pure addition — nothing to back up, and uninstall
// will simply delete it again.
func expansion2Install(target, modFileName string) (status string, err error) {
	newContent, err := fs.ReadFile(modFiles, modAssetsRoot+"/"+modFileName)
	if err != nil {
		return "", err
	}

	if current, err := os.ReadFile(target); err == nil {
		if string(current) == string(newContent) {
			return "уже установлено", nil
		}
		backup := target + ".namemute_backup"
		if !fileExists(backup) {
			if err := os.WriteFile(backup, current, 0644); err != nil {
				return "", fmt.Errorf("бэкап: %w", err)
			}
		}
		if err := os.WriteFile(target, newContent, 0644); err != nil {
			return "", err
		}
		return "заменено", nil
	}

	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(target, newContent, 0644); err != nil {
		return "", err
	}
	return "добавлено", nil
}

// expansion2Uninstall reverses expansion2Install: restores target from
// backup if one exists (it was a genuine overwrite), or deletes target if
// its content still matches the mod (it was a pure addition).
func expansion2Uninstall(target string) {
	backup := target + ".namemute_backup"
	if fileExists(backup) {
		restoreOne(target)
		return
	}
	if !fileExists(target) {
		return
	}
	name := filepath.Base(target)
	modContent, err := fs.ReadFile(modFiles, modAssetsRoot+"/"+name)
	if err != nil {
		return
	}
	current, err := os.ReadFile(target)
	if err != nil || string(current) != string(modContent) {
		return
	}
	if err := os.Remove(target); err != nil {
		fmt.Printf("  ОШИБКА удаления (%s): %v\n", target, err)
		return
	}
	fmt.Printf("  удалено: %s\n", target)
}

// directReplace overwrites the base game file at path with the mod's version
// of modFileName, backing up the original on first write. Idempotent: if the
// file already matches the mod content, it does nothing.
func directReplace(path, modFileName string) (status string, err error) {
	newContent, err := fs.ReadFile(modFiles, modAssetsRoot+"/"+modFileName)
	if err != nil {
		return "", err
	}

	current, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if string(current) == string(newContent) {
		return "уже заменено", nil
	}

	backup := path + ".namemute_backup"
	if !fileExists(backup) {
		if err := os.WriteFile(backup, current, 0644); err != nil {
			return "", fmt.Errorf("бэкап: %w", err)
		}
	}

	if err := os.WriteFile(path, newContent, 0644); err != nil {
		return "", err
	}
	return "заменено", nil
}
