package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/opentracing-contrib/go-stdlib/nethttp"

	jaeger "github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

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

/*The format parameter refers to one of the three standard encodings the OpenTracing API defines:
  - TextMap where span context is encoded as a collection of string key-value pairs,
  - Binary where span context is encoded as an opaque byte array,
  - HTTPHeaders, which is similar to TextMap except that the keys must be safe to be used as HTTP headers.*/
func main() {
	var service, baggage, helloTo string
	flag.StringVar(&service, "s", "hello-world", "set jaeger server name.")
	flag.StringVar(&helloTo, "t", "hello", "set jaeger server name.")
	flag.StringVar(&baggage, "b", "", "set jaeger server name.")
	flag.Parse()
	fmt.Println("svc name: ", service)

	tracer, closer, err := jaegerInit(service, "")
	if err != nil {
		panic(err)
	}
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	span := tracer.StartSpan("say-hello")
	span.SetTag("hello-to", helloTo)
	//set Baggage
	if baggage != "" {
		// after starting the span
		span.SetBaggageItem("greeting", baggage)
	}
	defer span.Finish()

	ctx := opentracing.ContextWithSpan(context.Background(), span)
	helloStr := formatString(ctx, helloTo)
	printHello(ctx, helloStr)
}

func formatString(ctx context.Context, helloTo string) string {
	v := url.Values{}
	v.Set("helloTo", helloTo)
	url := "http://localhost:8081/format?" + v.Encode()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err.Error())
	}

	op := nethttp.OperationName("formatClient")
	req1, ht := nethttp.TraceRequest(opentracing.GlobalTracer(), req.WithContext(ctx), op)
	defer ht.Finish()

	client := &http.Client{Transport: &nethttp.Transport{}}
	resp, err := client.Do(req1)
	if err != nil {
		panic(err.Error())
	}
	resp.Body.Close()

	helloStr := helloTo //string(resp)

	return helloStr
}

func printHello(ctx context.Context, helloStr string) {
	span, _ := opentracing.StartSpanFromContext(ctx, "printHello")
	defer span.Finish()

	v := url.Values{}
	v.Set("helloStr", helloStr)
	url := "http://localhost:8082/publish?" + v.Encode()
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err.Error())
	}

	ext.SpanKindRPCClient.Set(span)
	ext.HTTPUrl.Set(span, url)
	ext.HTTPMethod.Set(span, "GET")
	//注入span.Context到HTTP请求Header中
	span.Tracer().Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))

	client := &http.Client{}
	if _, err := client.Do(req); err != nil {
		panic(err.Error())
	}

	span.LogFields(
		log.String("event", "string-format"),
		log.String("value", helloStr),
	)
}
