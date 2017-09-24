package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/vim-volt/go-volt/lockjson"
	"github.com/vim-volt/go-volt/pathutil"
	"github.com/vim-volt/go-volt/transaction"
)

type profileCmd struct {
	showedUsage bool
}

var profileSubCmd = make(map[string]func([]string) error)

func init() {
	cmd := profileCmd{}
	profileSubCmd["get"] = cmd.doGet
	profileSubCmd["set"] = cmd.doSet
	profileSubCmd["show"] = cmd.doShow
	profileSubCmd["new"] = cmd.doNew
	profileSubCmd["destroy"] = cmd.doDestroy
	profileSubCmd["add"] = cmd.doAdd
	profileSubCmd["rm"] = cmd.doRm
}

func Profile(args []string) int {
	cmd := profileCmd{}

	// Parse args
	args, err := cmd.parseArgs(args)
	if err != nil {
		fmt.Println(err.Error())
		return 10
	}

	if cmd.showedUsage {
		return 0
	}

	if fn, exists := profileSubCmd[args[0]]; exists {
		err = fn(args[1:])
		if err != nil {
			fmt.Println("[ERROR]", err.Error())
			return 11
		}
	}

	return 0
}

func (cmd *profileCmd) showUsage() {
	cmd.showedUsage = true
	fmt.Println(`
Usage
  profile [get]
    Get current profile name

  profile set {name}
    Set profile name

  profile show {name}
    Show profile info

  profile new {name}
    Create new profile

  profile destroy {name}
    Delete profile

  profile add {name} {repository} [{repository2} ...]
    Add one or more repositories to profile

  profile rm {name} {repository} [{repository2} ...]
    Remove one or more repositories to profile

Description
  Subcommands about profile feature
`)
}

func (cmd *profileCmd) parseArgs(args []string) ([]string, error) {
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = cmd.showUsage
	fs.Parse(args)

	if len(fs.Args()) == 0 {
		return append([]string{"get"}, fs.Args()...), nil
	}

	subCmd := fs.Args()[0]
	if _, exists := profileSubCmd[subCmd]; !exists {
		return nil, errors.New("unknown subcommand: " + subCmd)
	}
	return fs.Args(), nil
}

func (*profileCmd) doGet(_ []string) error {
	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return errors.New("failed to read lock.json: " + err.Error())
	}

	// Show profile name
	fmt.Println(lockJSON.ActiveProfile)

	return nil
}

func (cmd *profileCmd) doSet(args []string) error {
	if len(args) == 0 {
		cmd.showUsage()
		fmt.Println("[ERROR] 'volt profile set' receives profile name.")
		return nil
	}
	profileName := args[0]

	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return errors.New("failed to read lock.json: " + err.Error())
	}

	// Exit if current active_profile is same as profileName
	if lockJSON.ActiveProfile == profileName {
		fmt.Println("[INFO] Unchanged active profile '" + profileName + "'")
		return nil
	}

	// Begin transaction
	err = transaction.Create()
	if err != nil {
		return err
	}
	defer transaction.Remove()

	// Return error if profiles[]/name does not match profileName
	found := false
	for _, profile := range lockJSON.Profiles {
		if profile.Name == profileName {
			found = true
			break
		}
	}
	if !found {
		return errors.New("profile '" + profileName + "' does not exist")
	}

	// Set profile name
	lockJSON.ActiveProfile = profileName

	// Write to lock.json
	err = lockjson.Write(lockJSON)
	if err != nil {
		return err
	}

	fmt.Println("[INFO] Set active profile to '" + profileName + "'")
	return nil
}

func (cmd *profileCmd) doShow(args []string) error {
	if len(args) == 0 {
		cmd.showUsage()
		fmt.Println("[ERROR] 'volt profile show' receives profile name.")
		return nil
	}
	profileName := args[0]

	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return errors.New("failed to read lock.json: " + err.Error())
	}

	// Return error if profiles[]/name does not match profileName
	var profile *lockjson.Profile
	for i := range lockJSON.Profiles {
		if lockJSON.Profiles[i].Name == profileName {
			profile = &lockJSON.Profiles[i]
			break
		}
	}
	if profile == nil {
		return errors.New("profile '" + profileName + "' does not exist")
	}

	fmt.Println("name:", profile.Name)
	fmt.Println("load_init:", profile.LoadInit)
	fmt.Println("repos_path:")
	for _, reposPath := range profile.ReposPath {
		fmt.Println("  " + reposPath)
	}

	return nil
}

func (cmd *profileCmd) doNew(args []string) error {
	if len(args) == 0 {
		cmd.showUsage()
		fmt.Println("[ERROR] 'volt profile new' receives profile name.")
		return nil
	}
	profileName := args[0]

	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return errors.New("failed to read lock.json: " + err.Error())
	}

	// Return error if profiles[]/name matches profileName
	for _, profile := range lockJSON.Profiles {
		if profile.Name == profileName {
			return errors.New("profile '" + profileName + "' already exists")
		}
	}

	// Begin transaction
	err = transaction.Create()
	if err != nil {
		return err
	}
	defer transaction.Remove()

	// Add profile
	lockJSON.Profiles = append(lockJSON.Profiles, lockjson.Profile{
		Name:      profileName,
		ReposPath: make([]string, 0),
		LoadInit:  true,
	})

	// Write to lock.json
	err = lockjson.Write(lockJSON)
	if err != nil {
		return err
	}

	fmt.Println("[INFO] Created new profile '" + profileName + "'")

	return nil
}

