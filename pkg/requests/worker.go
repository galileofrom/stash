package requests

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Importer copies (or hardlinks) a finished download from the download client's
// output path into the stash library so that a normal stash scan picks it up.
type Importer interface {
	Import(srcPath, libraryPath string) (destPath string, err error)
}

// LibraryScanner triggers stash to scan its library after import. The
// background worker calls this once per successful import; concrete impls
// will call manager.GetInstance().RunSingleTask or the scan service.
type LibraryScanner interface {
	ScanLibrary(ctx context.Context) error
}

// WorkerConfig groups dependencies the worker needs at construction time.
type WorkerConfig struct {
	Repo        Repository
	Prowlarr    *ProwlarrClient
	QBittorrent *QBittorrentClient
	Importer    Importer
	Scanner     LibraryScanner
	LibraryPath string
	TickEvery   time.Duration // default 30s
	Logger      Logger
}

// Logger is intentionally minimal so the package does not couple to any
// specific logging library; stash logger satisfies it.
type Logger interface {
	Infof(fmt string, args ...any)
	Errorf(fmt string, args ...any)
}

type Worker struct {
	cfg WorkerConfig
}

func NewWorker(cfg WorkerConfig) *Worker {
	if cfg.TickEvery == 0 {
		cfg.TickEvery = 30 * time.Second
	}
	return &Worker{cfg: cfg}
}

// Run blocks until ctx is cancelled, ticking on cfg.TickEvery and processing
// any in-flight downloads on each tick.
func (w *Worker) Run(ctx context.Context) {
	t := time.NewTicker(w.cfg.TickEvery)
	defer t.Stop()
	w.tick(ctx) // run once immediately
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			w.tick(ctx)
		}
	}
}

func (w *Worker) tick(ctx context.Context) {
	queued, err := w.cfg.Repo.ListDownloads(ctx, DownloadQueued)
	if err != nil {
		w.logErr("listing queued downloads", err)
		return
	}
	inFlight, err := w.cfg.Repo.ListDownloads(ctx, DownloadDownloading)
	if err != nil {
		w.logErr("listing downloading downloads", err)
		return
	}

	for _, d := range append(queued, inFlight...) {
		if err := w.process(ctx, d); err != nil {
			w.logErr(fmt.Sprintf("processing download %d", d.ID), err)
		}
	}
}

func (w *Worker) process(ctx context.Context, d *Download) error {
	if w.cfg.QBittorrent == nil {
		return errors.New("qbittorrent client not configured")
	}
	if d.DownloadID == "" {
		// Nothing to poll yet; Prowlarr's grab response did not give us the hash.
		return nil
	}

	t, err := w.cfg.QBittorrent.TorrentByHash(ctx, d.DownloadID)
	if err != nil {
		return err
	}
	if t == nil {
		return nil
	}

	d.Progress = t.Progress
	d.ETASeconds = t.ETA
	d.OutputPath = t.ContentPath
	if d.Status == DownloadQueued && t.Progress > 0 {
		d.Status = DownloadDownloading
		now := time.Now()
		d.StartedAt = &now
	}

	if t.IsComplete() {
		dest, err := w.importDownload(d)
		if err != nil {
			d.Status = DownloadFailed
			d.Error = err.Error()
			_ = w.cfg.Repo.UpdateDownload(ctx, d)
			return err
		}
		now := time.Now()
		d.CompletedAt = &now
		d.ImportedAt = &now
		d.OutputPath = dest
		d.Status = DownloadImported
		if w.cfg.Scanner != nil {
			if err := w.cfg.Scanner.ScanLibrary(ctx); err != nil {
				w.logErr("triggering library scan", err)
			}
		}
	}

	return w.cfg.Repo.UpdateDownload(ctx, d)
}

func (w *Worker) importDownload(d *Download) (string, error) {
	if w.cfg.Importer != nil {
		return w.cfg.Importer.Import(d.OutputPath, w.cfg.LibraryPath)
	}
	return defaultImport(d.OutputPath, w.cfg.LibraryPath)
}

// defaultImport hardlinks src into libraryPath, falling back to a copy if
// the two paths sit on different filesystems. Returns the destination path.
func defaultImport(src, libraryPath string) (string, error) {
	if src == "" {
		return "", errors.New("download has no output path")
	}
	info, err := os.Stat(src)
	if err != nil {
		return "", fmt.Errorf("stat %q: %w", src, err)
	}
	if info.IsDir() {
		// stash will scan the directory recursively; hardlink each file.
		dest := filepath.Join(libraryPath, filepath.Base(src))
		if err := hardlinkTree(src, dest); err != nil {
			return "", err
		}
		return dest, nil
	}
	dest := filepath.Join(libraryPath, filepath.Base(src))
	if err := os.Link(src, dest); err == nil {
		return dest, nil
	}
	// fall through to copy on EXDEV / unsupported FS
	return dest, copyFile(src, dest)
}

func hardlinkTree(srcDir, destDir string) error {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destDir, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		if err := os.Link(path, target); err == nil {
			return nil
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	buf := make([]byte, 1<<20)
	for {
		n, rerr := in.Read(buf)
		if n > 0 {
			if _, werr := out.Write(buf[:n]); werr != nil {
				return werr
			}
		}
		if rerr != nil {
			if rerr.Error() == "EOF" {
				return nil
			}
			return nil
		}
	}
}

func (w *Worker) logErr(msg string, err error) {
	if w.cfg.Logger != nil {
		w.cfg.Logger.Errorf("requests worker: %s: %v", msg, err)
	}
}
