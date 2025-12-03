package search

import (
    "context"
    "path"
    "runtime"
    "strings"
    "sync"
    "sync/atomic"
    "time"

    "github.com/OpenListTeam/OpenList/v4/internal/conf"
    "github.com/OpenListTeam/OpenList/v4/internal/errs"
    "github.com/OpenListTeam/OpenList/v4/internal/fs"
    "github.com/OpenListTeam/OpenList/v4/internal/model"
    "github.com/OpenListTeam/OpenList/v4/internal/op"
    "github.com/OpenListTeam/OpenList/v4/internal/search/searcher"
    "github.com/OpenListTeam/OpenList/v4/internal/setting"
    "github.com/OpenListTeam/OpenList/v4/pkg/mq"
    "github.com/OpenListTeam/OpenList/v4/pkg/utils"
    mapset "github.com/deckarep/golang-set/v2"
    log "github.com/sirupsen/logrus"
)

var (
	Quit = atomic.Pointer[chan struct{}]{}
)

func Running() bool {
	return Quit.Load() != nil
}

func BuildIndex(ctx context.Context, indexPaths, ignorePaths []string, maxDepth int, count bool) error {
	var (
		err      error
		objCount uint64 = 0
		fi       model.Obj
	)
	log.Infof("build index for: %+v", indexPaths)
	log.Infof("ignore paths: %+v", ignorePaths)
	quit := make(chan struct{}, 1)
	if !Quit.CompareAndSwap(nil, &quit) {
		// other goroutine is running
		return errs.BuildIndexIsRunning
	}
	var (
		indexMQ = mq.NewInMemoryMQ[ObjWithParent]()
		running = atomic.Bool{} // current goroutine running
		wg      = &sync.WaitGroup{}
	)
	running.Store(true)
	wg.Add(1)
	go func() {
		ticker := time.NewTicker(time.Second)
		defer func() {
			Quit.Store(nil)
			wg.Done()
			// notify walk to exit when StopIndex api called
			running.Store(false)
			ticker.Stop()
		}()
		tickCount := 0
		for {
			select {
			case <-ticker.C:
				tickCount += 1
                if indexMQ.Len() < 10000 && tickCount != 3 {
                    continue
                } else if tickCount >= 3 {
                    tickCount = 0
                }
                log.Infof("index obj count: %d", objCount)
                indexMQ.ConsumeAll(func(messages []mq.Message[ObjWithParent]) {
					if len(messages) != 0 {
						log.Debugf("current index: %s", messages[len(messages)-1].Content.Parent)
					}
					if err = BatchIndex(ctx, utils.MustSliceConvert(messages,
						func(src mq.Message[ObjWithParent]) ObjWithParent {
							return src.Content
						})); err != nil {
						log.Errorf("build index in batch error: %+v", err)
					} else {
						objCount = objCount + uint64(len(messages))
					}
					if count {
						WriteProgress(&model.IndexProgress{
							ObjCount:     objCount,
							IsDone:       false,
							LastDoneTime: nil,
						})
					}
				})

			case <-quit:
				log.Debugf("build index for %+v received quit", indexPaths)
				eMsg := ""
				now := time.Now()
				originErr := err
				indexMQ.ConsumeAll(func(messages []mq.Message[ObjWithParent]) {
					if err = BatchIndex(ctx, utils.MustSliceConvert(messages,
						func(src mq.Message[ObjWithParent]) ObjWithParent {
							return src.Content
						})); err != nil {
						log.Errorf("build index in batch error: %+v", err)
					} else {
						objCount = objCount + uint64(len(messages))
					}
					if originErr != nil {
						log.Errorf("build index error: %+v", originErr)
						eMsg = originErr.Error()
					} else {
						log.Infof("success build index, count: %d", objCount)
					}
					if count {
						WriteProgress(&model.IndexProgress{
							ObjCount:     objCount,
							IsDone:       true,
							LastDoneTime: &now,
							Error:        eMsg,
						})
					}
				})
				log.Debugf("build index for %+v quit success", indexPaths)
				return
			}
		}
	}()
	defer func() {
		if !running.Load() || Quit.Load() != &quit {
			log.Debugf("build index for %+v stopped by StopIndex", indexPaths)
			return
		}
		select {
		// avoid goroutine leak
		case quit <- struct{}{}:
		default:
		}
		wg.Wait()
	}()
	if count {
		WriteProgress(&model.IndexProgress{
			ObjCount: 0,
			IsDone:   false,
		})
	}
    for _, indexPath := range indexPaths {
        fi, err = fs.Get(ctx, indexPath, &fs.GetArgs{})
        if err != nil {
            return err
        }
        type item struct{ p string; info model.Obj; depth int }
        workCh := make(chan item, 1024)
        tasks := &sync.WaitGroup{}
        workers := runtime.NumCPU() * 8
        wgw := &sync.WaitGroup{}
        for i := 0; i < workers; i++ {
            wgw.Add(1)
            go func() {
                defer wgw.Done()
                for it := range workCh {
                    if !running.Load() {
                        tasks.Done()
                        continue
                    }
                    skip := false
                    for _, avoidPath := range ignorePaths {
                        if strings.HasPrefix(it.p, avoidPath) {
                            skip = true
                            break
                        }
                    }
                    if skip {
                        tasks.Done()
                        continue
                    }
                    if storage, _, err := op.GetStorageAndActualPath(it.p); err == nil {
                        if storage.GetStorage().DisableIndex {
                            tasks.Done()
                            continue
                        }
                    }
                    if it.p != "/" {
                        indexMQ.Publish(mq.Message[ObjWithParent]{
                            Content: ObjWithParent{Obj: it.info, Parent: path.Dir(it.p)},
                        })
                    }
                    if !it.info.IsDir() || it.depth == 0 {
                        tasks.Done()
                        continue
                    }
                    meta, _ := op.GetNearestMeta(it.p)
                    objs, err := fs.List(context.WithValue(ctx, conf.MetaKey, meta), it.p, &fs.ListArgs{})
                    if err != nil {
                        tasks.Done()
                        continue
                    }
                    for _, o := range objs {
                        if setting.GetBool(conf.IgnoreSystemFiles) && utils.IsSystemFile(o.GetName()) {
                            continue
                        }
                        indexMQ.Publish(mq.Message[ObjWithParent]{
                            Content: ObjWithParent{Obj: o, Parent: it.p},
                        })
                        if o.IsDir() && it.depth != 1 {
                            tasks.Add(1)
                            workCh <- item{p: path.Join(it.p, o.GetName()), info: o, depth: it.depth - 1}
                        }
                    }
                    tasks.Done()
                }
            }()
        }
        tasks.Add(1)
        workCh <- item{p: indexPath, info: fi, depth: maxDepth}
        go func() {
            tasks.Wait()
            close(workCh)
        }()
        wgw.Wait()
    }
    return nil
}

