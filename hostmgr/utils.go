package hostmgr

import (
	"encoding/json"
	"github.com/giantswarm/microerror"
	"os"
)

func saveJson(data interface{}, filepath string) error {
	marshalled, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return microerror.Mask(err)
	}
	file, err := os.Create(filepath)
	if err != nil {
		return microerror.Mask(err)
	}
	defer file.Close()

	file.Write(marshalled)
	return nil
}

func loadJson(target interface{}, filepath string) error {

	jsonFile, err := os.Open(filepath)
	if err != nil {
		return microerror.Mask(err)
	}

	defer jsonFile.Close()

	jsonDec := json.NewDecoder(jsonFile)
	err = jsonDec.Decode(target)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}
