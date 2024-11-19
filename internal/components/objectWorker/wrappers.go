package objectWorker

import (
	"bot/api/v1alpha3"
	"bot/internal/managers/objectStorage"
	"fmt"
	"strings"
)

func (ow *ObjectWorkerT) getBackendObject(object objectStorage.ObjectT) (backend objectStorage.ObjectT, source string, err error) {
	backend = objectStorage.ObjectT{}
	found := false
	route := v1alpha3.RouteConfigT{}
	switch ow.config.ObjectWorker.Routing.Type {
	case "bucket":
		{
			if route, found = ow.config.ObjectWorker.Routing.Routes[object.Bucket]; found {
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
	case "metadata":
		{
			conditionKey := object.Metadata.Get(ow.config.ObjectWorker.Routing.MetadataKey)
			if route, found = ow.config.ObjectWorker.Routing.Routes[conditionKey]; found {
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
	if !found {
		return backend, source, fmt.Errorf("not route for backend object")
	}

	if backend.Bucket == "" || backend.Path == "" {
		err = fmt.Errorf("empty bucket or path object, check modifiers")
	}

	return backend, source, err
}

func (ow *ObjectWorkerT) getFrontendObject(object objectStorage.ObjectT) (frontend objectStorage.ObjectT, source string, err error) {
	frontend = objectStorage.ObjectT{}
	found := false
	route := v1alpha3.RouteConfigT{}
	switch ow.config.ObjectWorker.Routing.Type {
	case "bucket":
		{
			if route, found = ow.config.ObjectWorker.Routing.Routes[object.Bucket]; found {
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
	case "metadata":
		{
			conditionKey := object.Metadata.Get(ow.config.ObjectWorker.Routing.MetadataKey)
			if route, found = ow.config.ObjectWorker.Routing.Routes[conditionKey]; found {
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
	if !found {
		return frontend, source, fmt.Errorf("not route for backend object")
	}

	if frontend.Bucket == "" || frontend.Path == "" {
		err = fmt.Errorf("empty bucket or path object, check modifiers")
	}

	return frontend, source, err
}
