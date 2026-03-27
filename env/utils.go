package env

import (
	"bufio"
	"os"
	"strings"
)

func GetEnv(name string, defaults ...string) string {
	if value, exists := os.LookupEnv(name); exists {
		return value
	}

	if len(defaults) > 0 {
		return defaults[0]
	}

	return ""
}

func SetEnv(name, value string) error {
	return os.Setenv(name, value)
}

func Load(files ...string) error {
	for _, filename := range files {
		file, err := os.Open(filename)

		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()

			if strings.HasPrefix(line, "#") {
				continue
			}

			separator := strings.Index(line, "=")

			if separator == -1 {
				continue
			}

			key, value := line[:separator], line[separator+1:]
			if err = os.Setenv(key, value); err != nil {
				return err
			}
		}
	}
	return nil
}
