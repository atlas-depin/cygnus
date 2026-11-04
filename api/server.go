package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/atlas-depin/cygnus/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type API struct {
	port int64
	srv  *fiber.App
	cfg  *config.APIConfig
}

// multipartOverheadSlack covers multipart boundaries and part headers that
// make the raw request body slightly larger than the file itself.
const multipartOverheadSlack = 1 * 1024 * 1024

// NewAPI creates a new API instance using the provided API configuration.
func NewAPI(cfg *config.APIConfig) *API {
	// fasthttp treats maxBodySize <= 0 as "no limit", but fiber rewrites
	// BodyLimit == 0 back to its 4MB default, so -1 is how to disable it.
	bodyLimit := -1
	if cfg.MaxUploadSize > 0 {
		bodyLimit = int(cfg.MaxUploadSize + multipartOverheadSlack)
	}

	srv := fiber.New(fiber.Config{
		AppName:                      "Cygnus DePIN Storage Provider",
		ReadTimeout:                  cfg.Timeout(),
		WriteTimeout:                 cfg.Timeout(),
		IdleTimeout:                  cfg.Timeout(),
		DisablePreParseMultipartForm: false,
		BodyLimit:                    bodyLimit,
		StreamRequestBody:            cfg.StreamRequestBody(),
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			fmt.Println("ERROR:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	srv.Use(cors.New())

	return &API{
		port: cfg.Port,
		cfg:  cfg,
		srv:  srv,
	}
}

func (a *API) Close() error {
	if a.srv == nil {
		return fmt.Errorf("no server available")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	return a.srv.ShutdownWithContext(shutdownCtx)
}

func (a *API) Serve() {
	addr := fmt.Sprintf(":%d", a.port)
	err := a.srv.Listen(addr)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		fmt.Println("ERROR STARTING SERVER:", err)
	}
}
