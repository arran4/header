package cards

import (
	"bytes"
	"context"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/arran4/header"
)

func createTestPNG(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{100, 150, 200, 255}}, image.Point{}, draw.Src)
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func TestImage_FirstCoverFailsSecondSucceeds(t *testing.T) {
	pngData := createTestPNG(10, 10)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/bad.png":
			w.WriteHeader(http.StatusNotFound)
		case "/good.png":
			w.Header().Set("Content-Type", "image/png")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(pngData)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	info := &header.PageInfo{
		URL:          ts.URL,
		OGImage:      ts.URL + "/bad.png",
		TwitterImage: ts.URL + "/good.png",
	}

	opts := AssetFetchOptions{
		Timeout:       1 * time.Second,
		RetryAttempts: 1,
	}

	assets := FetchCardAssets(context.Background(), info, opts)
	if assets.CoverImage == nil {
		t.Fatal("expected CoverImage to be loaded, got nil")
	}
	if assets.CoverResult.Status != ImageStatusLoaded {
		t.Errorf("status = %q, want %q", assets.CoverResult.Status, ImageStatusLoaded)
	}
	if assets.CoverResult.SelectedURL != ts.URL+"/good.png" {
		t.Errorf("selected URL = %q, want %q", assets.CoverResult.SelectedURL, ts.URL+"/good.png")
	}
	if len(assets.CoverResult.Candidates) != 2 {
		t.Fatalf("candidates len = %d, want 2", len(assets.CoverResult.Candidates))
	}
	if assets.CoverResult.Candidates[0].Status != ImageStatusHTTPStatus {
		t.Errorf("candidate 0 status = %q, want %q", assets.CoverResult.Candidates[0].Status, ImageStatusHTTPStatus)
	}
	if assets.CoverResult.Candidates[1].Status != ImageStatusLoaded {
		t.Errorf("candidate 1 status = %q, want %q", assets.CoverResult.Candidates[1].Status, ImageStatusLoaded)
	}
}

func TestImage_OGSecureURLSucceedsAfterOGImageFailure(t *testing.T) {
	pngData := createTestPNG(10, 10)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/og-insecure.png":
			w.WriteHeader(http.StatusBadGateway)
		case "/og-secure.png":
			w.Header().Set("Content-Type", "image/png")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(pngData)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	info := &header.PageInfo{
		URL:              ts.URL,
		OGImage:          ts.URL + "/og-insecure.png",
		OGImageSecureURL: ts.URL + "/og-secure.png",
	}

	opts := AssetFetchOptions{
		Timeout:       1 * time.Second,
		RetryAttempts: 1,
	}

	assets := FetchCardAssets(context.Background(), info, opts)
	if assets.CoverImage == nil {
		t.Fatal("expected CoverImage to be loaded, got nil")
	}
	if assets.CoverResult.Status != ImageStatusLoaded {
		t.Errorf("status = %q, want %q", assets.CoverResult.Status, ImageStatusLoaded)
	}
	if assets.CoverResult.SelectedURL != ts.URL+"/og-secure.png" {
		t.Errorf("selected URL = %q, want %q", assets.CoverResult.SelectedURL, ts.URL+"/og-secure.png")
	}
}

func TestImage_TwitterFallbackSucceeds(t *testing.T) {
	pngData := createTestPNG(10, 10)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/og1.png":
			w.WriteHeader(http.StatusNotFound)
		case "/og2.png":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write([]byte("not valid image bytes"))
		case "/tw.png":
			w.Header().Set("Content-Type", "image/png")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(pngData)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	info := &header.PageInfo{
		URL:              ts.URL,
		OGImage:          ts.URL + "/og1.png",
		OGImageSecureURL: ts.URL + "/og2.png",
		TwitterImage:     ts.URL + "/tw.png",
	}

	opts := AssetFetchOptions{
		Timeout:       1 * time.Second,
		RetryAttempts: 1,
	}

	assets := FetchCardAssets(context.Background(), info, opts)
	if assets.CoverImage == nil {
		t.Fatal("expected CoverImage from Twitter fallback, got nil")
	}
	if assets.CoverResult.SelectedURL != ts.URL+"/tw.png" {
		t.Errorf("selected URL = %q, want %q", assets.CoverResult.SelectedURL, ts.URL+"/tw.png")
	}
	if len(assets.CoverResult.Candidates) != 3 {
		t.Errorf("candidates count = %d, want 3", len(assets.CoverResult.Candidates))
	}
}

