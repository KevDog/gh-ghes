/*
Copyright © 2023 Kevin Stevens <kevdog@github.com>
*/
package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"

	//	"sort"

	"github.com/spf13/cobra"
)

// manifestCmd represents the manifest command

var manifestCmd = &cobra.Command{
	Use:   "manifest -v <version> -d <directory>",
	Short: "Create a csv manifest of files in a directory, sorted by dependency name",
	Long: `Create a csv manifest for GHES of based on a directory of manifests, taking the union of the dependencies.
For example:

manifest --dir /path/to/directory --version 1.0.0`,
	Run: func(cmd *cobra.Command, args []string) {	
		resultsDir := manifestDir + "/results"
		if err := os.MkdirAll(resultsDir, 0755); err != nil {
			log.Fatal(err)
		}

		lines, err := createUnion(manifestDir)
		if err != nil {
			log.Fatal(err)
		}
		if err := writeLines(resultsDir+"/union.txt", lines); err != nil {
			log.Fatal(err)
		}

		lines, err = removeDuplicates(resultsDir + "/union.txt")
		if err != nil {
			functionName := runtime.FuncForPC(reflect.ValueOf(cmd.Run).Pointer()).Name()
			log.Fatalf("%s: %v", functionName, err)
		}
		if err := writeLines(resultsDir+"/unsorted.txt", lines); err != nil {
			log.Fatal(err)
		}

		lines, err = sortLines(resultsDir + "/unsorted.txt")
		if err != nil {
			log.Fatal(err)
		}
		if err := writeLines(resultsDir+"/sorted.txt", lines); err != nil {
			log.Fatal(err)
		}
		
		lines, err = parseFile(resultsDir + "/sorted.txt")
		if err != nil {
			log.Fatal(err)
		}
		if err := writeLines(resultsDir+"/manifest.csv", lines); err != nil {
			log.Fatal(err)
		}
	},
}

var manifestDir string
var version string

func init() {
	rootCmd.AddCommand(manifestCmd)
	manifestCmd.Flags().StringVarP(&manifestDir, "dir", "d", "", "Directory of files to be processed")
	manifestCmd.MarkFlagRequired("dir")
	manifestCmd.Flags().StringVarP(&version, "version", "v", "", "Version of the manifest being processed")
	manifestCmd.MarkFlagRequired("version")
}

//createUnion iterates through the files in the directory and creates a union of the files

func createUnion(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var lines []string
	for _, file := range files {
		if !file.IsDir() {
			filePath := filepath.Join(dir, file.Name())
			fileLines, err := readLines(filePath)
			if err != nil {
				return nil, err
			}
			lines = append(lines, fileLines...)
		}
	}
	return lines, nil
}

// readLines reads a file and returns a slice of its lines
func readLines(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// writeLines writes a slice of lines to a file
func writeLines(filePath string, lines []string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}
	return writer.Flush()
}

// removeDuplicates removes duplicates from a slice of strings
func removeDuplicates(filePath string) ([]string, error) {
	lines, err := readLines(filePath)
	if err != nil {
		return nil, err
	}

	encountered := map[string]bool{}
	result := []string{}

	for _, line := range lines {
		if !encountered[line] {
			encountered[line] = true
			result = append(result, line)
		}
	}

	// Write result to file
	if err := writeLines(filePath, result); err != nil {
		return nil, err
	}

	return result, nil
}

func sortLines(filePath string) ([]string, error) {
	// Read lines from file
	lines, err := readLines(filePath)
	if err != nil {
		return nil, err
	}

	// Sort lines
	sort.Strings(lines)

	return lines, nil
}

func parseFile(filePath string) ([]string, error) {
	// Read lines from file
	lines, err := readLines(filePath)
	if err != nil {
		return nil, err
	}


	var result []string
	result = append(result, "Dependency,Version")
	for _, line := range lines {
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid line: %s", line)
		}
		result = append(result, fmt.Sprintf("%s,%s", parts[0], parts[1]))
	}

	return result, nil
}


