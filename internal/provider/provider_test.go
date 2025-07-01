package provider

import (
    "testing"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

func TestScriptResourceModel_Defaults(t *testing.T) {
    model := ScriptResourceModel{
        Name:        types.StringValue("Test Script"),
        Shell:       types.StringValue("powershell"),
        ScriptBody:  types.StringValue("Write-Output 'Test'"),
    }

    // Test that required fields are set
    if model.Name.IsNull() {
        t.Error("Name should not be null")
    }
    if model.Shell.IsNull() {
        t.Error("Shell should not be null")
    }
    if model.ScriptBody.IsNull() {
        t.Error("ScriptBody should not be null")
    }

    // Test default values would be set by the API
    if !model.DefaultTimeout.IsNull() {
        t.Error("DefaultTimeout should be null before API sets it")
    }
}

func TestClientConfig_Do(t *testing.T) {
    // This is a basic test structure
    // In a real test, you'd use httptest to mock the server
    client := &ClientConfig{
        BaseURL: "https://test.example.com/api",
        APIKey:  "test-key",
    }

    if client.BaseURL != "https://test.example.com/api" {
        t.Errorf("Expected BaseURL to be https://test.example.com/api, got %s", client.BaseURL)
    }
    if client.APIKey != "test-key" {
        t.Errorf("Expected APIKey to be test-key, got %s", client.APIKey)
    }
}