func Del(ctx context.Context, prefix string) error {
	return instance.Del(ctx, prefix)
}

func Clear(ctx context.Context) error {
	return instance.Clear(ctx)
}

func Config(ctx context.Context) searcher.Config {
	return instance.Config()
}

func Update(ctx context.Context, parent string, objs []model.Obj) {
	if instance == nil || !instance.Config().AutoUpdate || !setting.GetBool(conf.AutoUpdateIndex) || Running() {
		return
	}
	if isIgnorePath(parent) {
		return
	}
	// only update when index have built
	progress, err := Progress()
	if err != nil {
		log.Errorf("update search index error while get progress: %+v", err)
		return
	}
	if !progress.IsDone {
		return
	}

	// Use task queue for Meilisearch to avoid race conditions with async indexing
	if msInstance, ok := instance.(interface {
		EnqueueUpdate(parent string, objs []model.Obj)
	}); ok {
		// Enqueue task for async processing (diff calculation happens at consumption time)
		msInstance.EnqueueUpdate(parent, objs)
		return
	}

	nodes, err := instance.Get(ctx, parent)
	if err != nil {
		log.Errorf("update search index error while get nodes: %+v", err)
		return
	}
	now := mapset.NewSet[string]()
	for i := range objs {
		now.Add(objs[i].GetName())
	}
	old := mapset.NewSet[string]()
	for i := range nodes {
		old.Add(nodes[i].Name)
	}
	// delete data that no longer exists
	toDelete := old.Difference(now)
	toAdd := now.Difference(old)
	for i := range nodes {
		if toDelete.Contains(nodes[i].Name) && !op.HasStorage(path.Join(parent, nodes[i].Name)) {
			log.Debugf("delete index: %s", path.Join(parent, nodes[i].Name))
			err = instance.Del(ctx, path.Join(parent, nodes[i].Name))
			if err != nil {
				log.Errorf("update search index error while del old node: %+v", err)
				return
			}
		}
	}
	// collect files and folders to add in batch
	var toAddObjs []ObjWithParent
	for i := range objs {
		if toAdd.Contains(objs[i].GetName()) {
			log.Debugf("add index: %s", path.Join(parent, objs[i].GetName()))
			toAddObjs = append(toAddObjs, ObjWithParent{
				Parent: parent,
				Obj:    objs[i],
			})
		}
	}
	// batch index all files and folders at once
	if len(toAddObjs) > 0 {
		err = BatchIndex(ctx, toAddObjs)
		if err != nil {
			log.Errorf("update search index error while batch index new nodes: %+v", err)
			return
		}
	}
}

func init() {
	op.RegisterObjsUpdateHook(Update)
}