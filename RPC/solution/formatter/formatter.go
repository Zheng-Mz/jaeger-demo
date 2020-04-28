package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing-contrib/go-stdlib/nethttp"

	jaeger "github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

func main() {
	tracer, closer, err := jaegerInit("formatServer", "")
	if err != nil {
		panic(err)
	}
	defer closer.Close()

	op := nethttp.MWComponentName("format")
	http.HandleFunc("/format", nethttp.MiddlewareFunc(tracer, func(w http.ResponseWriter, r *http.Request, ) {
		helloTo := r.FormValue("helloTo")
		helloStr := fmt.Sprintf("Holle, %s!", helloTo)
		w.Write([]byte(helloStr))
	}, op))

	log.Fatal(http.ListenAndServe(":8081", nil))
}

func jaegerInit(service string, host string) (opentracing.Tracer, io.Closer, error) {
	cfg := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans: true,
			LocalAgentHostPort: host,
		},
	}
	tracer, closer, err := cfg.New(service, config.Logger(jaeger.StdLogger))
	if err != nil {
		fmt.Printf("ERROR: cannot init Jaeger: %v\n", err)
	}
	return tracer, closer, err
}