func TestImage_DuplicateCandidates(t *testing.T) {
	var count int32
	pngData := createTestPNG(10, 10)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&count, 1)
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(pngData)
	}))
	defer ts.Close()

	info := &header.PageInfo{
		URL:              ts.URL,
		OGImage:          ts.URL + "/image.png",
		OGImageSecureURL: ts.URL + "/image.png", // duplicate
		TwitterImage:     ts.URL + "/image.png", // duplicate
	}

	opts := AssetFetchOptions{
		Timeout:       1 * time.Second,
		RetryAttempts: 1,
	}

	assets := FetchCardAssets(context.Background(), info, opts)
	if assets.CoverImage == nil {
		t.Fatal("expected CoverImage, got nil")
	}
	if len(assets.CoverResult.Candidates) != 1 {
		t.Errorf("expected 1 deduplicated candidate, got %d", len(assets.CoverResult.Candidates))
	}
	if atomic.LoadInt32(&count) != 1 {
		t.Errorf("expected 1 server request, got %d", atomic.LoadInt32(&count))
	}
}

func TestImage_InvalidImageBytes(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not a real png or jpeg file"))
	}))
	defer ts.Close()

	info := &header.PageInfo{
		URL:     ts.URL,
		OGImage: ts.URL + "/corrupt.png",
	}

	opts := AssetFetchOptions{
		Timeout:       1 * time.Second,
		RetryAttempts: 1,
	}

	assets := FetchCardAssets(context.Background(), info, opts)
	if assets.CoverImage != nil {
		t.Errorf("expected nil CoverImage, got %v", assets.CoverImage)
	}
	if assets.CoverResult.Status != ImageStatusDecodeError {
		t.Errorf("status = %q, want %q", assets.CoverResult.Status, ImageStatusDecodeError)
	}
}

func TestImage_OversizedImage(t *testing.T) {
	t.Run("dimensions", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "image/png")
			// Correct PNG header with 15000 x 15000 dimensions and valid CRC
			ihdrData := []byte{
				0x49, 0x48, 0x44, 0x52, // "IHDR"
				0x00, 0x00, 0x3a, 0x98, // width: 15000
				0x00, 0x00, 0x3a, 0x98, // height: 15000
				0x08, 0x02, 0x00, 0x00, 0x00, // 8-bit truecolor
			}
			crc := crc32.ChecksumIEEE(ihdrData)
			crcBytes := make([]byte, 4)
			binary.BigEndian.PutUint32(crcBytes, crc)

			var buf bytes.Buffer
			buf.Write([]byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}) // PNG signature
			buf.Write([]byte{0x00, 0x00, 0x00, 0x0d})                         // length 13
			buf.Write(ihdrData)
			buf.Write(crcBytes)
			_, _ = w.Write(buf.Bytes())
		}))
		defer ts.Close()

		info := &header.PageInfo{
			URL:     ts.URL,
			OGImage: ts.URL + "/huge.png",
		}

		opts := AssetFetchOptions{
			Timeout:       1 * time.Second,
			RetryAttempts: 1,
		}

		assets := FetchCardAssets(context.Background(), info, opts)
		if assets.CoverImage != nil {
			t.Errorf("expected nil CoverImage for oversized image, got %v", assets.CoverImage)
		}
		if assets.CoverResult.Status != ImageStatusImageTooLarge {
			t.Errorf("status = %q, want %q", assets.CoverResult.Status, ImageStatusImageTooLarge)
		}
	})

	t.Run("byte-size", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "image/png")
			// Write more than MaxImageBytes
			chunk := make([]byte, 64*1024)
			for written := 0; written <= MaxImageBytes+100; written += len(chunk) {
				_, _ = w.Write(chunk)
			}
		}))
		defer ts.Close()

		info := &header.PageInfo{
			URL:     ts.URL,
			OGImage: ts.URL + "/huge-bytes.png",
		}

		opts := AssetFetchOptions{
			Timeout:       2 * time.Second,
			RetryAttempts: 1,
		}

		assets := FetchCardAssets(context.Background(), info, opts)
		if assets.CoverImage != nil {
			t.Errorf("expected nil CoverImage for oversized byte size, got %v", assets.CoverImage)
		}
		if assets.CoverResult.Status != ImageStatusImageTooLarge {
			t.Errorf("status = %q, want %q", assets.CoverResult.Status, ImageStatusImageTooLarge)
		}
	})
}

