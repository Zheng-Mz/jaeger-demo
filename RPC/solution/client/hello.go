package main

import (
	"context"
	"net/http"
	"net/url"
	"os"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
//	"github.com/opentracing/opentracing-go/log"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/yurishkuro/opentracing-tutorial/go/lib/http"
	"github.com/yurishkuro/opentracing-tutorial/go/lib/tracing"
)

/*The format parameter refers to one of the three standard encodings the OpenTracing API defines:
  - TextMap where span context is encoded as a collection of string key-value pairs,
  - Binary where span context is encoded as an opaque byte array,
  - HTTPHeaders, which is similar to TextMap except that the keys must be safe to be used as HTTP headers.*/
func main() {
	if len(os.Args) != 2 && len(os.Args) != 3 {
		panic("ERROR: Expecting one argument")
	}

	tracer, closer := tracing.Init("hello-world")
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	helloTo := os.Args[1]

	span := tracer.StartSpan("say-hello")
	span.SetTag("hello-to", helloTo)
	//set Baggage
	if len(os.Args) == 3 {
		greeting := os.Args[2]
		// after starting the span
		span.SetBaggageItem("greeting", greeting)
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
/*
	span, _ := opentracing.StartSpanFromContext(ctx, "formatString")
	defer span.Finish()
	ext.SpanKindRPCClient.Set(span)
	ext.HTTPUrl.Set(span, url)
	ext.HTTPMethod.Set(span, "GET")
	//注入span.Context到HTTP请求Header中
	span.Tracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	)
*/
	req = req.WithContext(ctx)
	req1, ht := nethttp.TraceRequest(opentracing.GlobalTracer(), req)
	defer ht.Finish()

	resp, err := xhttp.Do(req1)
	if err != nil {
		panic(err.Error())
	}

	helloStr := string(resp)
/*
	span.LogFields(
		log.String("event", "string-format"),
		log.String("value", helloStr),
	)
*/
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

	if _, err := xhttp.Do(req); err != nil {
		panic(err.Error())
	}
}
