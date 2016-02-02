package hostmgr

import (
	"encoding/json"
	"os"
)

func saveJson(data interface{}, filepath string) error {
	marshalled, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	file.Write(marshalled)
	return nil
}

func loadJson(target interface{}, filepath string) error {

	jsonFile, err := os.Open(filepath)
	if err != nil {
		return err
	}

	defer jsonFile.Close()

	jsonDec := json.NewDecoder(jsonFile)
	return jsonDec.Decode(target)
}
