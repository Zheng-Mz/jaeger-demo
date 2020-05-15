# 分布式链路追踪系统

## 服务安装启动
### 官方镜像下载：
https://www.jaegertracing.io/download/ 
   
### Note
本地测试的话 建议直接运行./jaeger-all-in-one \
如果是线上环境建议分开操作

## 启动jaeger容器
    docker run -d --name jaeger -e COLLECTOR_ZIPKIN_HTTP_PORT=9411 -p 5775:5775/udp -p 6831:6831/udp -p 6832:6832/udp -p 5778:5778 -p 16686:16686 -p 14268:14268 -p 9411:9411 jaegertracing/all-in-one:1.17

## jaeger-demo编译
    go build -mod=vendor

## 参考
https://github.com/yurishkuro/opentracing-tutorial/tree/master/go \
https://github.com/opentracing-contrib/go-stdlib
