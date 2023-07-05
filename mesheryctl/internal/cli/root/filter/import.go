// Copyright 2023 Layer5, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package filter

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/asaskevich/govalidator"
	"github.com/layer5io/meshery/mesheryctl/internal/cli/root/config"
	"github.com/layer5io/meshery/mesheryctl/pkg/utils"
	"github.com/layer5io/meshery/server/models"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfg string
)

var importCmd = &cobra.Command{
	Use:   "import [URI]",
	Short: "Import a WASM filter",
	Long:  "Import a WASM filter from a URI (http/s) or local filesystem path",
	Example: `
// Import a filter file from local filesystem
mesheryctl exp filter import /path/to/filter.wasm

// Import a filter file from a remote URI
mesheryctl exp filter import https://example.com/myfilter.wasm

// Add WASM configuration. Config contains configuration string/filepath that is passed to the filter.
mesheryctl exp filter import /path/to/filter.wasm -config [filepath|string]
	`,
	Args: cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		mctlCfg, err := config.GetMesheryCtl(viper.GetViper())
		if err != nil {
			return errors.Wrap(err, "error processing config")
		}

		filterURL := mctlCfg.GetBaseMesheryURL() + "/api/filter"

		if len(args) == 0 {
			return errors.New(utils.FilterImportError("URI is required. Use 'mesheryctl exp filter import --help' to display usage guide.\n"))
		}

		body := models.MesheryFilterRequestBody{
			Save:       true,
			FilterData: &models.MesheryFilterPayload{},
		}

		uri := args[0]

		if validURL := govalidator.IsURL(uri); validURL {
			body.URL = uri
		} else {
			filterFile, err := os.ReadFile(uri)
			if err != nil {
				return errors.New("Unable to read file. " + err.Error())
			}

			content := string(filterFile)
			body.FilterData.FilterFile = content
		}

		if cfg != "" {
			// Check if the config is a file path or a string
			if _, err := os.Stat(cfg); err == nil {
				cfgFile, err := os.ReadFile(cfg)
				if err != nil {
					return errors.New("Unable to read config file. " + err.Error())
				}

				content := string(cfgFile)
				body.FilterData.Config = content
			} else {
				body.FilterData.Config = cfg
			}
		}

		// Convert the request body to JSON
		marshalledBody, err := json.Marshal(body)

		if err != nil {
			return err
		}

		req, err := utils.NewRequest("POST", filterURL, bytes.NewBuffer(marshalledBody))
		if err != nil {
			return err
		}

		resp, err := utils.MakeRequest(req)
		if err != nil {
			return err
		}

		if resp.StatusCode == 200 {
			utils.Log.Info("filter successfully imported")
		} else {
			return errors.Errorf("Response Status Code %d, possible Server Error", resp.StatusCode)
		}

		return nil
	},
}

func init() {
	importCmd.Flags().StringVarP(&cfg, "config", "c", "", "WASM configuration filepath/string")
}
