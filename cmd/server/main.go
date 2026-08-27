package main

import (
	"archive-deacidification/internal/application"
	"archive-deacidification/internal/httpapi"
	"archive-deacidification/internal/storage"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := parseConfig(flag.CommandLine)
	if e := validateConfig(cfg); e != nil {
		panic(e)
	}
	path := cfg.Database
	if cfg.Selfcheck {
		path = "file:selfcheck?mode=memory&cache=shared"
	}
	st, e := storage.Open(path)
	if e != nil {
		panic(e)
	}
	defer st.Close()
	app := application.New(st)
	srv := &http.Server{Addr: cfg.Addr, Handler: httpapi.JSON(httpapi.New(app).Handler())}
	if cfg.Selfcheck {
		if e := runSelfcheck(cfg.Addr, srv); e != nil {
			panic(e)
		}
		return
	}
	go func() {
		if e := srv.ListenAndServe(); e != nil && e != http.ErrServerClosed {
			panic(e)
		}
	}()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
func runSelfcheck(addr string, srv *http.Server) error {
	go srv.ListenAndServe()
	defer srv.Shutdown(context.Background())
	time.Sleep(100 * time.Millisecond)
	base := "http://" + addr
	post := func(path string, v any, headers map[string]string) (map[string]any, error) {
		raw, _ := json.Marshal(v)
		req, _ := http.NewRequest("POST", base+path, bytes.NewReader(raw))
		for k, x := range headers {
			req.Header.Set(k, x)
		}
		resp, e := http.DefaultClient.Do(req)
		if e != nil {
			return nil, e
		}
		defer resp.Body.Close()
		var out map[string]any
		json.NewDecoder(resp.Body).Decode(&out)
		if resp.StatusCode >= 300 {
			return out, fmt.Errorf("%s: %d", path, resp.StatusCode)
		}
		return out, nil
	}
	b, e := post("/v1/batches", map[string]any{"title": "自检批次", "institution": "档案馆", "targetProcess": "喷洒脱酸", "createdBy": "selfcheck", "documentItems": []any{map[string]any{"itemID": "D1", "callNumber": "A-001", "title": "档案", "material": "棉纸", "pageCount": 10}}}, nil)
	if e != nil {
		return e
	}
	batch := b["batch"].(map[string]any)
	id := batch["batchID"].(string)
	ver := int(batch["version"].(float64))
	x, e := post("/v1/batches/"+id+"/samples", map[string]any{"sampleID": "S1", "material": "棉纸", "initialPH": 5, "initialStrength": 100, "observer": "lab", "observedAt": time.Now().UTC()}, map[string]string{"If-Match": fmt.Sprint(ver)})
	if e != nil {
		return e
	}
	ver = int(x["batch"].(map[string]any)["version"].(float64))
	x, e = post("/v1/batches/"+id+"/trials", map[string]any{"trialID": "T1", "sampleID": "S1", "reagent": "碳酸镁", "concentration": 2, "temperatureCelsius": 25, "durationMinutes": 30, "initialPH": 5, "finalPH": 6.5, "strengthBefore": 100, "strengthAfter": 95}, map[string]string{"If-Match": fmt.Sprint(ver)})
	if e != nil {
		return e
	}
	ver = int(x["batch"].(map[string]any)["version"].(float64))
	x, e = post("/v1/batches/"+id+"/review", map[string]any{"approve": true, "reviewer": "负责人"}, map[string]string{"If-Match": fmt.Sprint(ver)})
	if e != nil {
		return e
	}
	ver = int(x["batch"].(map[string]any)["version"].(float64))
	x, e = post("/v1/batches/"+id+"/freeze", map[string]any{"reviewer": "负责人"}, map[string]string{"If-Match": fmt.Sprint(ver)})
	if e != nil {
		return e
	}
	credential := x["credential"].(map[string]any)
	code := credential["verificationCode"].(string)
	resp, e := http.Get(base + "/v1/batches/" + id + "/verify?code=" + code)
	if e != nil {
		return e
	}
	defer resp.Body.Close()
	var verification map[string]any
	if e = json.NewDecoder(resp.Body).Decode(&verification); e != nil || verification["valid"] != true {
		return fmt.Errorf("凭据验真失败: %v", e)
	}
	fmt.Println("selfcheck passed")
	return nil
}
