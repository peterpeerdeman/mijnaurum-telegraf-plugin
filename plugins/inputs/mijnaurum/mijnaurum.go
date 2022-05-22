package mijnaurum

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const description = "Gathers Heat usage metrics from MijnAurum.nl V2 REST API"

type MijnAurum struct {
	Username   string          `toml:"username"`
	Password   string          `toml:"password"`
	Collectors []string        `toml:"collectors"`
	Log        telegraf.Logger `toml:"-"`
	tls.ClientConfig

	userId  string
	sources []Source
	url     string
	cookie  string
	client  *http.Client
}

func (ma *MijnAurum) getSourceString() string {
	result := []string{}
	for _, source := range ma.sources {
		result = append(result, source.Source)
	}
	return strings.Join(result, ",")
}

func (ma *MijnAurum) Init() error {
	if ma.Username == "" {
		return errors.New("username cannot be empty")
	}
	if ma.Password == "" {
		return errors.New("password cannot be empty")
	}

	if ma.url == "" {
		ma.url = "https://mijnaurum.nl"
	}

	availableCollectors := []string{"heat"}
	if len(ma.Collectors) == 0 {
		ma.Collectors = availableCollectors
	}

	tlsCfg, err := ma.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	ma.client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsCfg,
		},
	}

	return nil
}

func Find(a []Source, x string) int {
	for i, n := range a {
		if x == n.Type {
			return i
		}
	}
	return -1
}

func (ma *MijnAurum) Gather(acc telegraf.Accumulator) error {
	if err := ma.authenticate(); err != nil {
		acc.AddError(err)
		return err
	}

	if err := ma.getSources(); err != nil {
		acc.AddError(err)
		return err
	}
	resp, err := ma.call("actuals?sources=" + ma.getSourceString())
	if err != nil {
		acc.AddError(err)
		return err
	}

	data := ActualsResponse{}
	err = json.Unmarshal([]byte(resp), &data)
	if err != nil {
		acc.AddError(err)
		return err
	}

	for _, actual := range data.Actuals {
		if actual.Type == "heat" {
			idx := Find(ma.sources, "heat")
			if idx < 0 {
				break
			}
			tags := map[string]string{
				"source":      actual.Source,
				"source_type": actual.Type,
				"rate_unit":   ma.sources[idx].RateUnit,
				"unit":        ma.sources[idx].Unit,
				"meter_id":    ma.sources[idx].MeterID,
				"location_id": ma.sources[idx].LocationID,
			}
			fields := map[string]interface{}{
				"day_cost":    actual.ThisDay.Cost,
				"day_value":   actual.ThisDay.Value,
				"month_cost":  actual.ThisMonth.Cost,
				"month_value": actual.ThisMonth.Value,
				"week_cost":   actual.ThisWeek.Cost,
				"week_value":  actual.ThisWeek.Value,
				"year_cost":   actual.ThisYear.Cost,
				"year_value":  actual.ThisYear.Value,
			}
			acc.AddFields("mijnaurum", fields, tags)
		}
	}

	ma.cookie = ""
	return nil
}

func (ma *MijnAurum) call(endpoint string) (string, error) {
	req, err := http.NewRequest("GET", ma.url+"/user/v2/users/"+ma.userId+"/"+endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Auth-Token", ma.cookie)
	resp, err := ma.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (ma *MijnAurum) gatherActuals() error {
	resp, err := ma.call("actuals")
	if err != nil {
		return err
	}

	data := SourcesResponse{}
	err = json.Unmarshal([]byte(resp), &data)
	if err != nil {
		return err
	}

	return nil
}

func (ma *MijnAurum) getSources() error {
	resp, err := ma.call("sources")
	if err != nil {
		return err
	}

	data := SourcesResponse{}
	err = json.Unmarshal([]byte(resp), &data)
	if err != nil {
		return err
	}

	ma.sources = data.Sources

	return nil
}

func (ma *MijnAurum) authenticate() error {

	values := map[string]string{"loginName": ma.Username, "password": ma.Password}
	jsonValue, _ := json.Marshal(values)
	req, err := http.NewRequest("POST", ma.url+"/user/v2/authentication", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if err != nil {
		return err
	}

	resp, err := ma.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("statuscode from mijnaurum authentication was not 200 but %d", resp.StatusCode)
	}

	ma.cookie = resp.Header.Get("Auth-Token")

	stringdata, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	data := AuthenticationResponse{}
	err = json.Unmarshal([]byte(stringdata), &data)

	ma.userId = data.UserId

	return nil
}

// Description returns description of the plugin.
func (ma *MijnAurum) Description() string {
	return description
}

func init() {
	inputs.Add("mijnaurum", func() telegraf.Input {
		return &MijnAurum{}
	})
}
