package env

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	DbDriver          string `json:"DB_DRIVER"`
	DbSource          string `json:"DB_SOURCE"`
	ServerAddress     string `json:"SERVER_ADDRESS"`
	TokenSymmetricKey string `json:"TOKEN_SYMMETRIC_KEY"`
	TokenDuration     int    `json:"TOKEN_DURATION,string"`
}

func NewConfig() (Config, error) {
	pwd, err := currDir()
	if err != nil {
		return Config{}, err
	}
	dirSeparator := findDirSeparator(pwd)

	dirWithEnv, err := findDirWithEnv(pwd, dirSeparator)
	if err != nil {
		return Config{}, err
	}

	envStruct, err := readEnvContent(dirWithEnv)
	if err != nil {
		return Config{}, err
	}

	return envStruct, nil
}

func currDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return wd, nil
}

func findDirSeparator(dir string) string {
	var separator string

	if strings.Contains(dir, "/") {
		separator = "/"
	}
	if strings.Contains(dir, "\\") {
		separator = "\\"
	}

	return separator
}

func searchEnvFile(dir string) (string, error) {
	de, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	for _, e := range de {
		if strings.HasSuffix(e.Name(), ".env") {
			return e.Name(), nil
		}
	}

	return "", nil
}

func findDirWithEnv(dir, separator string) (string, error) {
	var (
		currDir, env   string
		separatorIndex int
		err            error
	)

	currDir = dir
	numberOfSeparators := strings.Count(dir, separator)

	for i := 0; i < numberOfSeparators; i++ {
		env, err = searchEnvFile(currDir)
		if err != nil {
			return "", err
		}
		if env != "" {
			return currDir, nil
		}
		separatorIndex = strings.LastIndex(currDir, separator)
		currDir = currDir[:separatorIndex]
	}

	err = fmt.Errorf("could not find .env file")
	return "", err
}

func readEnvContent(dwe string) (Config, error) {
	var content []byte
	envStruct := Config{}

	dirEntries, err := os.ReadDir(dwe)
	if err != nil {
		return envStruct, err
	}

	for _, entry := range dirEntries {
		if strings.Contains(entry.Name(), ".env") {
			content, err = os.ReadFile(dwe + "/" + entry.Name())
			if err != nil {
				return envStruct, err
			}
		}
	}
	ss := strings.Fields(string(content))
	transformedSlice := make([]string, len(ss))
	for i, s := range ss {
		split := strings.SplitN(s, "=", 2)
		if len(split) != 2 {
			return envStruct, err
		}
		transformedSlice[i] = fmt.Sprintf(`"%v":"%v"`,split[0], split[1])
	}
	transformedString := fmt.Sprintf(`{%v}`, strings.Join(transformedSlice, ","))

	err = json.Unmarshal([]byte(transformedString), &envStruct)
	if err != nil {
		return envStruct, err
	}
	fmt.Println(envStruct)
	return envStruct, nil
}
