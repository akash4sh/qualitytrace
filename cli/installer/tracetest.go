package installer

import (
	"bytes"
	_ "embed"
	"html/template"

	"fmt"

	cliUI "github.com/intelops/qualitytrace/cli/ui"
)

func configureDemoApp(conf configuration, ui cliUI.UI) configuration {
	conf.set("demo.enable.pokeshop", !conf.Bool("installer.only_qualitytrace"))
	conf.set("demo.enable.otel", false)

	switch conf.String("installer") {
	case "docker-compose":
		conf.set("demo.endpoint.pokeshop.http", "http://demo-api:8081")
		conf.set("demo.endpoint.pokeshop.grpc", "demo-rpc:8082")
		conf.set("demo.endpoint.pokeshop.kafka", "stream:9092")
		conf.set("demo.endpoint.otel.frontend", "http://otel-frontend:8084")
		conf.set("demo.endpoint.otel.product_catalog", "otel-productcatalogservice:3550")
		conf.set("demo.endpoint.otel.cart", "otel-cartservice:7070")
		conf.set("demo.endpoint.otel.checkout", "otel-checkoutservice:5050")
	case "kubernetes":
		conf.set("demo.endpoint.pokeshop.http", "http://demo-pokemon-api.demo")
		conf.set("demo.endpoint.pokeshop.grpc", "demo-pokemon-api.demo:8082")
		conf.set("demo.endpoint.pokeshop.kafka", "stream.demo:9092")
		conf.set("demo.endpoint.otel.frontend", "http://otel-frontend.otel-demo:8084")
		conf.set("demo.endpoint.otel.product_catalog", "otel-productcatalogservice.otel-demo:3550")
		conf.set("demo.endpoint.otel.cart", "otel-cartservice.otel-demo:7070")
		conf.set("demo.endpoint.otel.checkout", "otel-checkoutservice.otel-demo:5050")
	}

	return conf
}

func configureQualitytrace(conf configuration, ui cliUI.UI) configuration {
	conf = configureBackend(conf, ui)
	conf.set("qualitytrace.analytics", true)

	return conf
}

func configureBackend(conf configuration, ui cliUI.UI) configuration {
	installBackend := !conf.Bool("installer.only_qualitytrace")
	conf.set("qualitytrace.backend.install", installBackend)

	if !installBackend {
		conf.set("qualitytrace.backend.type", "")
		return conf
	}

	// default values
	switch conf.String("installer") {
	case "docker-compose":
		conf.set("qualitytrace.backend.type", "otlp")
		conf.set("qualitytrace.backend.tls.insecure", true)
		conf.set("qualitytrace.backend.endpoint.collector", "http://otel-collector:4317")
		conf.set("qualitytrace.backend.endpoint", "qualitytrace:4317")
	case "kubernetes":
		conf.set("qualitytrace.backend.type", "otlp")
		conf.set("qualitytrace.backend.tls.insecure", true)
		conf.set("qualitytrace.backend.endpoint.collector", "http://otel-collector.qualitytrace:4317")
		conf.set("qualitytrace.backend.endpoint", "qualitytrace:4317")

	default:
		conf.set("qualitytrace.backend.type", "")
	}

	return conf
}

//go:embed templates/config.yaml.tpl
var configTemplate string

func getQualitytraceConfigFileContents(pHost, pUser, pPasswd string, ui cliUI.UI, config configuration) []byte {
	vals := map[string]string{
		"pHost":   pHost,
		"pUser":   pUser,
		"pPasswd": pPasswd,
	}

	tpl, err := template.New("page").Parse(configTemplate)
	if err != nil {
		ui.Panic(fmt.Errorf("cannot parse config template: %w", err))
	}

	out := &bytes.Buffer{}
	tpl.Execute(out, vals)

	return out.Bytes()
}

//go:embed templates/provision.yaml.tpl
var provisionTemplate string

func getQualitytraceProvisionFileContents(ui cliUI.UI, config configuration) []byte {
	vals := map[string]string{
		"installBackend":   fmt.Sprintf("%t", config.Bool("qualitytrace.backend.install")),
		"backendType":      config.String("qualitytrace.backend.type"),
		"backendEndpoint":  config.String("qualitytrace.backend.endpoint.query"),
		"backendInsecure":  config.String("qualitytrace.backend.tls.insecure"),
		"backendAddresses": config.String("qualitytrace.backend.addresses"),
		"backendIndex":     config.String("qualitytrace.backend.index"),
		"backendToken":     config.String("qualitytrace.backend.token"),
		"backendRealm":     config.String("qualitytrace.backend.realm"),

		"analyticsEnabled": fmt.Sprintf("%t", config.Bool("qualitytrace.analytics")),

		"enablePokeshopDemo": fmt.Sprintf("%t", config.Bool("demo.enable.pokeshop")),
		"enableOtelDemo":     fmt.Sprintf("%t", config.Bool("demo.enable.otel")),
		"pokeshopHttp":       config.String("demo.endpoint.pokeshop.http"),
		"pokeshopGrpc":       config.String("demo.endpoint.pokeshop.grpc"),
		"pokeshopKafka":      config.String("demo.endpoint.pokeshop.kafka"),
		"otelFrontend":       config.String("demo.endpoint.otel.frontend"),
		"otelProductCatalog": config.String("demo.endpoint.otel.product_catalog"),
		"otelCart":           config.String("demo.endpoint.otel.cart"),
		"otelCheckout":       config.String("demo.endpoint.otel.checkout"),
	}

	tpl, err := template.New("page").Parse(provisionTemplate)
	if err != nil {
		ui.Panic(fmt.Errorf("cannot parse config template: %w", err))
	}

	out := &bytes.Buffer{}
	tpl.Execute(out, vals)

	return out.Bytes()
}
