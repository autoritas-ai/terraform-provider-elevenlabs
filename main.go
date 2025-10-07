package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/user/terraform-provider-elevenlabs/elevenlabs"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: elevenlabs.Provider,
	})
}