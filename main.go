package main

//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest

import (
	"context"
	"log"

	"terraform-provider-zerossl/provider"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func main() {
	if err := tfsdk.Serve(context.Background(), provider.New, tfsdk.ServeOpts{
		Name: "zerossl",
	}); err != nil {
		log.Fatal(err)
	}
}
