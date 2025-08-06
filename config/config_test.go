package config

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) string // Returns the path to the config file, if any
		want     *Config
		wantErr  bool
	}{
		{
			name: "Defaults and API Key from Env",
			setup: func(t *testing.T) string {
				t.Setenv("OPENAI_API_KEY", "env-key")
				t.Setenv("OPENAI_BASE_URL", "https://api.example.com/v1")
				return ""
			},
			want: &Config{
				OpenAIAPIKey:  "env-key",
				OpenAIBaseURL: "https://api.example.com/v1",
				ModelName:     "gpt-4-turbo",
			},
		},
		{
			name: "From File",
			setup: func(t *testing.T) string {
				content := `{"openai_api_key": "file-key", "openai_base_url": "http://localhost:8080", "model_name": "test-model"}`
				tmpfile, err := os.CreateTemp("", "config-*.json")
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { os.Remove(tmpfile.Name()) })
				if _, err := tmpfile.Write([]byte(content)); err != nil {
					t.Fatal(err)
				}
				if err := tmpfile.Close(); err != nil {
					t.Fatal(err)
				}
				return tmpfile.Name()
			},
			want: &Config{
				OpenAIAPIKey:  "file-key",
				OpenAIBaseURL: "http://localhost:8080",
				ModelName:     "test-model",
			},
		},
		{
			name: "Env Overrides File",
			setup: func(t *testing.T) string {
				content := `{"openai_api_key": "file-key", "model_name": "file-model"}`
				tmpfile, err := os.CreateTemp("", "config-*.json")
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { os.Remove(tmpfile.Name()) })
				tmpfile.Write([]byte(content))
				tmpfile.Close()
				t.Setenv("OPENAI_API_KEY", "env-key-override")
				t.Setenv("MODEL_NAME", "env-model-override")
				return tmpfile.Name()
			},
			want: &Config{
				OpenAIAPIKey:  "env-key-override",
				OpenAIBaseURL: "", // Not set in file or env
				ModelName:     "env-model-override",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment for the test case
			originalEnv := os.Environ()
			os.Clearenv()
			t.Cleanup(func() {
				os.Clearenv()
				for _, v := range originalEnv {
					parts := strings.SplitN(v, "=", 2)
					if len(parts) == 2 {
						os.Setenv(parts[0], parts[1])
					}
				}
			})

			path := tt.setup(t)

			got, err := Load(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Load() = %+v, want %+v", got, tt.want)
			}
		})
	}
}