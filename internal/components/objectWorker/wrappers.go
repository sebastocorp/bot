package objectWorker

import (
	"bot/internal/managers/objectStorage"
	"strings"
)

func (ow *ObjectWorkerT) getBackendObject(object objectStorage.ObjectT) (backend objectStorage.ObjectT, source string) {
	backend = object
	switch ow.config.ObjectWorker.Routing.Type {
	case "bucket":
		{
			if route, ok := ow.config.ObjectWorker.Routing.Routes[object.Bucket]; ok {
				source = route.Backend.Source
				for _, modv := range route.Backend.Modifiers {
					for _, modcongv := range ow.config.ObjectWorker.Modifiers {
						if modcongv.Name == modv {
							backend.Bucket = modcongv.Bucket
							backend.Path = modcongv.AddPrefix + strings.TrimPrefix(object.Path, modcongv.RemovePrefix)
						}
					}
				}
			}
		}
	}

	return backend, source
}

func (ow *ObjectWorkerT) getFrontendObject(object objectStorage.ObjectT) (frontend objectStorage.ObjectT, source string) {
	frontend = object
	switch ow.config.ObjectWorker.Routing.Type {
	case "bucket":
		{
			if route, ok := ow.config.ObjectWorker.Routing.Routes[object.Bucket]; ok {
				source = route.Front.Source
				for _, modv := range route.Front.Modifiers {
					for _, modcongv := range ow.config.ObjectWorker.Modifiers {
						if modcongv.Name == modv {
							frontend.Bucket = modcongv.Bucket
							frontend.Path = modcongv.AddPrefix + strings.TrimPrefix(object.Path, modcongv.RemovePrefix)
						}
					}
				}
			}
		}
	}

	return frontend, source
}