func (cmd *profileCmd) doDestroy(args []string) error {
	if len(args) == 0 {
		cmd.showUsage()
		fmt.Println("[ERROR] 'volt profile destroy' receives profile name.")
		return nil
	}
	profileName := args[0]

	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return errors.New("failed to read lock.json: " + err.Error())
	}

	// Return error if profiles[]/name does not match profileName
	found := -1
	for i, profile := range lockJSON.Profiles {
		if profile.Name == profileName {
			found = i
			break
		}
	}
	if found < 0 {
		return errors.New("profile '" + profileName + "' does not exist")
	}

	// Begin transaction
	err = transaction.Create()
	if err != nil {
		return err
	}
	defer transaction.Remove()

	// Delete the specified profile
	lockJSON.Profiles = append(lockJSON.Profiles[:found], lockJSON.Profiles[found+1:]...)

	// Write to lock.json
	err = lockjson.Write(lockJSON)
	if err != nil {
		return err
	}

	fmt.Println("[INFO] Deleted profile '" + profileName + "'")

	return nil
}

func (cmd *profileCmd) doAdd(args []string) error {
	// Parse args
	profileName, reposPathList, err := cmd.parseAddArgs("add", args)

	added := make([]string, 0, len(reposPathList))

	// Read modified profile and write to lock.json
	err = cmd.transactProfile(profileName, func(profile *lockjson.Profile) {
		// Add repositories to profile if the repository does not exist
		for _, reposPath := range reposPathList {
			found := false
			for i := range profile.ReposPath {
				if profile.ReposPath[i] == reposPath {
					found = true
					break
				}
			}
			if found {
				fmt.Println("[WARN] repository '" + reposPath + "' already exists")
			} else {
				profile.ReposPath = append(profile.ReposPath, reposPath)
				added = append(added, reposPath)
			}
		}
	})
	if err != nil {
		return err
	}

	for _, reposPath := range added {
		fmt.Println("[INFO] Added repository '" + reposPath + "'")
	}

	return nil
}

func (cmd *profileCmd) doRm(args []string) error {
	// Parse args
	profileName, reposPathList, err := cmd.parseAddArgs("rm", args)

	removed := make([]string, 0, len(reposPathList))

	// Read modified profile and write to lock.json
	err = cmd.transactProfile(profileName, func(profile *lockjson.Profile) {
		// Remove repositories from profile if the repository does not exist
		for _, reposPath := range reposPathList {
			found := -1
			for i := range profile.ReposPath {
				if profile.ReposPath[i] == reposPath {
					found = i
					break
				}
			}
			if found >= 0 {
				// Remove profile.ReposPath[found]
				profile.ReposPath = append(profile.ReposPath[:found], profile.ReposPath[found+1:]...)
				removed = append(removed, reposPath)
			} else {
				fmt.Println("[WARN] repository '" + reposPath + "' does not exist")
			}
		}
	})
	if err != nil {
		return err
	}

	for _, reposPath := range removed {
		fmt.Println("[INFO] Removed repository '" + reposPath + "'")
	}

	return nil
}

func (cmd *profileCmd) parseAddArgs(subCmd string, args []string) (string, []string, error) {
	if len(args) == 0 {
		cmd.showUsage()
		fmt.Printf("[ERROR] 'volt profile %s' receives profile name and one or more repositories.\n", subCmd)
		return "", nil, nil
	}

	profileName := args[0]
	reposPathList := make([]string, 0, len(args)-1)
	for _, arg := range args[1:] {
		reposPath, err := pathutil.NormalizeRepository(arg)
		if err != nil {
			return "", nil, err
		}
		reposPathList = append(reposPathList, reposPath)
	}
	return profileName, reposPathList, nil
}

func (*profileCmd) transactProfile(profileName string, modifyProfile func(*lockjson.Profile)) error {
	// Read lock.json
	lockJSON, err := lockjson.Read()
	if err != nil {
		return errors.New("failed to read lock.json: " + err.Error())
	}

	// Return error if profiles[]/name does not match profileName
	var profile *lockjson.Profile
	for i := range lockJSON.Profiles {
		if lockJSON.Profiles[i].Name == profileName {
			profile = &lockJSON.Profiles[i]
			break
		}
	}
	if profile == nil {
		return errors.New("profile '" + profileName + "' does not exist")
	}

	// Begin transaction
	err = transaction.Create()
	if err != nil {
		return err
	}
	defer transaction.Remove()

	modifyProfile(profile)

	// Write to lock.json
	return lockjson.Write(lockJSON)
}
