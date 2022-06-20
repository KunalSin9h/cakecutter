package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"

	"github.com/AlecAivazis/survey/v2"
	"github.com/cake-cutter/cc/utils"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cutCmd)
}

var cutCmd = &cobra.Command{
	Use:     "cut",
	Short:   "Cuts an online cake (Creates a template from an online template)",
	Long:    "Cuts an online cake (Creates a template from an online template)",
	Aliases: []string{"run"},

	Args: func(cmd *cobra.Command, args []string) error {

		if len(args) < 1 {
			return fmt.Errorf("Enter the cake's name!")
		}
		if len(args) < 2 {
			return fmt.Errorf("Enter the directory's path!")
		}
		path_to_dir := args[1]
		if _, err := os.Stat(path_to_dir); os.IsNotExist(err) {
		} else {

			dir, err := os.ReadDir(path_to_dir)
			utils.Check(err)
			if len(dir) > 0 {

				fmt.Println()

				var res string

				survey.AskOne(&survey.Select{
					Message: "The directory `" + path_to_dir + "` already has some files. Are you sure you want to use it? This may have unforseen consequences.",
					Options: []string{"Yes", "No"},
					Default: "No",
				}, &res)

				if res == "No" {
					return fmt.Errorf("Aborted")
				}

				fmt.Println()

				return nil

			}
		}

		return nil

	},
	Run: func(cmd *cobra.Command, args []string) {

		path_to_dir := args[1]

		var (
			conf *utils.Config
			err  error
		)

		fmt.Println(utils.Colorize("blue", "\nCutting cake...\n "))

		utils.MakeItSpin(func() {

			resp, err := http.Get(utils.BackendURL + "/get?name=" + args[0])
			utils.Check(err)

			var j struct {
				Success int    `json:"success"`
				Error   string `json:"error"`
				Data    struct {
					Name   string `json:"name"`
					Short  string `json:"short"`
					Dsc    string `json:"dsc"`
					Author string `json:"author"`
					Cake   string `json:"cake"`
				} `json:"data"`
			}

			defer resp.Body.Close()

			b, err := io.ReadAll(resp.Body)
			utils.Check(err)

			err = json.Unmarshal([]byte(b), &j)
			utils.Check(err)

			if resp.StatusCode != 200 {
				if j.Success == 1 {
					fmt.Println(utils.Colorize("red", j.Error))
					os.Exit(1)
				}
			}

			conf, err = utils.ParseToml(j.Data.Cake)
			utils.Check(err)

		}, "Fetching cake...")

		ans := make(map[string]string)

		err = utils.CutTheQuestions(&ans, conf)
		utils.Check(err)

		utils.MakeItSpin(func() {
			err = utils.CutDir(path_to_dir, conf, ans)
			utils.Check(err)
		}, "🔪 Cutting file structure...")

		utils.MakeItSpin(func() {
			err = utils.CutFiles(path_to_dir, conf, ans)
			utils.Check(err)
		}, "🔪 Cutting files...")

		cs, err := utils.ParseCommands(conf.Commands, ans)
		utils.Check(err)

		sort.Ints(cs)

		fmt.Println("\n" + utils.Colorize("green", "These commands are going to run... If these commands seems suspicious or harmful please report them by making an issue on the repo - `https://github.com/cake-cutter/cakes.run`"))
		for _, v := range cs {
			fmt.Println(utils.Colorize("gray", "  "+fmt.Sprintf("%s", v)))
		}

		var _res string

		survey.AskOne(&survey.Select{
			Message: "Do you want to run these commands?",
			Options: []string{"Yes", "No"},
		}, &_res)

		if _res == "No" {
			fmt.Println("\n" + utils.Colorize("red", "Aborted!"))
			fmt.Println(utils.Colorize("green", "The rest of the cake has been cut."))
			os.Exit(0)
		}

		fmt.Println()

		utils.MakeItSpin(func() {
			err = utils.CutDaCommands(path_to_dir, conf.Commands, cs)
			utils.Check(err)
		}, "Cutting commands...")

		fmt.Println(utils.Colorize("green", "Successfully cut `"+path_to_dir+"`"))
	},
}
