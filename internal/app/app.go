package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"go-echo-server-template/internal/auth"
	authapi "go-echo-server-template/internal/auth/api"
	authpg "go-echo-server-template/internal/auth/postgres"
	authweb "go-echo-server-template/internal/auth/web"
	"go-echo-server-template/internal/platform/authn"
	"go-echo-server-template/internal/platform/config"
	"go-echo-server-template/internal/platform/database"
	"go-echo-server-template/internal/platform/events"
	"go-echo-server-template/internal/platform/httpx"
	"go-echo-server-template/internal/platform/jobs"
	"go-echo-server-template/internal/platform/logging"
	"go-echo-server-template/internal/reports"
	reportsapi "go-echo-server-template/internal/reports/api"
	reportspg "go-echo-server-template/internal/reports/postgres"
	reportsweb "go-echo-server-template/internal/reports/web"
	"go-echo-server-template/internal/todos"
	todoapi "go-echo-server-template/internal/todos/api"
	todopg "go-echo-server-template/internal/todos/postgres"
	todoweb "go-echo-server-template/internal/todos/web"
	webassets "go-echo-server-template/web"

	"github.com/labstack/echo/v4"
)

type App struct {
	Config     *config.Config
	Echo       *echo.Echo
	DB         *sql.DB
	Queue      *jobs.Queue
	Bus        *events.Bus
	reportsSvc *reports.Service
	cancel     context.CancelFunc
}

type queueEnqueuer struct {
	q *jobs.Queue
}

func (e queueEnqueuer) Enqueue(ctx context.Context, jobType string, payload any) error {
	_, err := e.q.Enqueue(ctx, jobType, payload, time.Time{})
	return err
}

type todoSummaryAdapter struct {
	svc *todos.Service
}

func (a todoSummaryAdapter) Summary(ctx context.Context) (reports.TodoSummary, error) {
	s, err := a.svc.Summary(ctx)
	if err != nil {
		return reports.TodoSummary{}, err
	}
	return reports.TodoSummary{
		Total: s.Total, Completed: s.Completed, Outstanding: s.Outstanding,
	}, nil
}

func New(cfg *config.Config) (*App, error) {
	if err := logging.Initialize(cfg.AppEnv, cfg.Logging.Level); err != nil {
		return nil, err
	}

	db, err := database.Open(cfg.Database)
	if err != nil {
		return nil, err
	}

	bus := events.NewBus()
	todoStore := todopg.NewStore(db)
	todoSvc := todos.NewService(todoStore, bus, cfg.Auth.Disabled)

	authStore := authpg.NewStore(db)
	authSvc := auth.NewService(authStore, cfg.Auth)

	queue := jobs.NewQueue(db, cfg.Jobs.MaxAttempts, cfg.Jobs.PollInterval)
	queue.Register("todos.audit_created", func(ctx context.Context, payload json.RawMessage) error {
		logging.WithContext(ctx, "jobs").Info("todo created audit", map[string]interface{}{"payload": string(payload)})
		return nil
	})
	bus.Subscribe(todos.EventTodoCreated, func(ctx context.Context, event events.Event) error {
		_, err := queue.Enqueue(ctx, "todos.audit_created", event.Payload, time.Time{})
		return err
	})

	reportsStore := reportspg.NewStore(db)
	reportsSvc := reports.NewService(
		reportsStore,
		todoSummaryAdapter{svc: todoSvc},
		queueEnqueuer{q: queue},
		cfg.Auth.Disabled,
		cfg.Reports.ProcessingLease,
		cfg.Reports.ReconcileAfter,
	)
	queue.Register(reports.JobGenerate, func(ctx context.Context, payload json.RawMessage) error {
		var job reports.GenerateReportJob
		if err := json.Unmarshal(payload, &job); err != nil {
			logging.WithContext(ctx, "jobs").Error("invalid reports.generate payload", err, nil)
			return nil // terminal: bad payload
		}
		// Jobs are trusted internal work; attach a system principal for capability calls.
		ctx = authn.WithPrincipal(ctx, authn.Principal{
			UserID: "system", Roles: []string{"system"}, AuthMethod: "job",
		})
		return reportsSvc.Generate(ctx, job.ReportID)
	})

	e := httpx.NewEcho(cfg)
	tmplFS, err := fs.Sub(webassets.Files, "templates")
	if err != nil {
		return nil, fmt.Errorf("embed templates: %w", err)
	}
	staticFS, err := fs.Sub(webassets.Files, "static")
	if err != nil {
		return nil, fmt.Errorf("embed static: %w", err)
	}
	renderer := httpx.NewRendererFS(tmplFS)
	httpx.SetDefaultRenderer(renderer)
	e.Renderer = renderer
	httpx.MountStaticFS(e, staticFS)
	e.Use(auth.Middleware(authSvc, cfg.Auth))
	e.Use(auth.CSRFMiddleware(cfg.Auth))

	httpx.RegisterOps(e, cfg, db, func(ctx context.Context) error {
		if !cfg.Jobs.Enabled {
			return nil
		}
		return queue.Ready(ctx)
	})

	todoapi.Register(e, todoSvc)
	todoweb.Register(e, todoSvc, cfg.Auth)
	authapi.Register(e, authSvc, cfg.Auth)
	authweb.Register(e, authSvc, cfg.Auth)
	reportsapi.Register(e, reportsSvc)
	reportsweb.Register(e, reportsSvc, cfg.Auth)

	return &App{
		Config: cfg, Echo: e, DB: db, Queue: queue, Bus: bus,
		reportsSvc: reportsSvc,
	}, nil
}

func (a *App) Start(ctx context.Context) error {
	ctx, a.cancel = context.WithCancel(ctx)

	if a.Config.Jobs.Enabled && a.Config.Jobs.EmbedInServer {
		go a.Queue.Run(ctx)
	}
	if a.Config.Jobs.Enabled {
		a.startReportsReconcile(ctx)
	}

	addr := fmt.Sprintf(":%s", a.Config.HTTP.Port)
	logging.WithContext(ctx, "startup").Info(fmt.Sprintf("Server starting on %s", addr), nil)
	go func() {
		if err := a.Echo.Start(addr); err != nil && err != http.ErrServerClosed {
			logging.WithContext(context.Background(), "startup").Fatal("Server failed", err, nil)
		}
	}()
	return nil
}

func (a *App) StartWorkerOnly(ctx context.Context) {
	ctx, a.cancel = context.WithCancel(ctx)
	logging.WithContext(ctx, "worker").Info("Worker starting", nil)
	a.startReportsReconcile(ctx)
	a.Queue.Run(ctx)
}

func (a *App) startReportsReconcile(ctx context.Context) {
	if a.reportsSvc == nil {
		return
	}
	interval := a.Config.Reports.ReconcileInterval
	if interval <= 0 {
		interval = 30 * time.Second
	}
	go func() {
		log := logging.WithContext(ctx, "reports")
		if err := a.reportsSvc.ReconcilePending(ctx); err != nil {
			log.Error("initial reconcile", err, nil)
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := a.reportsSvc.ReconcilePending(ctx); err != nil {
					log.Error("reconcile pending", err, nil)
				}
			}
		}
	}()
}

func (a *App) Shutdown(ctx context.Context) error {
	if a.cancel != nil {
		a.cancel()
	}
	if a.Echo != nil {
		if err := a.Echo.Shutdown(ctx); err != nil {
			logging.WithContext(ctx, "shutdown").Error("echo shutdown", err, nil)
		}
	}
	_ = logging.Sync()
	if a.DB != nil {
		return a.DB.Close()
	}
	return nil
}