func TestImage_NoImageMetadata(t *testing.T) {
	info := &header.PageInfo{
		URL:   "https://example.com/no-images",
		Title: "No Images Here",
	}

	assets := FetchCardAssets(context.Background(), info, DefaultAssetFetchOptions())
	if assets.CoverImage != nil {
		t.Errorf("expected nil CoverImage, got %v", assets.CoverImage)
	}
	if assets.CoverResult.Status != ImageStatusNoImageMetadata {
		t.Errorf("status = %q, want %q", assets.CoverResult.Status, ImageStatusNoImageMetadata)
	}
}

func TestImage_AllCandidatesFail(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	info := &header.PageInfo{
		URL:          ts.URL,
		OGImage:      ts.URL + "/cand1.png",
		TwitterImage: ts.URL + "/cand2.png",
	}

	opts := AssetFetchOptions{
		Timeout:       1 * time.Second,
		RetryAttempts: 1,
	}

	assets := FetchCardAssets(context.Background(), info, opts)
	if assets.CoverImage != nil {
		t.Errorf("expected nil CoverImage, got %v", assets.CoverImage)
	}
	if assets.CoverResult.Status != ImageStatusAllCandidatesFailed {
		t.Errorf("status = %q, want %q", assets.CoverResult.Status, ImageStatusAllCandidatesFailed)
	}
	if len(assets.CoverResult.Candidates) != 2 {
		t.Errorf("candidates count = %d, want 2", len(assets.CoverResult.Candidates))
	}
}

func TestImage_MissingFavicon(t *testing.T) {
	pngData := createTestPNG(10, 10)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cover.png":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(pngData)
		case "/favicon.ico":
			w.WriteHeader(http.StatusNotFound)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	info := &header.PageInfo{
		URL:     ts.URL,
		OGImage: ts.URL + "/cover.png",
		Icons: []header.Icon{
			{URL: ts.URL + "/favicon.ico"},
		},
	}

	opts := AssetFetchOptions{
		Timeout:       1 * time.Second,
		RetryAttempts: 1,
	}

	assets := FetchCardAssets(context.Background(), info, opts)
	if assets.CoverImage == nil {
		t.Fatal("expected CoverImage, got nil")
	}
	if assets.FaviconImage != nil {
		t.Errorf("expected nil FaviconImage, got %v", assets.FaviconImage)
	}
	if assets.FaviconResult.Status != ImageStatusHTTPStatus {
		t.Errorf("favicon status = %q, want %q", assets.FaviconResult.Status, ImageStatusHTTPStatus)
	}

	// Verify both basic and darkmode generators still render cleanly with placeholder favicon
	basic := &BasicCard{}
	img1, err := basic.GenerateWithAssets(info, assets)
	if err != nil {
		t.Errorf("basic.GenerateWithAssets error: %v", err)
	}
	if img1 == nil {
		t.Error("basic image is nil")
	}

	dark := &DarkmodeCard{}
	img2, err := dark.GenerateWithAssets(info, assets)
	if err != nil {
		t.Errorf("dark.GenerateWithAssets error: %v", err)
	}
	if img2 == nil {
		t.Error("dark image is nil")
	}
}

