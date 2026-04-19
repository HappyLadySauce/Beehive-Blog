package config

import (
	"fmt"
	"path/filepath"

	"github.com/zeromicro/go-zero/core/conf"
	zgateway "github.com/zeromicro/go-zero/gateway"
)

type upstreamFile struct {
	Upstreams []zgateway.Upstream `json:",optional"`
}

func LoadUpstreams(baseDir string, files []string) ([]zgateway.Upstream, error) {
	if len(files) == 0 {
		return nil, nil
	}

	var merged []zgateway.Upstream
	for _, file := range files {
		path := file
		if !filepath.IsAbs(path) {
			path = filepath.Join(baseDir, file)
		}

		var cfg upstreamFile
		conf.MustLoad(path, &cfg)
		for i := range cfg.Upstreams {
			for j := range cfg.Upstreams[i].ProtoSets {
				protoSet := cfg.Upstreams[i].ProtoSets[j]
				if filepath.IsAbs(protoSet) {
					continue
				}
				cfg.Upstreams[i].ProtoSets[j] = filepath.Clean(filepath.Join(filepath.Dir(path), protoSet))
			}
		}
		merged = append(merged, cfg.Upstreams...)
	}

	if len(merged) == 0 {
		return nil, fmt.Errorf("no upstreams loaded")
	}
	return merged, nil
}
