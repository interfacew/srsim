//go:generate go run generate.go types.go

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/simimpact/srsim/pkg/engine/equip/lightcone"
	"github.com/simimpact/srsim/pkg/model"
)

// data for templates
type dataTmpl struct {
	Key           string
	KeyLower      string
	Rarity        string
	Path          model.Path
	PromotionData []lightcone.PromotionData
}

var keyRegex = regexp.MustCompile(`\W+`) // for removing spaces
var rarityRegex = regexp.MustCompile(`CombatPowerLightconeRarity(\d+)`)

func main() {
	dmPath := os.Getenv("DM_PATH")
	if dmPath == "" {
		fmt.Println("Please provide the path to StarRailData (environment variable DM_PATH).")
		return
	}

	var cones map[string]EquipmentConfig
	var promotions map[string]PromotionConfig
	var textMap map[string]string

	err := OpenConfig(&cones, dmPath, "ExcelOutput", "EquipmentConfig.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = OpenConfig(&promotions, dmPath, "ExcelOutput", "EquipmentPromotionConfig.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = OpenConfig(&textMap, dmPath, "TextMap", "TextMapEN.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	for key, value := range cones {
		if err != nil {
			fmt.Println(err)
			return
		}
		coneName := GetName(textMap, value.EquipmentName.Hash)
		if coneName == "" {
			continue
		}
		ProcessLightCone(coneName, value, promotions[key])
	}
}

func OpenConfig(result interface{}, path ...string) error {
	jsonFile := filepath.Join(path...)
	file, err := os.Open(jsonFile)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return err
	}
	return nil
}

func GetName(textMap map[string]string, hash int) string {
	return textMap[strconv.Itoa(hash)]
}

func ProcessLightCone(name string, cone EquipmentConfig, promotions PromotionConfig) {
	data := dataTmpl{}
	data.Key = keyRegex.ReplaceAllString(name, "")
	data.KeyLower = strings.ToLower(data.Key)
	data.Rarity = rarityRegex.FindStringSubmatch(cone.Rarity)[1]
	data.Path = cone.GetPath()

	data.PromotionData = make([]lightcone.PromotionData, len(promotions))
	for i := 0; i < len(promotions); i++ {
		val, ok := promotions[strconv.Itoa(i)]
		if !ok {
			break
		}
		data.PromotionData[i] = lightcone.PromotionData{
			MaxLevel: val.MaxLevel,
			ATKBase:  val.BaseAttack.Value,
			ATKAdd:   val.BaseAttackAdd.Value,
			DEFBase:  val.BaseDefence.Value,
			DEFAdd:   val.BaseDefenceAdd.Value,
			HPBase:   val.BaseHP.Value,
			HPAdd:    val.BaseHPAdd.Value,
		}
	}

	// save .go files
	path := filepath.Join(".", "result", strings.ToLower(data.Path.String()), data.KeyLower)
	os.MkdirAll(path, os.ModePerm)

	fcone, err := os.Create(filepath.Join(path, data.KeyLower+".go"))
	if err != nil {
		log.Fatal(err)
	}
	defer fcone.Close()
	tchar, err := template.New("outchar").Parse(tmplCone)
	if err != nil {
		log.Fatal(err)
	}
	if err := tchar.Execute(fcone, data); err != nil {
		log.Fatal(err)
	}

	fdata, err := os.Create(filepath.Join(path, "data.go"))
	if err != nil {
		log.Fatal(err)
	}
	defer fdata.Close()
	tdata, err := template.New("outdata").Parse(tmplData)
	if err != nil {
		log.Fatal(err)
	}
	if err := tdata.Execute(fdata, data); err != nil {
		log.Fatal(err)
	}
}

var tmplCone = `package {{.KeyLower}}

import (
	"github.com/simimpact/srsim/pkg/engine"
	"github.com/simimpact/srsim/pkg/engine/equip/lightcone"
	"github.com/simimpact/srsim/pkg/engine/info"
	"github.com/simimpact/srsim/pkg/key"
	"github.com/simimpact/srsim/pkg/model"
)

func init() {
	lightcone.Register(key.{{.Key}}, lightcone.Config{
		CreatePassive: Create,
		Rarity:        {{.Rarity}},
		Path:          model.Path_{{.Path}},
		Promotions:    promotions,
	})
}

func Create(engine engine.Engine, owner key.TargetID, lc info.LightCone) {

}
`

var tmplData = `package {{.KeyLower}}

import "github.com/simimpact/srsim/pkg/engine/equip/lightcone"

var promotions = []lightcone.PromotionData{
	{{- range $e := $.PromotionData}}
	{
		MaxLevel: {{$e.MaxLevel}},
		HPBase:   {{$e.HPBase}},
		HPAdd:    {{$e.HPAdd}},
		ATKBase:  {{$e.ATKBase}},
		ATKAdd:   {{$e.ATKAdd}},
		DEFBase:  {{$e.DEFBase}},
		DEFAdd:   {{$e.DEFAdd}},
	},
	{{- end}}
}`
