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

### Last test cover
```
go tool cover -func=coverage.filtered.out
github.com/ibeloyar/metrics/cmd/agent/main.go:11:                               main                            0.0%
github.com/ibeloyar/metrics/cmd/server/main.go:11:                              main                            0.0%
github.com/ibeloyar/metrics/internal/agent/agent.go:20:                         pointer                         100.0%
github.com/ibeloyar/metrics/internal/agent/agent.go:24:                         Run                             0.0%
github.com/ibeloyar/metrics/internal/agent/agent.go:51:                         readRuntimeMetricsLoop          66.7%
github.com/ibeloyar/metrics/internal/agent/agent.go:69:                         readGopsutilMetricsLoop         62.5%
github.com/ibeloyar/metrics/internal/agent/agent.go:85:                         sendMetricsLoop                 41.7%
github.com/ibeloyar/metrics/internal/agent/config/config.go:26:                 Read                            90.9%
github.com/ibeloyar/metrics/internal/agent/repository/repository.go:17:         NewRepository                   100.0%
github.com/ibeloyar/metrics/internal/agent/repository/repository.go:24:         Get                             100.0%
github.com/ibeloyar/metrics/internal/agent/repository/repository.go:33:         GetAll                          100.0%
github.com/ibeloyar/metrics/internal/agent/repository/repository.go:45:         set                             100.0%
github.com/ibeloyar/metrics/internal/agent/repository/repository.go:52:         SetFromMemStats                 100.0%
github.com/ibeloyar/metrics/internal/agent/repository/repository.go:82:         GetPollCounter                  100.0%
github.com/ibeloyar/metrics/internal/agent/repository/repository.go:89:         IncrementPollCounter            100.0%
github.com/ibeloyar/metrics/internal/agent/repository/repository.go:96:         ResetPollCounter                100.0%
github.com/ibeloyar/metrics/internal/agent/repository/repository.go:103:        SetGopsutilMetrics              100.0%
github.com/ibeloyar/metrics/internal/agent/service/service.go:40:               NewService                      100.0%
github.com/ibeloyar/metrics/internal/agent/service/service.go:55:               SendMetrics                     78.3%
github.com/ibeloyar/metrics/internal/agent/service/service.go:95:               CustomBackoff                   100.0%
github.com/ibeloyar/metrics/internal/agent/service/service.go:108:              GetHashBodySHA256               100.0%
github.com/ibeloyar/metrics/internal/agent/workerpool/workerpool.go:22:         New                             100.0%
github.com/ibeloyar/metrics/internal/agent/workerpool/workerpool.go:31:         Start                           100.0%
github.com/ibeloyar/metrics/internal/agent/workerpool/workerpool.go:38:         Dispatch                        100.0%
github.com/ibeloyar/metrics/internal/agent/workerpool/workerpool.go:46:         worker                          100.0%
github.com/ibeloyar/metrics/internal/agent/workerpool/workerpool.go:58:         Shutdown                        100.0%
github.com/ibeloyar/metrics/internal/app/server/server.go:28:                   Run                             0.0%
github.com/ibeloyar/metrics/internal/app/server/server.go:69:                   buildServer                     0.0%
github.com/ibeloyar/metrics/internal/app/server/server.go:90:                   runServer                       0.0%
github.com/ibeloyar/metrics/internal/app/server/server.go:120:                  initAudit                       0.0%
github.com/ibeloyar/metrics/internal/audit/audit.go:21:                         NewSubject                      100.0%
github.com/ibeloyar/metrics/internal/audit/audit.go:25:                         Register                        100.0%
github.com/ibeloyar/metrics/internal/audit/audit.go:31:                         NotifyAll                       100.0%
github.com/ibeloyar/metrics/internal/audit/audit.go:39:                         Close                           66.7%
github.com/ibeloyar/metrics/internal/audit/fileobserver.go:14:                  NewFileAuditObserver            0.0%
github.com/ibeloyar/metrics/internal/audit/fileobserver.go:30:                  Notify                          0.0%
github.com/ibeloyar/metrics/internal/audit/httpobserver.go:17:                  NewHTTPAuditObserver            0.0%
github.com/ibeloyar/metrics/internal/audit/httpobserver.go:28:                  Notify                          0.0%
github.com/ibeloyar/metrics/internal/config/server/server.go:33:                Read                            93.3%
github.com/ibeloyar/metrics/internal/handler/metrics.go:63:                     NewMetricsHandler               100.0%
github.com/ibeloyar/metrics/internal/handler/metrics.go:75:                     GetMetricQuery                  88.2%
github.com/ibeloyar/metrics/internal/handler/metrics.go:103:                    UpdateMetricQuery               52.0%
github.com/ibeloyar/metrics/internal/handler/metrics.go:153:                    GetMetric                       59.4%
github.com/ibeloyar/metrics/internal/handler/metrics.go:203:                    UpdateMetric                    63.0%
github.com/ibeloyar/metrics/internal/handler/metrics.go:247:                    UpdateMetrics                   60.0%
github.com/ibeloyar/metrics/internal/handler/metrics.go:287:                    GetMetricsPage                  76.9%
github.com/ibeloyar/metrics/internal/handler/metrics.go:309:                    Ping                            100.0%
github.com/ibeloyar/metrics/internal/handler/metrics.go:320:                    checkHash                       83.3%
github.com/ibeloyar/metrics/internal/handler/metrics.go:335:                    GetHashBodySHA256               100.0%
github.com/ibeloyar/metrics/internal/handler/router.go:33:                      InitRoutes                      0.0%
github.com/ibeloyar/metrics/internal/logger/logger.go:11:                       New                             0.0%
github.com/ibeloyar/metrics/internal/logger/logger.go:24:                       LoggingMiddleware               0.0%
github.com/ibeloyar/metrics/internal/logger/responsewriter.go:15:               Write                           0.0%
github.com/ibeloyar/metrics/internal/logger/responsewriter.go:21:               WriteHeader                     0.0%
github.com/ibeloyar/metrics/internal/middleware/gzip/gzip.go:15:                newCompressWriter               100.0%
github.com/ibeloyar/metrics/internal/middleware/gzip/gzip.go:22:                Header                          100.0%
github.com/ibeloyar/metrics/internal/middleware/gzip/gzip.go:26:                Write                           100.0%
github.com/ibeloyar/metrics/internal/middleware/gzip/gzip.go:30:                WriteHeader                     100.0%
github.com/ibeloyar/metrics/internal/middleware/gzip/gzip.go:37:                Close                           100.0%
github.com/ibeloyar/metrics/internal/middleware/gzip/gzip.go:46:                newCompressReader               75.0%
github.com/ibeloyar/metrics/internal/middleware/gzip/gzip.go:58:                Read                            0.0%
github.com/ibeloyar/metrics/internal/middleware/gzip/gzip.go:62:                Close                           0.0%
github.com/ibeloyar/metrics/internal/middleware/gzip/gzip.go:69:                Middleware                      90.9%
github.com/ibeloyar/metrics/internal/repository/filestorage/filestorage.go:25:  New                             0.0%
github.com/ibeloyar/metrics/internal/repository/filestorage/filestorage.go:36:  Save                            0.0%
github.com/ibeloyar/metrics/internal/repository/filestorage/filestorage.go:55:  Load                            0.0%
github.com/ibeloyar/metrics/internal/repository/memstorage/memstorage.go:35:    New                             100.0%
github.com/ibeloyar/metrics/internal/repository/memstorage/memstorage.go:54:    Init                            0.0%
github.com/ibeloyar/metrics/internal/repository/memstorage/memstorage.go:74:    SetInitMetrics                  0.0%
github.com/ibeloyar/metrics/internal/repository/memstorage/memstorage.go:78:    startSavingMetrics              0.0%
github.com/ibeloyar/metrics/internal/repository/memstorage/memstorage.go:92:    Shutdown                        0.0%
github.com/ibeloyar/metrics/internal/repository/memstorage/memstorage.go:108:   GetMetric                       100.0%
github.com/ibeloyar/metrics/internal/repository/memstorage/memstorage.go:117:   GetMetrics                      100.0%
github.com/ibeloyar/metrics/internal/repository/memstorage/memstorage.go:127:   SetMetric                       62.5%
github.com/ibeloyar/metrics/internal/repository/memstorage/memstorage.go:164:   SetMetrics                      0.0%
github.com/ibeloyar/metrics/internal/repository/memstorage/memstorage.go:224:   IncrementCountMetricValue       85.7%
github.com/ibeloyar/metrics/internal/repository/memstorage/memstorage.go:264:   Ping                            0.0%
github.com/ibeloyar/metrics/internal/repository/pgstorage/pgerrors.go:23:       NewPostgresErrorClassifier      0.0%
github.com/ibeloyar/metrics/internal/repository/pgstorage/pgerrors.go:41:       Classify                        0.0%
github.com/ibeloyar/metrics/internal/repository/pgstorage/pgerrors.go:62:       classifyPgError                 0.0%
github.com/ibeloyar/metrics/internal/repository/pgstorage/pgstorage.go:43:      New                             0.0%
github.com/ibeloyar/metrics/internal/repository/pgstorage/pgstorage.go:82:      Ping                            0.0%
github.com/ibeloyar/metrics/internal/repository/pgstorage/pgstorage.go:90:      GetMetric                       0.0%
github.com/ibeloyar/metrics/internal/repository/pgstorage/pgstorage.go:111:     GetMetrics                      0.0%
github.com/ibeloyar/metrics/internal/repository/pgstorage/pgstorage.go:142:     SetMetric                       0.0%
github.com/ibeloyar/metrics/internal/repository/pgstorage/pgstorage.go:169:     SetMetrics                      0.0%
github.com/ibeloyar/metrics/internal/repository/pgstorage/pgstorage.go:208:     IncrementCountMetricValue       0.0%
github.com/ibeloyar/metrics/internal/repository/pgstorage/pgstorage.go:227:     Shutdown                        0.0%
github.com/ibeloyar/metrics/internal/repository/pgstorage/pgstorage.go:231:     executeWithRetryConnection      0.0%
github.com/ibeloyar/metrics/internal/repository/pgstorage/pgstorage.go:257:     getAttemptDelay                 0.0%
github.com/ibeloyar/metrics/internal/service/service.go:31:                     New                             100.0%
github.com/ibeloyar/metrics/internal/service/service.go:43:                     SetMetric                       0.0%
github.com/ibeloyar/metrics/internal/service/service.go:82:                     SetMetrics                      0.0%
github.com/ibeloyar/metrics/internal/service/service.go:102:                    GetMetric                       0.0%
github.com/ibeloyar/metrics/internal/service/service.go:115:                    GetMetrics                      0.0%
github.com/ibeloyar/metrics/internal/service/service.go:127:                    Ping                            0.0%
github.com/ibeloyar/metrics/internal/service/service.go:136:                    metricNames                     0.0%
github.com/ibeloyar/metrics/internal/service/service.go:144:                    parseIP                         0.0%
github.com/ibeloyar/metrics/internal/service/validate.go:10:                    IsValidMetricType               100.0%
github.com/ibeloyar/metrics/internal/service/validate.go:19:                    ValidateMetric                  0.0%
github.com/ibeloyar/metrics/internal/service/validate.go:36:                    ValidateMetrics                 0.0%
total:                                                                          (statements)                    41.9%
```