# go-musthave-metrics-tpl

Metrics and alerting collection service

### Last pprof diff
```
File: ___2go_run_LOCAL
Build ID: cb99a7f2106c24b99d7503bb5697bb1d9531a4b7
Type: inuse_space
Time: 2026-02-18 09:48:57 MSK
Showing nodes accounting for -3589.28kB, 58.33% of 6153.32kB total
Dropped 2 nodes (cum <= 30.77kB)
      flat  flat%   sum%        cum   cum%
   -1539kB 25.01% 25.01%    -1539kB 25.01%  runtime.allocm
    -514kB  8.35% 33.36%     -514kB  8.35%  bufio.NewReaderSize (inline)
 -512.17kB  8.32% 41.69%  -512.17kB  8.32%  net/textproto.MIMEHeader.Set (inline)
 -512.12kB  8.32% 50.01%  -512.12kB  8.32%  github.com/go-chi/chi/v5.NewRouteContext
  512.05kB  8.32% 41.69%   512.05kB  8.32%  sync.runtime_notifyListWait
 -512.02kB  8.32% 50.01%  -512.02kB  8.32%  text/template/parse.(*Tree).newCommand (inline)
 -512.01kB  8.32% 58.33%  -512.01kB  8.32%  text/template/parse.(*Tree).newText (inline)
         0     0% 58.33%     -514kB  8.35%  bufio.NewReader (inline)
         0     0% 58.33% -2048.32kB 33.29%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 58.33% -1536.20kB 24.97%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 58.33% -1536.20kB 24.97%  github.com/go-chi/chi/v5/middleware.Recoverer.func1
         0     0% 58.33% -1536.20kB 24.97%  github.com/ibeloyar/metrics/internal/app/server.buildServer.LoggingMiddleware.func1.1
         0     0% 58.33%  -512.12kB  8.32%  github.com/ibeloyar/metrics/internal/app/server.buildServer.NewRouter.NewMux.func2
         0     0% 58.33% -1536.20kB 24.97%  github.com/ibeloyar/metrics/internal/handler.(*MetricsHandler).GetMetricsPage
         0     0% 58.33% -1536.20kB 24.97%  github.com/ibeloyar/metrics/internal/middleware/gzip.Middleware.func1
         0     0% 58.33% -1024.03kB 16.64%  html/template.(*Template).Parse
         0     0% 58.33%     -513kB  8.34%  net/http.(*conn).readRequest
         0     0% 58.33% -2049.27kB 33.30%  net/http.(*conn).serve
         0     0% 58.33%   512.05kB  8.32%  net/http.(*connReader).abortPendingRead
         0     0% 58.33%   512.05kB  8.32%  net/http.(*response).finishRequest
         0     0% 58.33% -1536.20kB 24.97%  net/http.HandlerFunc.ServeHTTP
         0     0% 58.33%  -512.17kB  8.32%  net/http.Header.Set (inline)
         0     0% 58.33%     -514kB  8.35%  net/http.newBufioReader
         0     0% 58.33% -2048.32kB 33.29%  net/http.serverHandler.ServeHTTP
         0     0% 58.33%     -513kB  8.34%  runtime.findRunnable
         0     0% 58.33%     -513kB  8.34%  runtime.injectglist
         0     0% 58.33%     -513kB  8.34%  runtime.injectglist.func1
         0     0% 58.33%     -513kB  8.34%  runtime.mcall
         0     0% 58.33%    -1026kB 16.67%  runtime.mstart
         0     0% 58.33%    -1026kB 16.67%  runtime.mstart0
         0     0% 58.33%    -1026kB 16.67%  runtime.mstart1
         0     0% 58.33%    -1539kB 25.01%  runtime.newm
         0     0% 58.33%     -513kB  8.34%  runtime.park_m
         0     0% 58.33%    -1026kB 16.67%  runtime.resetspinning
         0     0% 58.33%    -1539kB 25.01%  runtime.schedule
         0     0% 58.33%    -1539kB 25.01%  runtime.startm
         0     0% 58.33%    -1026kB 16.67%  runtime.wakep
         0     0% 58.33%   512.05kB  8.32%  sync.(*Cond).Wait
         0     0% 58.33%  -512.12kB  8.32%  sync.(*Pool).Get
         0     0% 58.33% -1024.03kB 16.64%  text/template.(*Template).Parse
         0     0% 58.33% -1024.03kB 16.64%  text/template/parse.(*Tree).Parse
         0     0% 58.33% -1024.03kB 16.64%  text/template/parse.(*Tree).action
         0     0% 58.33%  -512.02kB  8.32%  text/template/parse.(*Tree).command
         0     0% 58.33% -1024.03kB 16.64%  text/template/parse.(*Tree).itemList
         0     0% 58.33% -1024.03kB 16.64%  text/template/parse.(*Tree).parse
         0     0% 58.33% -1024.03kB 16.64%  text/template/parse.(*Tree).parseControl
         0     0% 58.33%  -512.02kB  8.32%  text/template/parse.(*Tree).pipeline
         0     0% 58.33% -1024.03kB 16.64%  text/template/parse.(*Tree).rangeControl
         0     0% 58.33% -1024.03kB 16.64%  text/template/parse.(*Tree).textOrAction
         0     0% 58.33% -1024.03kB 16.64%  text/template/parse.Parse
```