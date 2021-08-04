package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/crispgm/kickertool-analyzer/model"
	"github.com/crispgm/kickertool-analyzer/stat"
	monsterdyb "github.com/crispgm/kickertool-analyzer/stat/monster_dyb"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

// flags
var (
	mode    string
	nocolor bool
	player  string
	files   []string
)

func main() {
	flag.StringVar(&mode, "mode", "", "Stat mode. Supported: mts")
	flag.BoolVar(&nocolor, "nocolor", false, "Disable colors")
	flag.StringVar(&player, "player", "", "Players' data file")
	flag.Parse()

	if nocolor {
		color.NoColor = true
	}

	// check mode
	if supported, ok := stat.SupportedStat[mode]; !ok || !supported {
		fmt.Println("Invalid mode", mode)
		os.Exit(1)
	}
	fmt.Println("Statistics mode:", mode)

	// load players
	if len(player) == 0 {
		fmt.Println("No given player file")
		os.Exit(1)
	}
	fmt.Println("Loading players ...")
	players, err := parsePlayer(player)
	if err != nil {
		fmt.Println("Load players failed:", err)
	}

	// load tournaments
	files := flag.Args()
	if len(files) == 0 {
		fmt.Println("No given files")
		os.Exit(1)
	}
	fmt.Println("Loading tournaments ...")
	var tournaments []model.Tournament

	// parsing
	for _, fn := range files {
		fmt.Println("Parsing tournaments data:", fn)
		t, err := parseTournament(fn)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		tournaments = append(tournaments, *t)
	}

	// calculating
	var statInfo stat.BaseStat
	statInfo = monsterdyb.NewMultipleTournamentStats(tournaments, players)
	fmt.Println("Outputing ...")
	if statInfo.ValidMode() {
		data := statInfo.Output()
		outputTable(data)
	}
}

func parseTournament(fn string) (*model.Tournament, error) {
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}

	var tournament model.Tournament
	err = json.Unmarshal(data, &tournament)
	if err != nil {
		return nil, err
	}
	return &tournament, err
}

func parsePlayer(fn string) ([]model.EntityPlayer, error) {
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}

	var players []model.EntityPlayer
	err = json.Unmarshal(data, &players)
	if err != nil {
		return nil, err
	}
	return players, err
}

func outputTable(data []model.EntityPlayer) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Num", "Won", "Lost", "G+", "G-", "G±", "WR%", "PPG", "TPG"})
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_DEFAULT,
		tablewriter.ALIGN_DEFAULT,
		tablewriter.ALIGN_DEFAULT,
		tablewriter.ALIGN_DEFAULT,
		tablewriter.ALIGN_DEFAULT,
		tablewriter.ALIGN_DEFAULT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_DEFAULT,
		tablewriter.ALIGN_DEFAULT,
	})
	for _, d := range data {
		goalDiff := fmt.Sprintf("%d", d.GoalDiff)
		winRate := fmt.Sprintf("%.0f%%", d.WinRate)
		if d.GoalDiff > 0 {
			goalDiff = color.GreenString(goalDiff)
		} else if d.GoalDiff < 0 {
			goalDiff = color.RedString(goalDiff)
		} else {
			goalDiff = color.YellowString(goalDiff)
		}

		if d.WinRate >= 80.0 {
			winRate = color.RedString(winRate)
		} else if d.WinRate >= 70.0 {
			winRate = color.MagentaString(winRate)
		} else if d.WinRate >= 60.0 {
			winRate = color.GreenString(winRate)
		} else if d.WinRate >= 50.0 {
			winRate = color.YellowString(winRate)
		}
		table.Append([]string{
			d.Name,
			fmt.Sprintf("%d", d.Played),
			fmt.Sprintf("%d", d.Won),
			fmt.Sprintf("%d", d.Lost),
			fmt.Sprintf("%d", d.Goals),
			fmt.Sprintf("%d", d.GoalsIn),
			goalDiff,
			winRate,
			fmt.Sprintf("%.2f", d.PointsPerGame),
			fmt.Sprintf("%02d:%02d", d.TimePerGame/60, d.TimePerGame%60),
		})
	}

	table.Render()
}
