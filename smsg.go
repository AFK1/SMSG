package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var mod_count int = 10
var mod_installed_count int = 0

var col_reg = "<div class=\"rightSectionTopTitle\">A collection of [0-9]+ items created by</div>"
var col_item_reg = "<a href=\"https://steamcommunity.com/sharedfiles/filedetails/.id=[0-9]+\"><div class=\"workshopItemTitle\">"
var appid_reg = "href=\"https://steamcommunity.com/app/[0-9]+\""
var num_regex, _ = regexp.Compile("[0-9]+")

func get_appid(page_code string) string {
	re, _ := regexp.Compile(appid_reg)
	return num_regex.FindAllString(re.FindAllString(page_code, -1)[0], -1)[0]
}

func get_code(url string) string {
	return num_regex.FindAllString(url, -1)[0]
}

func is_collection(page_code string) bool {
	matched, _ := regexp.MatchString(col_reg, page_code)
	return matched
}

func all_collection_ids(page_code string) []string {
	var urls []string
	re, _ := regexp.Compile(col_item_reg)
	res := re.FindAllString(page_code, -1)
	for _, url_raw := range res {
		urls = append(urls, num_regex.FindAllString(url_raw, -1)[0])
	}
	return urls
}

func get_req(url string) string {
	resp, _ := http.Get(url)
	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
}

func start(main_url string, progress *widget.ProgressBar) {
	workshoppage := get_req(main_url)
	os.MkdirAll("download", os.ModePerm)
	os.Remove("script.txt")
	outf, _ := os.OpenFile("script.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	outf.WriteString("@ShutdownOnFailedCommand 0\n@NoPromptForPassword 1\nlogin anonymous\n")

	path, _ := os.Getwd()
	if runtime.GOOS == "windows" {
		outf.WriteString(fmt.Sprintf("force_install_dir %s\\download\n", path))
	} else {
		outf.WriteString(fmt.Sprintf("force_install_dir %s/download\n", path))
	}
	if is_collection(workshoppage) {
		appid := get_appid(workshoppage)
		ids := all_collection_ids(workshoppage)
		mod_count = len(ids)
		for _, modid := range ids {
			outf.WriteString(fmt.Sprintf("workshop_download_item %s %s validate\n", appid, modid))
			mod_installed_count += 1
			progress.SetValue(float64(mod_installed_count) / float64(mod_count))
		}
	} else {
		modid := get_code(main_url)
		appid := get_appid(workshoppage)
		outf.WriteString(fmt.Sprintf("workshop_download_item %s %s validate\n", appid, modid))
		mod_installed_count += 1
		progress.SetValue(float64(mod_installed_count) / float64(mod_count))
	}
	outf.WriteString("quit\n")
}

func main() {
	a := app.New()
	a.Settings().SetTheme(theme.DarkTheme())
	w := a.NewWindow("Swt SMSG")

	w.Resize(fyne.NewSize(600, 100))

	progress := widget.NewProgressBar()
	progress.SetValue(0)

	input := widget.NewEntry()
	input.SetPlaceHolder("link")

	w.SetContent(container.NewVBox(
		progress,
		input,
		widget.NewButton("Download", func() {
			start(input.Text, progress)
			os.Exit(0)
		}),
	))

	w.ShowAndRun()
}
