package dotenv

import (
	"bufio"
	"os"
	"regexp"
)

// TODO make this into a package i could reuse
func Load(filename string) (err error) {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	r := regexp.MustCompile(`([A-Z\_]{0,255})=(.{0,255})`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if matches := r.FindAllStringSubmatch(scanner.Text(), -1); matches != nil {
			os.Setenv(matches[0][1], matches[0][2])
		}
	}

	return nil
}
