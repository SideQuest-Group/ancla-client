package cli

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(imagesCmd)
	imagesCmd.AddCommand(imagesListCmd)
	imagesCmd.AddCommand(imagesBuildCmd)
	imagesCmd.AddCommand(imagesLogCmd)
}

var imagesCmd = &cobra.Command{
	Use:     "images",
	Aliases: []string{"image", "img"},
	Short:   "Manage images",
	Long: `Manage container images for an application.

Images are built from your application source code. Each build produces a new
versioned image that can be used to create a release. Use sub-commands to list
images, trigger a new build, or view build logs.`,
	Example: "  ancla images list <app-id>\n  ancla images build <app-id>",
	GroupID: "resources",
}

var imagesListCmd = &cobra.Command{
	Use:     "list <app-id>",
	Short:   "List images for an application",
	Example: "  ancla images list abc12345",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req, _ := http.NewRequest("GET", apiURL("/images/"+args[0]), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var result struct {
			Items []struct {
				ID      string `json:"id"`
				Version int    `json:"version"`
				Built   bool   `json:"built"`
				Error   bool   `json:"error"`
				Created string `json:"created"`
			} `json:"items"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if isJSON() {
			return printJSON(result)
		}

		var rows [][]string
		for _, img := range result.Items {
			status := "building"
			if img.Error {
				status = "error"
			} else if img.Built {
				status = "built"
			}
			id := img.ID
			if len(id) > 8 {
				id = id[:8]
			}
			rows = append(rows, []string{fmt.Sprintf("v%d", img.Version), id, colorStatus(status), img.Created})
		}
		table([]string{"VERSION", "ID", "STATUS", "CREATED"}, rows)
		return nil
	},
}

var imagesBuildCmd = &cobra.Command{
	Use:     "build <app-id>",
	Short:   "Trigger an image build",
	Example: "  ancla images build abc12345",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req, _ := http.NewRequest("POST", apiURL("/images/"+args[0]+"/build"), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var result struct {
			ImageID string `json:"image_id"`
			Version int    `json:"version"`
		}
		json.Unmarshal(body, &result)
		fmt.Printf("Build triggered. Image: %s (v%d)\n", result.ImageID, result.Version)
		return nil
	},
}

var imagesLogCmd = &cobra.Command{
	Use:     "log <image-id>",
	Short:   "Show build log for an image",
	Example: "  ancla images log <image-id>",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		req, _ := http.NewRequest("GET", apiURL("/images/"+args[0]+"/log"), nil)
		body, err := doRequest(req)
		if err != nil {
			return err
		}

		var result struct {
			Status  string `json:"status"`
			Version int    `json:"version"`
			LogText string `json:"log_text"`
		}
		json.Unmarshal(body, &result)

		fmt.Printf("Image v%d — %s\n\n", result.Version, result.Status)
		if result.LogText != "" {
			fmt.Println(result.LogText)
		} else {
			fmt.Println("(no log output yet)")
		}
		return nil
	},
}
