package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"time"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

// jaegerInit returns an instance of Jaeger Tracer that samples 100% of traces and logs all spans to stdout.
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

func event1(req string, ctx context.Context) (reply string){
	//1.创建子span
	span, _ := opentracing.StartSpanFromContext(ctx, "span_event1")
	defer func() {
		span.LogFields(
			log.String("event", "string-format"),
			log.String("testl", "test log"),
		)
		span.LogKV("event", "replace")

		//4.接口调用完，在tag中设置request和reply
		span.SetTag("request", req)
		span.SetTag("reply", reply)
		span.Finish()
	}()

	fmt.Println(req)
	//2.模拟处理耗时
	time.Sleep(time.Second/2)
	//3.返回reply
	reply = "event1_Reply"
	return
}

func event2(req string, ctx context.Context) (reply string){
	span, _ := opentracing.StartSpanFromContext(ctx, "span_event2")
	defer func() {
		span.SetTag("request", req)
		span.SetTag("reply", reply)
		span.Finish()
	}()

	fmt.Println(req)
	time.Sleep(time.Second)
	reply = "event2_Reply"
	return
}

func event3(rootSpan opentracing.Span, req string) (reply string){
	span := rootSpan.Tracer().StartSpan("event3", opentracing.ChildOf(rootSpan.Context()))
	defer span.Finish()

	time.Sleep(time.Second)
	reply = "event3_Reply"
	span.SetTag("request", req)
	span.SetTag("reply", reply)
	return
}

func main() {
	var host, service string
	flag.StringVar(&host, "h", "127.0.0.1:6831", "set jaeger host->ip:port.")
	flag.StringVar(&service, "s", "jaeger-demo", "set jaeger server name.")
	flag.Parse()
	fmt.Println("host    : ", host)
	fmt.Println("svc name: ", service)

	tracer, closer, err := jaegerInit(service, host)
	if err != nil {
		panic(err)
	}
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer) //StartspanFromContext创建新span时会用到

	span := tracer.StartSpan("span_root")
	ctx := opentracing.ContextWithSpan(context.Background(), span)
	r1 := event1("Hello event1", ctx)
	time.Sleep(time.Second)
	r2 := event2("Hello event2", ctx)
	fmt.Println(r1, r2)
	event3(span, "Hello event3")
	span.Finish()
}
