package main

// Configuration for the diff tool

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/function61/gokit/encoding/hcl2json"
	"github.com/function61/gokit/encoding/jsonfile"
)

type item struct {
	Key string `json:"key"`
}

type config struct {
	AlwaysIgnoreDir  []item `json:"dir_always_ignore"`
	AlwaysIgnoreFile []item `json:"file_always_ignore"`
}

func loadConf() (*config, []string, []string, error) {
	hcl, err := os.Open("misc/state-diff-config.hcl")
	if err != nil {
		return nil, nil, nil, err
	}
	defer hcl.Close()

	asJson, err := hcl2json.Convert(hcl)
	if err != nil {
		return nil, nil, nil, err
	}

	conf := &config{}
	if err := jsonfile.UnmarshalDisallowUnknownFields(asJson, conf); err != nil {
		return nil, nil, nil, err
	}

	allowedChangeSubtrees := func() []string {
		items := []string{}

		for _, item := range conf.AlwaysIgnoreDir {
			items = append(items, item.Key)
		}

		return items
	}()

	allowedChangeFiles := func() []string {
		items := []string{}

		for _, item := range conf.AlwaysIgnoreFile {
			items = append(items, item.Key)
		}

		return items
	}()

	ignoreTemp, err := os.Open("misc/state-diff-ignore-temp.txt")
	if err != nil {
		return nil, nil, nil, err
	}
	defer ignoreTemp.Close()

	commentPrefix := "#"
	dirPrefix := "dir "
	filePrefix := "file "

	ignoreTempScanner := bufio.NewScanner(ignoreTemp)
	for ignoreTempScanner.Scan() {
		line := ignoreTempScanner.Text()

		switch {
		case line == "" || strings.HasPrefix(line, commentPrefix): // empty line or comment
			continue
		case strings.HasPrefix(line, dirPrefix):
			allowedChangeSubtrees = append(allowedChangeSubtrees, strings.TrimPrefix(line, dirPrefix))
		case strings.HasPrefix(line, filePrefix):
			allowedChangeFiles = append(allowedChangeFiles, strings.TrimPrefix(line, filePrefix))
		default:
			return nil, nil, nil, fmt.Errorf("invalid line: %s", line)
		}
	}
	if err := ignoreTempScanner.Err(); err != nil {
		return nil, nil, nil, err
	}

	return conf, allowedChangeSubtrees, allowedChangeFiles, nil
}

var runningSystemIdFromKernelCommandLineRe = regexp.MustCompile("root=LABEL=([^ ]+)")

func readRunningSystemId() (string, error) {
	kernelCommandLine, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		return "", fmt.Errorf("readRunningSystemId: %w", err)
	}

	matches := runningSystemIdFromKernelCommandLineRe.FindStringSubmatch(string(kernelCommandLine))
	if matches == nil {
		return "", errors.New("readRunningSystemId: failed to parse")
	}

	return matches[1], nil
}