func TestImage_BasicAndDarkmodeUseSameFetchedAssets(t *testing.T) {
	var coverRequests int32
	var faviconRequests int32
	pngData := createTestPNG(20, 20)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/cover.png":
			atomic.AddInt32(&coverRequests, 1)
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(pngData)
		case "/favicon.ico":
			atomic.AddInt32(&faviconRequests, 1)
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(pngData)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	info := &header.PageInfo{
		URL:     ts.URL,
		OGImage: ts.URL + "/cover.png",
		Icons:   []header.Icon{{URL: ts.URL + "/favicon.ico"}},
	}

	opts := AssetFetchOptions{
		Timeout:       1 * time.Second,
		RetryAttempts: 1,
	}

	// Fetch assets ONCE
	assets := FetchCardAssets(context.Background(), info, opts)
	if assets.CoverImage == nil || assets.FaviconImage == nil {
		t.Fatalf("failed to fetch assets: cover=%v favicon=%v", assets.CoverImage, assets.FaviconImage)
	}

	if atomic.LoadInt32(&coverRequests) != 1 {
		t.Errorf("cover requests after fetch = %d, want 1", atomic.LoadInt32(&coverRequests))
	}
	if atomic.LoadInt32(&faviconRequests) != 1 {
		t.Errorf("favicon requests after fetch = %d, want 1", atomic.LoadInt32(&faviconRequests))
	}

	// Render Basic Card
	basic := &BasicCard{}
	img1, err := basic.GenerateWithAssets(info, assets)
	if err != nil {
		t.Fatalf("basic render error: %v", err)
	}
	if img1 == nil {
		t.Fatal("basic card is nil")
	}

	// Render Darkmode Card
	dark := &DarkmodeCard{}
	img2, err := dark.GenerateWithAssets(info, assets)
	if err != nil {
		t.Fatalf("dark render error: %v", err)
	}
	if img2 == nil {
		t.Fatal("darkmode card is nil")
	}

	// Request counts must NOT have increased!
	if atomic.LoadInt32(&coverRequests) != 1 {
		t.Errorf("cover requests after rendering both styles = %d, want 1", atomic.LoadInt32(&coverRequests))
	}
	if atomic.LoadInt32(&faviconRequests) != 1 {
		t.Errorf("favicon requests after rendering both styles = %d, want 1", atomic.LoadInt32(&faviconRequests))
	}
}

func TestFetchCardAssets_CumulativeTimeoutMultipleSlowCandidates(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Each request sleeps longer than the overall budget
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	info := &header.PageInfo{
		URL:              ts.URL,
		OGImage:          ts.URL + "/cand1.png",
		OGImageSecureURL: ts.URL + "/cand2.png",
		TwitterImage:     ts.URL + "/cand3.png",
		Icons: []header.Icon{
			{URL: ts.URL + "/fav1.png"},
		},
	}

	opts := AssetFetchOptions{
		Timeout:       100 * time.Millisecond, // Overall budget for all asset fetching combined
		RetryAttempts: 1,
	}

	start := time.Now()
	assets := FetchCardAssets(context.Background(), info, opts)
	elapsed := time.Since(start)

	// Cumulative time must not blow out to 4 * 200ms = 800ms
	if elapsed > 250*time.Millisecond {
		t.Errorf("FetchCardAssets took %v, want <= 250ms (bounded by cumulative timeout)", elapsed)
	}

	if assets.CoverImage != nil {
		t.Errorf("expected CoverImage to be nil due to timeout, got non-nil")
	}
	if assets.CoverResult.Status != ImageStatusTimeout && assets.CoverResult.Status != ImageStatusAllCandidatesFailed {
		t.Errorf("CoverResult.Status = %q, want timeout or all-candidates-failed", assets.CoverResult.Status)
	}
}

func TestFetchCardAssets_CandidateFallbackWithinBudget(t *testing.T) {
	pngData := createTestPNG(10, 10)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/slow1.png":
			// Candidate 1 hangs
			time.Sleep(300 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		case "/fast2.png":
			// Candidate 2 succeeds promptly
			w.Header().Set("Content-Type", "image/png")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(pngData)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	info := &header.PageInfo{
		URL:          ts.URL,
		OGImage:      ts.URL + "/slow1.png",
		TwitterImage: ts.URL + "/fast2.png",
	}

	opts := AssetFetchOptions{
		Timeout:       250 * time.Millisecond, // Overall budget allows candidate 1 to time out and candidate 2 to succeed
		RetryAttempts: 1,
	}

	assets := FetchCardAssets(context.Background(), info, opts)
	if assets.CoverImage == nil {
		t.Fatal("expected CoverImage to be loaded from fallback candidate 2, got nil")
	}
	if assets.CoverResult.Status != ImageStatusLoaded {
		t.Errorf("status = %q, want %q", assets.CoverResult.Status, ImageStatusLoaded)
	}
	if assets.CoverResult.SelectedURL != ts.URL+"/fast2.png" {
		t.Errorf("SelectedURL = %q, want %q", assets.CoverResult.SelectedURL, ts.URL+"/fast2.png")
	}
}
