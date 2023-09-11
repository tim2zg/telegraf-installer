package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type content struct {
	TelegrafStable struct {
		Name      string `json:"name"`
		Title     string `json:"title"`
		Version   string `json:"version"`
		Downloads []struct {
			Platform string   `json:"platform"`
			Ref      string   `json:"ref"`
			Code     []string `json:"code"`
			Link     string   `json:"link,omitempty"`
			Sha256   string   `json:"sha256,omitempty"`
		} `json:"downloads"`
	} `json:"telegraf_stable"`
}

func main() {
	basePath := "C:\\Program Files\\Telegraf\\"

	// get current installed version
	currentVersion := getCurrentInstalledVersion(basePath)
	currentVersion = "v" + currentVersion

	c := getCurrentVersion()

	latestVersion := c.TelegrafStable.Version

	if currentVersion != latestVersion {
		fmt.Printf("Current Version: %s\n", currentVersion)
		fmt.Printf("Latest Version: %s\n", latestVersion)
		updateTelegraf(basePath, latestVersion)
	} else {
		fmt.Printf("Current Version: %s\n", currentVersion)
		fmt.Printf("Latest Version: %s\n", latestVersion)
		fmt.Printf("You are up to date!\n")
	}

}

func updateTelegraf(basePath string, latestVersion string) {
	// stop & uninstall service
	fmt.Printf("Stopping & uninstalling service...\n")
	handleService(basePath)
	// download latest version
	fmt.Printf("Downloading latest version...\n")
	downloadLatestVersion(latestVersion)
	// extract zip
	fmt.Printf("Extracting zip...\n")
	unzip(basePath)
	// copy files to basePath
	fmt.Printf("Copying files...\n")
	copyStuff(basePath, latestVersion)
	// cleanup
	fmt.Printf("Cleaning up...\n")
	cleanup(basePath, latestVersion)
	// restart service
	fmt.Printf("Restarting service...\n")
	restartService(basePath)
}

func handleService(path string) {
	cmd := exec.Command(path+"telegraf.exe", "--service", "stop")
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", out)

	cmd = exec.Command(path+"telegraf.exe", "--service", "uninstall")
	out, err = cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", out)
}

func restartService(path string) {
	cmd := exec.Command(path+"telegraf.exe", "--service", "install")
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", out)

	cmd = exec.Command(path+"telegraf.exe", "--service", "start")
	out, err = cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", out)
}

func cleanup(path string, version string) {
	// delete zip
	err := os.Remove("telegraf.zip")
	if err != nil {
		panic(err)
	}
	// delete folder
	version = strings.ReplaceAll(version, "v", "")
	err = os.RemoveAll(path + "telegraf-" + version)
	if err != nil {
		panic(err)
	}
}

func copyStuff(path string, version string) {
	// copy telegraf.exe out of telegraf folder
	version = strings.ReplaceAll(version, "v", "")
	originalPath := path + "telegraf-" + version + "\\telegraf.exe"
	newPath := path + "telegraf.exe"
	err := os.Rename(originalPath, newPath)
	if err != nil {
		panic(err)
	}
}

func unzip(basePath string) {
	command := "Expand-Archive .\\telegraf.zip -DestinationPath '" + basePath + "'"
	out, err := exec.Command("powershell", "-NoProfile", command).CombinedOutput()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", out)
}

func downloadLatestVersion(version string) {
	version = strings.ReplaceAll(version, "v", "")
	url := "https://dl.influxdata.com/telegraf/releases/telegraf-" + version + "_windows_amd64.zip"

	fmt.Printf("Downloading %s ...\n", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(resp.Body)

	fmt.Printf("Status: %s\n", resp.Status)

	// create file
	out, err := os.Create("telegraf.zip")
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(out)

	// write body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
}

func getCurrentInstalledVersion(basePath string) string {
	// run telegraf --version

	cmd := exec.Command(basePath+"telegraf.exe", "--version")
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	// parse output

	version := strings.Replace(string(out), "Telegraf ", "", -1)
	version = strings.Split(version, " ")[0]

	// return version
	return version
}

func getCurrentVersion() content {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://www.influxdata.com/versions.json", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en,en-US;q=0.7,en;q=0.3")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(resp.Body)
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var c content
	err = json.Unmarshal(bodyText, &c)
	if err != nil {
		log.Fatal(err)
	}

	return c
}
