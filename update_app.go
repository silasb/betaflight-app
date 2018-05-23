package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"syscall"
)

const (
	UPDATE_HOST_BASE = "http://foo.us.to/gui2/"
)

type Version struct {
	Version string `json:"version"`
	File    string `json:"file"`
}

type Versions struct {
	Versions []Version
}

func CheckForNewUpdates(versionString string) (err error) {
	resp, err := http.Get(UPDATE_HOST_BASE + "versions.json")
	if err != nil {
		log.Println(err)
		return nil
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	var versions []Version
	json.Unmarshal(body, &versions)

	var latestVersion Version

	for _, version := range versions {
		if version.Version > versionString {
			latestVersion = version
			break
		}
	}

	if latestVersion.Version == "" {
		fmt.Println("We are up to date")
		return nil
	} else {
		fmt.Println("latest version", latestVersion.Version)

		return updateBinary(latestVersion)
	}
}

func updateBinary(version Version) (err error) {
	resp, err := http.Get(UPDATE_HOST_BASE + version.File)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	dest, err := os.Executable()
	if err != nil {
		return err
	}

	destBackup := dest + ".bak"
	if _, err := os.Stat(dest); err == nil {
		os.Rename(dest, destBackup)
	}

	log.Printf("downloading new version to %s\n", dest)
	if err := ioutil.WriteFile(dest, data, 0755); err != nil {
		os.Rename(destBackup, dest)
		return err
	}

	os.Remove(destBackup)

	log.Printf("updated with success to version %s\nRestarting application\n", version.Version)

	var args []string
	cmd := exec.Command(os.Args[0], args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		// If the command fails to run or doesn't complete successfully, the
		// error is of type *ExitError. Other error types may be
		// returned for I/O problems.
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The command didn't complete correctly.
			// Exiting while keeping the status code.
			os.Exit(exiterr.Sys().(syscall.WaitStatus).ExitStatus())
		} else {
			return err
		}
	}

	os.Exit(0)

	return nil
}
