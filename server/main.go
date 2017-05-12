package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/conformal/gotk3/gtk"
	"github.com/doxxan/appindicator"
	"github.com/doxxan/appindicator/gtk-extensions/gotk3"
	"golang.org/x/exp/utf8string"
)

var (
	downloadDir, archivePath, icon string
	count                          int
	titleRe                        *regexp.Regexp = regexp.MustCompile("\\[download]\\s+Destination:\\s+(.+)")
	progressRe                     *regexp.Regexp = regexp.MustCompile("\\[download]\\s+(\\d+\\.\\d)%")
	logFile, cmdLogFile            string
	logger                         *log.Logger
	cmdLogger                      *os.File
)

func init() {
	flag.StringVar(&downloadDir, "dir", "", "Path to the download directory")
	flag.StringVar(&archivePath, "archive", "", "Path to the archive file")
	flag.StringVar(&icon, "icon", "", "Path to the icon")
	flag.StringVar(&logFile, "log", "", "Log file")
	flag.StringVar(&cmdLogFile, "cmdlog", "", "Command log file")

	flag.Parse()

	logger = log.New(os.Stdout, "", log.LstdFlags)
	cmdLogger = os.Stdout
}

func ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	url := "https://www.youtube.com/" + req.URL.Path[1:]
	playlist, _ := strconv.ParseBool(req.URL.Query().Get("playlist"))

	go download(url, playlist)
}

func main() {
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			logger.Fatalln(err)
		}
		defer f.Close()
		logger = log.New(f, "", log.LstdFlags)
	}

	if cmdLogFile != "" {
		f, err := os.OpenFile(cmdLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			logger.Fatalln(err)
		}
		defer f.Close()
		cmdLogger = f
	}

	gtk.Init(nil)

	go func() {
		gtk.Main()
	}()

	http.HandleFunc("/", ServeHTTP)
	logger.Println("Listening...")
	if err := http.ListenAndServe(":12345", nil); err != nil {
		logger.Fatalln(err)
	}
}

func download(url string, playlist bool) {
	logger.Println("Request", url, playlist)

	notify("Download (playlist:", playlist, ")")

	count++
	id := fmt.Sprintf("indicator-youtube-dl-%v", count)
	indicator := gotk3.NewAppIndicator(id, icon, appindicator.CategorySystemServices)

	indicator.SetStatus(appindicator.StatusActive)

	menu, err := gtk.MenuNew()
	if err != nil {
		panic(err)
	}

	menuItem, err := gtk.MenuItemNewWithLabel("")
	if err != nil {
		panic(err)
	}

	menu.Append(menuItem)

	menuItem.Show()
	indicator.SetMenu(menu)

	var playlistArg string
	if playlist {
		playlistArg = "--yes-playlist"
	} else {
		playlistArg = "--no-playlist"
	}

	cmd := exec.Command("youtube-dl",
		playlistArg,
		"--limit-rate", "1M",
		"--download-archive", archivePath,
		"--extract-audio",
		"--audio-quality", "0",
		"--audio-format", "aac",
		"--output", filepath.Join(downloadDir, "%(title)s.%(ext)s"),
		url)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	out := io.TeeReader(stdout, cmdLogger)

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	var title string
	b := make([]byte, 200)
	for {
		n, err := out.Read(b)
		if err != nil {
			break
		}

		read := b[:n]

		if submatches := titleRe.FindSubmatch(read); len(submatches) > 0 {
			path := string(submatches[1])
			file := filepath.Base(path)
			wholeTitle := utf8string.NewString(strings.TrimSuffix(file, filepath.Ext(file)))

			if wholeTitle.RuneCount() < 30 {
				title = wholeTitle.String()
			} else {
				title = wholeTitle.Slice(0, 30)
			}

			logger.Println("Title", wholeTitle)

			continue
		}

		if submatches := progressRe.FindSubmatch(read); len(submatches) > 0 {
			progress, _ := strconv.ParseFloat(string(submatches[1]), 64)

			indicator.SetLabel(fmt.Sprintf("[%.2f%%]%s", progress, title), "")
		}
	}

	if err := cmd.Wait(); err != nil {
		logger.Println("Failed", url, err)
		notify(err)
	}

	logger.Println("Done", title)
	notify("Done ", title)

	indicator.SetStatus(appindicator.StatusPassive)
}

func notify(msg ...interface{}) {
	exec.Command("notify-send", "-u", "critical", fmt.Sprint(msg...)).Run()
}
