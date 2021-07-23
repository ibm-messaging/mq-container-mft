/*
© Copyright IBM Corporation 2020, 2021

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"fmt"
)

// resolveLicenseFile returns the file name of the MQ MFT license file, taking into
// account the language set by the LANG environment variable
func resolveLicenseFile() string {
	lang, ok := os.LookupEnv("LANG")
	if !ok {
		return "Lic_en.txt"
	}
	switch {
	case strings.HasPrefix(lang, "zh_TW"):
		return "Lic_zh_tw.txt"
	case strings.HasPrefix(lang, "zh"):
		return "Lic_zh.txt"
	// Differentiate Czech (cs) and Kashubian (csb)
	case strings.HasPrefix(lang, "cs") && !strings.HasPrefix(lang, "csb"):
		return "Lic_cs.txt"
	case strings.HasPrefix(lang, "fr"):
		return "Lic_fr.txt"
	case strings.HasPrefix(lang, "de"):
		return "Lic_de.txt"
	case strings.HasPrefix(lang, "el"):
		return "Lic_el.txt"
	case strings.HasPrefix(lang, "id"):
		return "Lic_id.txt"
	case strings.HasPrefix(lang, "it"):
		return "Lic_it.txt"
	case strings.HasPrefix(lang, "ja"):
		return "Lic_ja.txt"
	// Differentiate Korean (ko) from Konkani (kok)
	case strings.HasPrefix(lang, "ko") && !strings.HasPrefix(lang, "kok"):
		return "Lic_ko.txt"
	case strings.HasPrefix(lang, "lt"):
		return "Lic_lt.txt"
	case strings.HasPrefix(lang, "pl"):
		return "Lic_pl.txt"
	case strings.HasPrefix(lang, "pt"):
		return "Lic_pt.txt"
	case strings.HasPrefix(lang, "ru"):
		return "Lic_ru.txt"
	case strings.HasPrefix(lang, "sl"):
		return "Lic_sl.txt"
	case strings.HasPrefix(lang, "es"):
		return "Lic_es.txt"
	case strings.HasPrefix(lang, "tr"):
		return "Lic_tr.txt"
	}
	return "Lic_en.txt"
}

func checkLicense() (bool, error) {
	lic, ok := os.LookupEnv("LICENSE")
	switch {
	case ok && lic == "accept":
		return true, nil
	case ok && lic == "view":
	    // Display MFT Redistributable package license file
		file := filepath.Join("/opt/mqm/mqft/licences/", resolveLicenseFile())
		// #nosec G304
		buf, err := ioutil.ReadFile(file)
		if err != nil {
			fmt.Println(err)
			return false, err
		}
		fmt.Println(string(buf))
		return false, nil
	}
	fmt.Println("Error: Set environment variable LICENSE=accept to indicate acceptance of license terms and conditions.")
	fmt.Println("License agreements and information can be viewed by setting the environment variable LICENSE=view.  You can also set the LANG environment variable to view the license in a different language.")
	return false, errors.New("Set environment variable LICENSE=accept to indicate acceptance of license terms and conditions")
}