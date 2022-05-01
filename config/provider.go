package config

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type WatchFileResponse struct {
	Key  string
	Data []byte
}

type WatchFileChan <-chan WatchFileResponse

type DataProvider interface {
	Name() string
	Read(string) ([]byte, error)
	Watch(context.Context, string) WatchFileChan
}

var providers map[string]DataProvider

func init() {
	providers = make(map[string]DataProvider)
	RegisterProvider(newFileProvider())
}

func RegisterProvider(p DataProvider) {
	providers[p.Name()] = p
}

func GetProvider(name string) DataProvider {
	return providers[name]
}

type FileProvider struct {
	ctx     context.Context
	cancel  context.CancelFunc
	watcher *fsnotify.Watcher

	mu      sync.RWMutex
	path    map[string]string                 // clean path -> path
	modTime map[string]int64                  // clean path -> seconds
	ch      map[string]chan WatchFileResponse // clean path -> chan
}

func newFileProvider() *FileProvider {
	ctx, cancel := context.WithCancel(context.Background())
	p := &FileProvider{
		ctx:     ctx,
		cancel:  cancel,
		watcher: nil,
		path:    make(map[string]string),
		modTime: make(map[string]int64),
		ch:      make(map[string]chan WatchFileResponse, 1),
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("new file watcher err: %v", err)
	}
	if watcher != nil {
		p.watcher = watcher
		go p.run()
	}
	return p
}

func (*FileProvider) Name() string {
	return "file"
}

func (p *FileProvider) Read(path string) ([]byte, error) {
	if p.watcher != nil {
		if err := p.watcher.Add(filepath.Dir(path)); err != nil {
			log.Printf("failed to watch file %v", err)
			return nil, err
		}
		cleanPath := filepath.Clean(path)
		p.mu.Lock()
		p.path[cleanPath] = path
		p.ch[cleanPath] = make(chan WatchFileResponse)
		p.mu.Unlock()
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("failed to read file %v", err)
		return nil, err
	}
	return data, nil
}

func (p *FileProvider) run() {
	if p.watcher == nil {
		return
	}
	for {
		select {
		case <-p.ctx.Done():
			return
		case event, ok := <-p.watcher.Events:
			if !ok {
				continue
			}
			if t, ok := p.isModified(event); ok {
				p.trigger(event, t)
			}
		case err, ok := <-p.watcher.Errors:
			if !ok {
				continue
			}
			log.Println(err)
		}
	}
}

func (p *FileProvider) isModified(event fsnotify.Event) (int64, bool) {
	cleanPath := filepath.Clean(event.Name)
	if event.Op&fsnotify.Write != fsnotify.Write {
		return 0, false
	}

	p.mu.RLock()
	defer p.mu.RUnlock()
	if _, ok := p.path[cleanPath]; !ok {
		return 0, false
	}

	stat, err := os.Stat(event.Name)
	if err != nil {
		log.Printf("failed to stat file %v", err)
		return 0, false
	}
	if stat.ModTime().Unix() > p.modTime[cleanPath] {
		return stat.ModTime().Unix(), true
	}
	return 0, false
}

func (p *FileProvider) trigger(event fsnotify.Event, t int64) {
	cleanPath := filepath.Clean(event.Name)
	data, err := ioutil.ReadFile(event.Name)
	if err != nil {
		log.Printf("failed to read file %v", err)
		return
	}
	p.mu.Lock()
	path := p.path[cleanPath]
	ch := p.ch[cleanPath]
	p.modTime[cleanPath] = t
	p.mu.Unlock()

	w := WatchFileResponse{
		Key:  path,
		Data: data,
	}
	go func() {
		select {
		case <-time.After(10 * time.Second):
		case ch <- w:
		}
	}()
}

func (p *FileProvider) Watch(ctx context.Context, path string) WatchFileChan {
	if p.watcher == nil {
		closeCh := make(chan WatchFileResponse)
		close(closeCh)
		return closeCh
	}
	cleanPath := filepath.Clean(path)

	p.mu.RLock()
	ch, ok := p.ch[cleanPath]
	p.mu.RUnlock()
	if !ok {
		closeCh := make(chan WatchFileResponse)
		close(closeCh)
		return closeCh
	}

	sendCh := make(chan WatchFileResponse, 1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(sendCh)
				return
			case c := <-ch:
				sendCh <- c
			}
		}
	}()

	return sendCh
}
