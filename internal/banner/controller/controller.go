package controller

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"banner/internal/entity"
	"banner/internal/responder"
)

type BannerServicer interface {
	IncrementClick(ctx context.Context, bannerID int) error
	GetStats(ctx context.Context, bannerID int, from, to time.Time) ([]entity.MinuteStat, error)
}

type BannerController struct {
	s BannerServicer
	r responder.Responder
}

type statsItem struct {
	TS string `json:"ts"`
	V  int64  `json:"v"`
}

type statsResponse struct {
	Stats []statsItem `json:"stats"`
}

type statsRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func NewBannerController(s BannerServicer, r responder.Responder) *BannerController {
	return &BannerController{
		s: s,
		r: r,
	}
}

func (c *BannerController) ClickCounter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		c.r.ErrorBadRequest(w, errors.New("method not allowed"))
		return
	}

	bannerID, err := parseBannerID(r.URL.Path, "/counter/")
	if err != nil {
		c.r.ErrorBadRequest(w, err)
		return
	}

	if err := c.s.IncrementClick(r.Context(), bannerID); err != nil {
		c.r.ErrorInternal(w, err)
		return
	}

	c.r.OutputJSON(w, map[string]string{"status": "ok"})
}

func (c *BannerController) GetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.r.ErrorBadRequest(w, errors.New("method not allowed"))
		return
	}

	bannerId, err := parseBannerID(r.URL.Path, "/stats/")
	if err != nil {
		c.r.ErrorBadRequest(w, err)
		return
	}

	var req statsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.r.ErrorBadRequest(w, err)
		return
	}

	from, err := time.Parse("2006-01-02T15:04:05", req.From)
	if err != nil {
		c.r.ErrorBadRequest(w, errors.New("bad from time"))
		return
	}
	to, err := time.Parse("2006-01-02T15:04:05", req.To)
	if err != nil {
		c.r.ErrorBadRequest(w, errors.New("bad to time"))
		return
	}
	if !to.After(from) {
		c.r.ErrorBadRequest(w, errors.New("time to must be greater than time from"))
		return
	}

	stats, err := c.s.GetStats(r.Context(), bannerId, from, to)
	if err != nil {
		c.r.ErrorInternal(w, err)
		return
	}

	out := make([]statsItem, 0, len(stats))
	for _, s := range stats {
		var ms entity.MinuteStat = s
		out = append(out, statsItem{TS: ms.Ts.UTC().Format("2006-01-02T15:04:05"), V: ms.V})
	}

	c.r.OutputJSON(w, statsResponse{Stats: out})
}

func parseBannerID(path string, prefix string) (int, error) {
	if !strings.HasPrefix(path, prefix) {
		return 0, errors.New("bad path")
	}
	idStr := strings.TrimPrefix(path, prefix)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, errors.New("invalid banner id")
	}
	return id, nil
}
